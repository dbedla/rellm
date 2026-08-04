package rellm

import (
	"encoding/json"
	"errors"
	"fmt"
)

type ImageURL struct {
	URL string `json:"url"`
}

type MessagePart struct {
	Type        string        `json:"type"`
	Text        string        `json:"text,omitempty"`
	ImageURL    *ImageURL     `json:"image_url,omitempty"`
	Annotations []interface{} `json:"annotations,omitempty"`
	Logprobs    []interface{} `json:"logprobs,omitempty"`
}

type OutputItem struct {
	Id     string `json:"id,omitempty"`
	Type   string `json:"type"`
	Status string `json:"status"`

	// Message fields
	Role    string        `json:"role,omitempty"`
	Content []MessagePart `json:"content,omitempty"`

	// Function call fields
	Name      string          `json:"name,omitempty"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
	CallId    string          `json:"call_id,omitempty"`

	// image_generation_call
	Result string `json:"result,omitempty"`
}

func (a *Agent) process(req *ResponsesApiReq) ([]json.RawMessage, string, *ResponsesApiResp, error) {
	for i := uint64(0); i < a.maxToolsIterationWithoutReturnMessage; i++ {
		conversationResponse, err := a.endpoint.Post(req, a.inspectReq, a.inspectResp)
		if err != nil {
			return nil, "", conversationResponse, err
		}

		if conversationResponse.Error != nil {
			a.logger.Error().Msgf("error in conversation response: %v", conversationResponse.Error.Message)
			return nil, "", conversationResponse, errors.Join(ErrInConversationResponse, fmt.Errorf("err msg: %v", conversationResponse.Error.Message))
		}

		outputAsConversation, msgRespFromLLM, hasFunctionCall, err := a.processOutput(conversationResponse.Output)
		req.Input = append(req.Input, outputAsConversation...)
		if err != nil {
			return req.Input, msgRespFromLLM, conversationResponse, err
		}

		if msgRespFromLLM != "" {
			return req.Input, msgRespFromLLM, conversationResponse, nil
		}

		if !hasFunctionCall {
			a.logger.Error().Msg("model returned output without a function call or text message")
			return req.Input, msgRespFromLLM, conversationResponse,
				errors.Join(ErrModelReturnedUnproductiveOutput, fmt.Errorf("response id: %s", conversationResponse.Id))
		}
	}

	a.logger.Error().Msgf("max tool iterations (%d) reached without a return message", a.maxToolsIterationWithoutReturnMessage)
	return req.Input, "", nil,
		errors.Join(ErrMaxToolIterationsReached,
			fmt.Errorf("exceeded %d iterations", a.maxToolsIterationWithoutReturnMessage))
}

func (a *Agent) processOutput(output []json.RawMessage) ([]json.RawMessage, string, bool, error) {
	var conversationElements []json.RawMessage
	var msgRespFromLLM string
	hasFunctionCall := false

	var outputErr error
	for _, o := range output {
		var item OutputItem
		if err := json.Unmarshal(o, &item); err != nil {
			outputErr = errors.Join(outputErr, fmt.Errorf("error unmarshaling output: %w", err))
			continue
		}

		elements, msg, isFuncCall, err := a.dispatchOutputItem(item, o)
		if err != nil {
			outputErr = errors.Join(outputErr, fmt.Errorf("error processing output: %w", err))
		} else if isFuncCall {
			hasFunctionCall = true
		}

		conversationElements = append(conversationElements, elements...)
		if msg != "" {
			msgRespFromLLM = msg
		}
	}

	return conversationElements, msgRespFromLLM, hasFunctionCall, outputErr
}

func (a *Agent) dispatchOutputItem(item OutputItem, raw json.RawMessage) ([]json.RawMessage, string, bool, error) {
	switch item.Type {
	case "function_call":
		fResp, err := a.handleFunctionCall(item, raw)
		if err != nil {
			return fResp, "", false, err
		}
		return fResp, "", true, nil
	case "message":
		result, msg, err := handleMessage(item, a.endpoint.provider)
		return result, msg, false, err
	case "reasoning":
		elems, msg, err := handleReasoning(raw, a.endpoint.provider)
		return elems, msg, false, err
	case "image_generation_call":
		result, msg, err := a.handleImageGenerationCall(item)
		return result, msg, false, err
	default:
		a.logger.Warn().Msgf("unknown output type: %s", item.Type)
		// todo: return some kind of error
		// todo: add custom output function
		return nil, "", false, nil
	}
}

func (a *Agent) handleImageGenerationCall(image OutputItem) ([]json.RawMessage, string, error) {

	// todo: return image as message or something?
	if a.handleImage == nil {
		return nil, "", ErrNoImageHandler
	}

	resultNote, err := a.handleImage(image)
	if err != nil {
		return nil, "", errors.Join(ErrCustomImageHandlerFailed, err)
	}

	imgResp := ImageGenerationConversationPlaceholder{
		Id:     image.Id,
		Type:   image.Type,
		Result: resultNote,
		Status: image.Status,
	}

	rawResp, err := json.Marshal(imgResp)
	if err != nil {
		return nil, "", errors.Join(ErrImageGenerationResp, err)
	}

	return []json.RawMessage{rawResp}, "[system <for user visibility only>] image generated, handler returned: " + resultNote, nil
}

func (a *Agent) handleFunctionCall(fn OutputItem, raw json.RawMessage) ([]json.RawMessage, error) {
	if a.toolset == nil {
		a.logger.Warn().Msgf("tool call (%s) but no tools provided)", fn.Name)
		return nil, ErrNoToolsetButToolCallRequested
	}

	a.logger.Debug().Msgf("tool call %s with id %s with args: %s", fn.Name, fn.CallId, string(fn.Arguments))
	funcCallResp, ok := a.toolset.DispatchTools(fn.Name, fn.CallId, fn.Arguments)
	if !ok {
		funcCallResp = invalidFunctionCallResp(fn)
		a.logger.Warn().Msgf("tool call (%s) not supported)", fn.Name)
		conversation, err := functionCallConversationElements(raw, funcCallResp)

		return conversation, errors.Join(ErrWhileDispatchToolCall, fmt.Errorf("unknown tool name (%s)", fn.Name), err)
	}

	a.logger.Debug().Msgf("tool returned call id: %s, value: %s", funcCallResp.CallId, funcCallResp.Output)
	return functionCallConversationElements(raw, funcCallResp)
}

func invalidFunctionCallResp(fn OutputItem) FunctionCallResp {
	return FunctionCallResp{
		Type:   "function_call_output",
		CallId: fn.CallId,
		Output: "invalid function call (function not found)" + fn.Name,
	}

}

func functionCallConversationElements(raw json.RawMessage, resp FunctionCallResp) ([]json.RawMessage, error) {
	rawResp, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	return []json.RawMessage{raw, rawResp}, nil
}

func handleMessage(msg OutputItem, provider Provider) ([]json.RawMessage, string, error) {
	allParts, msgRespFromLLM := collectMessageParts(msg.Content)

	switch provider {
	case Provider_OpenRouter:
		return handleOpenRouterMessage(msg, allParts, msgRespFromLLM)
	case Provider_LMStudio:
		fallthrough
	default:
		return handleOtherProviderMessage(msg, allParts, msgRespFromLLM)
	}
}

func collectMessageParts(parts []MessagePart) ([]MessagePart, string) {
	collected := make([]MessagePart, 0, len(parts))
	var msgRespFromLLM string
	for _, part := range parts {
		collected = append(collected, part)
		if part.Text != "" {
			if msgRespFromLLM != "" {
				msgRespFromLLM += " "
			}
			msgRespFromLLM += part.Text
		}
	}

	return collected, msgRespFromLLM
}

func handleOpenRouterMessage(msg OutputItem, allParts []MessagePart, msgRespFromLLM string) ([]json.RawMessage, string, error) {
	assistantMsg := openRouterTextMessage{
		Type:    "message",
		Role:    "assistant",
		Id:      msg.Id,
		Status:  msg.Status,
		Content: messagePartsToText(allParts),
	}
	rawAssistantMsg, err := json.Marshal(assistantMsg)
	if err != nil {
		return nil, "", err
	}

	return []json.RawMessage{rawAssistantMsg}, msgRespFromLLM, nil
}

func handleOtherProviderMessage(msg OutputItem, allParts []MessagePart, msgRespFromLLM string) ([]json.RawMessage, string, error) {
	assistantMsg := AssistantMessage{messageContent{
		Id:   msg.Id,
		Role: "assistant", Content: allParts}}
	rawAssistantMsg, err := json.Marshal(assistantMsg)
	if err != nil {
		return nil, "", err
	}

	return []json.RawMessage{rawAssistantMsg}, msgRespFromLLM, nil
}

func handleReasoning(raw json.RawMessage, provider Provider) ([]json.RawMessage, string, error) {
	elems, err := reasoningConversationElements(raw, provider)
	if err != nil {
		return nil, "", err
	}

	return elems, "", nil
}
