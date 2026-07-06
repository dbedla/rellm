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

	//image_generation_call
	Result string `json:"result,omitempty"`
}

func (a *Agent) Process(conversation []json.RawMessage, inspectReq InspectEachRequest, inspectResp InspectEachResponse) ([]json.RawMessage, string, *ResponsesApiResp, error) {
	for range a.maxToolsIterationWithoutReturnMessage {
		conversationResponse, err := a.endpoint.Post(conversation, inspectReq, inspectResp, a.toolset)
		if err != nil {
			return nil, "", conversationResponse, err
		}

		if conversationResponse.Error != nil {
			a.logger.Error().Msgf("error in conversation response: %v", conversationResponse.Error.Message)
			return nil, "", conversationResponse, errors.Join(ErrInConversationResponse, fmt.Errorf("err msg: %v", conversationResponse.Error.Message))
		}

		outputAsConversation, msgRespFromLLM, err := a.processOutput(conversationResponse.Output)
		conversation = append(conversation, outputAsConversation...)
		if err != nil {
			return conversation, msgRespFromLLM, conversationResponse, err
		}

		// only msg
		if msgRespFromLLM != "" {
			return conversation, msgRespFromLLM, conversationResponse, err
		}
	}

	//todo: too many function call should be error
	a.logger.Warn().Msg("too many function call iterations without return message")
	return conversation, "Warn too many function call iterations without return message", nil, nil
}

func (a *Agent) processOutput(output []json.RawMessage) ([]json.RawMessage, string, error) {
	var conversationElements []json.RawMessage
	var msgRespFromLLM string

	var outputErr error
	for _, o := range output {
		var item OutputItem
		if err := json.Unmarshal(o, &item); err != nil {
			outputErr = errors.Join(outputErr, fmt.Errorf("error unmarshaling output: %w", err))
			continue
			//return nil, "", err
		}

		elements, msg, err := a.dispatchOutputItem(item, o)
		if err != nil {
			outputErr = errors.Join(outputErr, fmt.Errorf("error processing output: %w", err))
		}

		conversationElements = append(conversationElements, elements...)
		if msg != "" {
			msgRespFromLLM = msg
		}
	}

	return conversationElements, msgRespFromLLM, outputErr
}

func (a *Agent) dispatchOutputItem(item OutputItem, raw json.RawMessage) ([]json.RawMessage, string, error) {
	switch item.Type {
	case "function_call":
		fResp, err := a.handleFunctionCall(item, raw)
		if err != nil {
			return fResp, "", err
		}
		return fResp, "", nil
	case "message":
		return handleMessage(item)
	case "reasoning":
		return handleReasoning(raw)
	case "image_generation_call":
		return a.handleImageGenerationCall(item, raw)
	default:
		a.logger.Warn().Msgf("unknown output type: %s", item.Type)
		//todo: return some kind of error
		//todo: add custom output function
		return nil, "", nil
	}
}

func (a *Agent) handleImageGenerationCall(image OutputItem, raw json.RawMessage) ([]json.RawMessage, string, error) {

	imgResp := ImageGenerationResp{
		Id:   image.Id,
		Type: image.Type,
		//Result: image.Result,
		Status: image.Status,
	}
	//todo: custom handler

	rawResp, err := json.Marshal(imgResp)
	if err != nil {
		return nil, "", errors.Join(ErrImageGenerationResp, err)
	}

	return []json.RawMessage{rawResp}, "image generated - cheat message", nil
}

func (a *Agent) handleFunctionCall(fn OutputItem, raw json.RawMessage) ([]json.RawMessage, error) {
	if a.toolset == nil {
		a.logger.Warn().Msgf("tool call (%s) but no tools provided)", fn.Name)
		return nil, ErrNoToolsetBuToolCall
	}

	a.logger.Debug().Msgf("tool call %s with id %s with args: %s", fn.Name, fn.CallId, string(fn.Arguments))
	funcCallResp, ok := a.toolset.DispatchTools(fn.Name, fn.CallId, fn.Arguments)
	if !ok {
		funcCallResp = invalidFunctionCallResp(fn)
		a.logger.Warn().Msgf("tool call (%s) not supported)", fn.Name)
		conversation, err := functionCallConversationElements(raw, funcCallResp)

		return conversation, errors.Join(ErrUnknownToolCallsErrorsWillBePassedToModelInNextReq, fmt.Errorf("unknown tool name (%s))", fn.Name), err)
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

func handleMessage(msg OutputItem) ([]json.RawMessage, string, error) {
	var conversationElements []json.RawMessage
	var msgRespFromLLM string
	var allParts []MessagePart

	for _, part := range msg.Content {
		allParts = append(allParts, part)
		if part.Text != "" {
			if msgRespFromLLM != "" {
				msgRespFromLLM += " "
			}
			msgRespFromLLM += part.Text
		}
	}

	assistantMsg := UserMessage{Role: "assistant", Content: allParts}
	rawAssistantMsg, err := json.Marshal(assistantMsg)
	if err != nil {
		return nil, "", err
	}
	conversationElements = append(conversationElements, rawAssistantMsg)

	return conversationElements, msgRespFromLLM, nil
}

func handleReasoning(raw json.RawMessage) ([]json.RawMessage, string, error) {
	return []json.RawMessage{raw}, "", nil
}

type UserMessage struct {
	Role    string        `json:"role"`
	Content []MessagePart `json:"content"`
}
