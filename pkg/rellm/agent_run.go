package rellm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (a *Agent) run(ctx context.Context, msg string, params promptParams) (string, error) {
	conversation, err := a.CurrentConversation()
	if err != nil {
		return "", err
	}

	userMsg, err := PromptMessageToConversation(msg, "user")
	if err != nil {
		return "", errors.Join(ErrUserMsgConversionFailed, err)
	}

	conversation = append(conversation, userMsg)
	if err := a.conversationStorage.Append([]ConversationElement{userMsg}); err != nil {
		return "", err
	}

	wire, err := a.provider.ToProviderRepresentation(conversation)
	if err != nil {
		return "", errors.Join(ErrConversationElementConversion, err)
	}

	req := toBaseResponsesAPIReq(params, a.provider.Model(), wire)

	if a.toolset != nil {
		req.Tools = a.toolset.BuildTools()
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

	apiUrl, err := url.Parse(a.provider.URL())
	if err != nil {
		return nil, err
	}

	httpReq := &http.Request{
		Method: "POST",
		Header: a.provider.Header(),
		URL:    apiUrl,
		Body:   io.NopCloser(bytes.NewReader(body)),
	}

	httpReq = httpReq.WithContext(ctx)

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
		return nil, newHTTPStatusError(resp, rawBody, apiUrl.String())
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
		msg, newConversationElements, imageHandled, err := a.dispatchConversation(ctx, conversation)
		conversation = append(conversation, newConversationElements...)

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
	newConversationElements := []ConversationElement{}
	imageHandled := false

	var outputErr error

	for _, conversationElement := range conversation {
		switch el := conversationElement.(type) {
		case *SystemMessage:
			continue
		case *UserMessage:
			continue
		case *FunctionCallResp:
			continue
		case *Reasoning:
			continue
		case *UnknownElement:
			unknownResp, err := a.executeUnknownConversationCallHandler(ctx, el)
			if err != nil {
				return "", nil, false, errors.Join(ErrCustomConversationElementHandlerFailed, err)
			}
			if unknownResp != nil {
				newConversationElements = append(newConversationElements, unknownResp)
			}
			continue

		case *FunctionCall:
			fResp, err := a.handleFunctionCall(ctx, el)
			if err != nil {
				outputErr = errors.Join(outputErr, err)
			}
			if fResp != nil {
				newConversationElements = append(newConversationElements, fResp)
			}

		case *AssistantMessage:
			message.WriteString(messagesFromParts(el.Content))
			continue

		case *ImageGeneration:
			err := a.executeImageGenerationCallHandler(ctx, el)
			if err != nil {
				outputErr = errors.Join(outputErr, err)
			}
			imageHandled = true
			continue

		default:
			return "", nil, false, ErrUnknownConversationElement
		}
	}

	return message.String(), newConversationElements, imageHandled, outputErr
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

func (a *Agent) executeUnknownConversationCallHandler(ctx context.Context, el *UnknownElement) (*UnknownElement, error) {
	if a.handleUnknownConversationElement == nil {
		return nil, ErrNoUnknownConversationElementHandler
	}

	fixed, err := a.handleUnknownConversationElement(ctx, el)
	if err != nil {
		return nil, errors.Join(ErrCustomConversationElementHandlerFailed, err)
	}

	return fixed, nil
}

func (a *Agent) handleFunctionCall(ctx context.Context, fn *FunctionCall) (*FunctionCallResp, error) {
	if a.toolset == nil {
		return nil, ErrNoToolsetButToolCallRequested
	}

	funcCallResp, ok := a.toolset.DispatchTools(ctx, fn.Name, fn.CallID, fn.Args)
	if !ok {
		funcCallResp = invalidFunctionCallResp(fn)
		return &funcCallResp, errors.Join(ErrWhileDispatchToolCall, fmt.Errorf("unknown tool name (%s)", fn.Name))
	}

	return &funcCallResp, nil
}

func invalidFunctionCallResp(fn *FunctionCall) FunctionCallResp {
	return FunctionCallResp{
		Type:   "function_call_output",
		CallID: fn.CallID,
		Output: "invalid function call (function not found) " + fn.Name,
	}

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
