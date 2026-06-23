package rellm

import (
	"encoding/json"
	"fmt"
)

type OutputItem struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	// Message fields
	Role    string `json:"role,omitempty"`
	Content []struct {
		Type        string        `json:"type"`
		Text        string        `json:"text"`
		Annotations []interface{} `json:"annotations"`
		Logprobs    []interface{} `json:"logprobs"`
	} `json:"content,omitempty"`
	// Function call fields
	Name      string          `json:"name,omitempty"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
	CallId    string          `json:"call_id,omitempty"`
}

func (a *Agent) Process(conversation []json.RawMessage, proParameterSet FuncLikeProSet) ([]json.RawMessage, string, *ResponsesApiResp, error) {
	for range a.maxToolsIterationWithoutReturnMessage {
		conversationResponse, err := a.endpoint.Post(conversation, proParameterSet, a.toolset)
		if err != nil {
			return nil, "", conversationResponse, err
		}

		if conversationResponse.Error != nil {
			a.logger.Error().Msgf("error in conversation response: %v", conversationResponse.Error.Message)
			return nil, "", conversationResponse, fmt.Errorf("error in conversation response: %v", conversationResponse.Error.Message)
		}

		functionResultAsConversation, msgRespFromLLM, err := a.dispatchFunctionOutput(conversationResponse.Output)
		if err != nil {
			return nil, "", conversationResponse, err
		}
		conversation = append(conversation, functionResultAsConversation...)

		// only msg
		if msgRespFromLLM != "" {
			return conversation, msgRespFromLLM, conversationResponse, nil
		}
	}

	a.logger.Warn().Msg("too many function call iterations without return message")
	return conversation, "Warn too many function call iterations without return message", nil, nil
}

func (a *Agent) dispatchFunctionOutput(output []json.RawMessage) ([]json.RawMessage, string, error) {
	var conversationElements []json.RawMessage
	var msgRespFromLLM string

	for _, o := range output {
		var item OutputItem
		if err := json.Unmarshal(o, &item); err != nil {
			return nil, "", err
		}

		elements, msg, err := a.processOutputItem(item, o)
		if err != nil {
			return nil, "", err
		}
		conversationElements = append(conversationElements, elements...)
		if msg != "" {
			msgRespFromLLM = msg
		}
	}

	return conversationElements, msgRespFromLLM, nil
}

func (a *Agent) processOutputItem(item OutputItem, raw json.RawMessage) ([]json.RawMessage, string, error) {
	switch item.Type {
	case "function_call":
		fResp, err := a.handleFunctionCall(item, raw)
		if err != nil {
			return nil, "", err
		}
		return fResp, "", nil
	case "message":
		return handleMessage(item)
	case "reasoning":
		return handleReasoning(raw)
	default:
		//TODO: handle other output: reasoning
		a.logger.Warn().Msgf("unknown output type: %s", item.Type)
		return nil, "", nil
	}
}

func (a *Agent) handleFunctionCall(fn OutputItem, raw json.RawMessage) ([]json.RawMessage, error) {
	if a.toolset == nil {
		a.logger.Warn().Msgf("tool call (%s) but no tools provider)", fn.Name)
		return nil, fmt.Errorf("tool call (%s) but no tools provider)", fn.Name)
	}

	a.logger.Debug().Msgf("tool call %s with id %s with args: %s", fn.Name, fn.CallId, string(fn.Arguments))
	if funcCallResp, ok := a.toolset.DispatchTools(fn.Name, fn.CallId, fn.Arguments); ok {
		a.logger.Debug().Msgf("tool returned call id: %s, value: %s", funcCallResp.CallId, funcCallResp.Output)
		var conversationElements []json.RawMessage
		conversationElements = append(conversationElements, raw)
		rawFuncCallResp, err := json.Marshal(funcCallResp)
		if err != nil {
			return nil, err
		}
		conversationElements = append(conversationElements, rawFuncCallResp)
		return conversationElements, nil
	}

	return nil, fmt.Errorf("tool call (%s) not found", fn.Name)
}

func handleMessage(msg OutputItem) ([]json.RawMessage, string, error) {
	if len(msg.Content) > 1 {
		return nil, "", fmt.Errorf("too many messages in response")
	}

	var conversationElements []json.RawMessage
	var msgRespFromLLM string
	for _, content := range msg.Content {
		assistantMsg := UserMessage{Role: "assistant", Content: content.Text}
		rawAssistantMsg, err := json.Marshal(assistantMsg)
		if err != nil {
			return nil, "", err
		}
		conversationElements = append(conversationElements, rawAssistantMsg)
		msgRespFromLLM = assistantMsg.Content
	}

	return conversationElements, msgRespFromLLM, nil
}

func handleReasoning(raw json.RawMessage) ([]json.RawMessage, string, error) {
	return []json.RawMessage{raw}, "", nil
}

type UserMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
