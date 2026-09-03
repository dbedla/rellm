package rellm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (a *Agent) run(ctx context.Context, msg string, params promptParams) (string, error) {
	conversation, err := a.appendConversation(msg)
	if err != nil {
		return "", err
	}

	wire, err := a.provider.ToProviderRepresentation(conversation)
	if err != nil {
		return "", errors.Join(ErrConversationElementConversion, err)
	}

	req := toBaseResponsesAPIReq(params, a.provider.Model(), wire)

	if a.toolset != nil {
		req.Tools = a.toolset.Definitions()
	}

	msgRespFromLLM, err := a.process(ctx, req)

	if err != nil {
		return "", err
	}

	return msgRespFromLLM, nil
}

func (a *Agent) post(ctx context.Context, req *ResponsesAPIReq) (_ *ResponsesAPIResp, err error) {
	if a.inspectReq != nil {
		a.inspectReq(req)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.provider.URL(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header = a.provider.Header()

	resp, err := a.provider.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, ErrEndpointNilResponse
	}

	if resp.Body == nil {
		return nil, errors.Join(ErrEndpointNilBodyInResponse, fmt.Errorf("response status: %s", resp.Status))
	}

	defer closeWithError(&err, resp.Body)
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(err, ErrUnableToReadResponseBody, fmt.Errorf("response status: %s", resp.Status))
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, newHTTPStatusError(resp, rawBody, httpReq.URL.String())
	}

	return parseResponsesAPIResponse(rawBody, a.inspectResp)
}

func (a *Agent) process(ctx context.Context, req *ResponsesAPIReq) (string, error) {
	for i := uint64(0); i < a.maxAgentSteps; i++ {

		err := ctx.Err()
		if err != nil {
			return "", err
		}

		conversationResponse, err := a.post(ctx, req)
		if err != nil {
			return "", err
		}

		if conversationResponse.Error != nil {
			return "", errors.Join(ErrInConversationResponse, fmt.Errorf("err msg: %v", conversationResponse.Error.Message))
		}

		conversation, err := a.provider.ToConversationElements(conversationResponse.Output)
		if err != nil {
			return "", errors.Join(ErrConversationElementConversion, err)
		}
		msg, conversation, imageHandled, err := a.dispatchConversation(ctx, conversation)
		if len(conversation) == 0 {
			return "", errors.Join(ErrNoNewConversationElementAfterDispatch, err)
		}

		raw, errProviderRep := a.provider.ToProviderRepresentation(conversation)
		if errProviderRep != nil {
			return "", errors.Join(ErrConversationElementConversion, errProviderRep)
		}
		req.Input = append(req.Input, raw...)

		conversationErr := a.conversationStorage.Append(conversation)
		if conversationErr != nil {
			return "", conversationErr
		}

		if err != nil {
			return "", err
		}

		if imageHandled || msg != "" {
			return msg, nil
		}

	}

	return "",
		errors.Join(ErrMaxAgentStepsReached,
			fmt.Errorf("exceeded %d iterations", a.maxAgentSteps))
}

func (a *Agent) dispatchConversation(ctx context.Context, conversation []ConversationElement) (string, []ConversationElement, bool, error) {

	var message strings.Builder
	processed := make([]ConversationElement, 0, len(conversation)+4)
	imageHandled := false

	var outputErr error

	for _, conversationElement := range conversation {
		switch el := conversationElement.(type) {
		case *SystemMessage:
			processed = append(processed, el)
		case *UserMessage:
			processed = append(processed, el)
		case *FunctionCallResp:
			processed = append(processed, el)
		case *Reasoning:
			processed = append(processed, el)
		case *UnknownElement:
			unknownResp, err := a.executeUnknownConversationElementHandler(ctx, el)
			if err != nil {
				// Drop unhandled unknowns instead of persisting them; otherwise
				// they poison later turns when reloaded from storage.
				outputErr = errors.Join(outputErr, err)
				continue
			}
			processed = append(processed, unknownResp...)

		case *FunctionCall:
			fResp, err := a.handleFunctionCall(ctx, el)
			if err != nil {
				outputErr = errors.Join(outputErr, err)
			}
			processed = append(processed, el)
			if fResp != nil {
				processed = append(processed, fResp)
			}

		case *AssistantMessage:
			message.WriteString(messagesFromParts(el.Content))
			processed = append(processed, el)

		case *ImageGeneration:
			err := a.executeImageGenerationCallHandler(ctx, el)
			if err != nil {
				outputErr = errors.Join(outputErr, err)
			}
			imageHandled = true
			processed = append(processed, el)

		default:
			return "", nil, false, ErrUnknownConversationElement
		}
	}

	return message.String(), processed, imageHandled, outputErr
}

func messagesFromParts(parts []MessagePart) string {
	var msg strings.Builder
	for _, part := range parts {
		msg.WriteString(part.Text)
	}
	return msg.String()
}

func (a *Agent) executeImageGenerationCallHandler(ctx context.Context, image *ImageGeneration) error {
	if a.handleImageGeneration == nil {
		return ErrNoImageHandler
	}

	resultNote, err := a.handleImageGeneration(ctx, image)
	if err != nil {
		return errors.Join(ErrCustomImageHandlerFailed, err)
	}

	image.Result = resultNote

	return nil
}

func (a *Agent) executeUnknownConversationElementHandler(ctx context.Context, el *UnknownElement) ([]ConversationElement, error) {
	if a.handleUnknownConversationElement == nil {
		return nil, ErrNoUnknownConversationElementHandler
	}

	response, err := a.handleUnknownConversationElement(ctx, el)
	if err != nil {
		return nil, errors.Join(ErrCustomConversationElementHandlerFailed, err)
	}

	return response, nil
}

func (a *Agent) handleFunctionCall(ctx context.Context, fn *FunctionCall) (*FunctionCallResp, error) {
	if a.toolset == nil {
		return nil, ErrNoToolsetButToolCallRequested
	}

	result, dispatchErr := a.toolset.Dispatch(ctx, fn.Name, fn.Args)
	if dispatchErr != nil {
		funcCallResp := invalidFunctionCallResp(fn)
		return &funcCallResp, errors.Join(ErrWhileDispatchToolCall, dispatchErr)
	}

	if result.Err != nil {
		funcCallResp := FuncResultToFunctionCallResp(fn.CallID, result.Err.Error())
		return &funcCallResp, nil
	}

	funcCallResp := FuncResultToFunctionCallResp(fn.CallID, result.Value)
	return &funcCallResp, nil
}

func invalidFunctionCallResp(fn *FunctionCall) FunctionCallResp {
	return FuncResultToFunctionCallResp(fn.CallID, "invalid function call (function not found) "+fn.Name)
}

func toBaseResponsesAPIReq(params promptParams, model Model, conversation []json.RawMessage) *ResponsesAPIReq {
	return &ResponsesAPIReq{
		Model:            string(model),
		Input:            conversation,
		Temperature:      params.Temperature,
		Reasoning:        params.Reasoning,
		MaxOutputTokens:  params.MaxOutputTokens,
		TopP:             params.TopP,
		PresencePenalty:  params.PresencePenalty,
		FrequencyPenalty: params.FrequencyPenalty,
		Seed:             params.Seed,
		Logprobs:         params.Logprobs,
		TopLogprobs:      params.TopLogprobs,
	}
}

func (a *Agent) appendConversation(msg string) ([]ConversationElement, error) {
	conversation, err := a.CurrentConversation()
	if err != nil {
		return nil, err
	}

	if len(conversation) == 0 && len(a.sysMsg) != 0 {
		systemMessage, err := PromptMessageToConversation(a.sysMsg, "system")
		if err != nil {
			return nil, err
		}
		err = a.conversationStorage.Append([]ConversationElement{systemMessage})
		if err != nil {
			return nil, err
		}
		conversation = append(conversation, systemMessage)
	}

	userMsg, err := PromptMessageToConversation(msg, "user")
	if err != nil {
		return nil, errors.Join(ErrUserMsgConversionFailed, err)
	}

	conversation = append(conversation, userMsg)
	if err := a.conversationStorage.Append([]ConversationElement{userMsg}); err != nil {
		return nil, err
	}

	return conversation, nil
}
