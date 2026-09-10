package rellm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (a *Agent) run(ctx context.Context, msg string, params promptParams) (Report, error) {
	conversation, err := a.appendConversation(ctx, msg)
	if err != nil {
		return Report{}, err
	}

	wire, err := a.provider.ToProviderRepresentation(conversation)
	if err != nil {
		return Report{}, errors.Join(ErrConversationElementConversion, err)
	}

	req := toBaseResponsesAPIReq(params, a.provider.Model(), wire)

	if a.toolset != nil {
		req.Tools = a.toolset.Definitions()
	}

	finalReport, err := a.process(ctx, req)

	if err != nil {
		return finalReport, err
	}

	return finalReport, nil
}

func (a *Agent) process(ctx context.Context, req *ResponsesAPIReq) (Report, error) {
	finalReport := Report{}
	for i := uint64(0); i < a.maxAgentSteps; i++ {
		if err := ctx.Err(); err != nil {
			return finalReport, err
		}

		response, err := a.post(ctx, req)
		if err != nil {
			return finalReport, err
		}
		stepStat := StepStat{ApiUsage: response.Usage}
		finalReport.StepsStats = append(finalReport.StepsStats, stepStat)

		if response.Error != nil {
			return finalReport, errors.Join(
				ErrInConversationResponse,
				fmt.Errorf("err msg: %v", response.Error.Message),
			)
		}

		stepReport, conversation, stepErr := a.processStep(ctx, response)
		finalReport.Image = append(finalReport.Image, stepReport.Image...)
		finalReport.Messages = stepReport.Messages

		if len(conversation) > 0 {
			err := a.appendNewConversationElements(ctx, conversation, req)
			if err != nil {
				return finalReport, errors.Join(err, stepErr)
			}
		}
		if stepErr != nil {
			return finalReport, stepErr
		}

		if finalReport.Messages != nil || len(finalReport.Image) > 0 {
			return finalReport, nil
		}
	}

	return finalReport, errors.Join(
		ErrMaxAgentStepsReached,
		fmt.Errorf("exceeded %d iterations", a.maxAgentSteps),
	)
}

func (a *Agent) processStep(ctx context.Context, response *ResponsesAPIResp) (Report, []ConversationElement, error) {
	stepReport := Report{}

	conversation, err := a.provider.ToConversationElements(response.Output)
	if err != nil {
		return stepReport, nil, errors.Join(ErrConversationElementConversion, err)
	}

	message, conversation, images, dispatchErr := a.dispatchConversation(ctx, conversation)
	stepReport.Image = images
	if len(conversation) == 0 {
		if len(images) > 0 && dispatchErr == nil {
			return stepReport, nil, nil
		}
		return stepReport, nil, errors.Join(ErrNoNewConversationElementAfterDispatch, dispatchErr)
	}
	if dispatchErr != nil {
		return stepReport, conversation, dispatchErr
	}
	if message != "" {
		stepReport.Messages = &message
	}

	return stepReport, conversation, nil
}

func (a *Agent) appendNewConversationElements(
	ctx context.Context,
	conversation []ConversationElement,
	req *ResponsesAPIReq,
) error {
	raw, err := a.provider.ToProviderRepresentation(conversation)
	if err != nil {
		return errors.Join(ErrConversationElementConversion, err)
	}
	req.Input = append(req.Input, raw...)

	return a.conversationStorage.Append(ctx, conversation)
}

func (a *Agent) dispatchConversation(ctx context.Context, conversation []ConversationElement) (string, []ConversationElement, []ImageReport, error) {

	var message strings.Builder
	processed := make([]ConversationElement, 0, len(conversation)+4)
	var images []ImageReport

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
			message.WriteString(TextFromContent(el.Content))
			processed = append(processed, el)

		case *ImageGeneration:
			// Keep the provider response intact for the caller. The policy may
			// mutate or replace the image used by the conversation loop.
			original := *el
			images = append(images, ImageReport{Original: &original})
			imageResp, err := a.executeImageGenerationCallHandler(ctx, el)
			if err != nil {
				outputErr = errors.Join(outputErr, err)
				continue
			}
			// Copy the policy output so the report snapshot does not alias the
			// elements appended to storage.
			images[len(images)-1].PolicyOutput = cloneConversationElements(imageResp)
			processed = append(processed, imageResp...)

		default:
			return "", nil, nil, ErrUnknownConversationElement
		}
	}

	return message.String(), processed, images, outputErr
}

// cloneConversationElements returns independent copies of the elements so
// report snapshots do not alias stored conversation history. Element types
// rellm does not know are returned as-is.
func cloneConversationElements(elements []ConversationElement) []ConversationElement {
	clones := make([]ConversationElement, 0, len(elements))
	for _, el := range elements {
		clones = append(clones, cloneConversationElement(el))
	}
	return clones
}

func cloneConversationElement(el ConversationElement) ConversationElement {
	switch e := el.(type) {
	case *ImageGeneration:
		c := *e
		return &c
	case *FunctionCallResp:
		c := *e
		return &c
	case *FunctionCall:
		c := *e
		c.Args = append(json.RawMessage(nil), e.Args...)
		return &c
	case *Reasoning:
		c := *e
		c.Summary = append([]ReasoningSummaryPart(nil), e.Summary...)
		return &c
	case *UserMessage:
		c := *e
		c.Content = cloneMessageParts(e.Content)
		return &c
	case *AssistantMessage:
		c := *e
		c.Content = cloneMessageParts(e.Content)
		return &c
	case *SystemMessage:
		c := *e
		c.Content = cloneMessageParts(e.Content)
		return &c
	case *UnknownElement:
		c := *e
		c.Raw = append(json.RawMessage(nil), e.Raw...)
		return &c
	default:
		return el
	}
}

func cloneMessageParts(parts []MessagePart) []MessagePart {
	clones := make([]MessagePart, len(parts))
	for i, p := range parts {
		clones[i] = p
		if p.ImageURL != nil {
			u := *p.ImageURL
			clones[i].ImageURL = &u
		}
		clones[i].Annotations = append([]interface{}(nil), p.Annotations...)
		clones[i].Logprobs = append([]interface{}(nil), p.Logprobs...)
	}
	return clones
}

func (a *Agent) executeImageGenerationCallHandler(ctx context.Context, image *ImageGeneration) ([]ConversationElement, error) {
	if a.handleImageGeneration == nil {
		return nil, ErrNoImageHandler
	}

	response, err := a.handleImageGeneration(ctx, image)
	if err != nil {
		return nil, errors.Join(ErrCustomImageHandlerFailed, err)
	}

	return response, nil
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
		funcCallResp := funcResultToFunctionCallResp(fn.CallID, result.Err.Error())
		return &funcCallResp, nil
	}

	funcCallResp := funcResultToFunctionCallResp(fn.CallID, result.Value)
	return &funcCallResp, nil
}

func invalidFunctionCallResp(fn *FunctionCall) FunctionCallResp {
	return funcResultToFunctionCallResp(fn.CallID, "invalid function call (function not found) "+fn.Name)
}

func toBaseResponsesAPIReq(params promptParams, model Model, conversation []json.RawMessage) *ResponsesAPIReq {
	return &ResponsesAPIReq{
		Model:            string(model),
		Input:            conversation,
		Temperature:      params.Temperature,
		Reasoning:        params.Reasoning,
		Text:             params.Text,
		MaxOutputTokens:  params.MaxOutputTokens,
		TopP:             params.TopP,
		PresencePenalty:  params.PresencePenalty,
		FrequencyPenalty: params.FrequencyPenalty,
		TopLogprobs:      params.TopLogprobs,
	}
}

func (a *Agent) appendConversation(ctx context.Context, msg string) ([]ConversationElement, error) {
	conversation, err := a.CurrentConversation(ctx)
	if err != nil {
		return nil, err
	}

	if len(conversation) == 0 && len(a.sysMsg) != 0 {
		systemMessage, err := promptMessageToConversation(a.sysMsg, "system")
		if err != nil {
			return nil, err
		}
		err = a.conversationStorage.Append(ctx, []ConversationElement{systemMessage})
		if err != nil {
			return nil, err
		}
		conversation = append(conversation, systemMessage)
	}

	userMsg, err := promptMessageToConversation(msg, "user")
	if err != nil {
		return nil, errors.Join(ErrUserMsgConversionFailed, err)
	}

	conversation = append(conversation, userMsg)
	if err := a.conversationStorage.Append(ctx, []ConversationElement{userMsg}); err != nil {
		return nil, err
	}

	return conversation, nil
}

func funcResultToFunctionCallResp(callID string, funcResult any) FunctionCallResp {
	b, err := json.Marshal(funcResult)
	if err != nil {
		errorMsg := "unable to execute function; " + err.Error()
		return FunctionCallResp{Type: "function_call_output", CallID: callID, Output: errorMsg}
	}

	return FunctionCallResp{Type: "function_call_output", CallID: callID, Output: string(b)}
}
