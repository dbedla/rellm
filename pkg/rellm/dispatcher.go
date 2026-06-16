package rellm

import (
	"encoding/json"
	"fmt"
	"log"
)

type OutputType struct {
	Type string `json:"type"`
}

func (a *Agent) Process(conversation []json.RawMessage, proParameterSet FuncLikeProSet) ([]json.RawMessage, string, *ResponsesApiResp, error) {
	for range a.maxToolsIterationWithoutReturnMessage {
		conversationResponse, err := a.endpoint.Post(conversation, proParameterSet, a.toolset)
		if err != nil {
			return nil, "", nil, err
		}

		if conversationResponse.Error != nil {
			a.logger.Error().Msgf("error in conversation response: %v", conversationResponse.Error.Message)
			return nil, "", nil, fmt.Errorf("error in conversation response: %v", conversationResponse.Error.Message)
		}

		functionResultAsConversation, msgRespFromLLM, err := a.dispatchFunctionOutput(conversationResponse.Output)
		if err != nil {
			return nil, "", nil, err
		}
		conversation = append(conversation, functionResultAsConversation...)

		// only msg
		if len(functionResultAsConversation) == 1 && msgRespFromLLM != "" {
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
		elements, msg, err := a.processOutputItem(o)
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

func (a *Agent) processOutputItem(o json.RawMessage) ([]json.RawMessage, string, error) {
	var outputType OutputType
	if err := json.Unmarshal(o, &outputType); err != nil {
		return nil, "", err
	}

	switch outputType.Type {
	case "function_call":
		fResp, err := a.handleFunctionCall(o)
		if err != nil {
			return nil, "", err
		}
		return fResp, "", nil
	case "message":
		return handleMessage(o)
	default:
		//TODO: handle other output: reasoning
		a.logger.Warn().Msgf("unknown output type: %s", outputType.Type)
		return nil, "", nil
	}
}

func (a *Agent) handleFunctionCall(o json.RawMessage) ([]json.RawMessage, error) {
	var fn ConversationFunctionOutput
	if err := json.Unmarshal(o, &fn); err != nil {
		log.Println(err)
		return nil, err
	}

	if a.toolset == nil {
		a.logger.Warn().Msgf("tool call (%s) but no tools provider)", fn.Name)
		return nil, fmt.Errorf("tool call (%s) but no tools provider)", fn.Name)
	}

	a.logger.Debug().Msgf("tool call %s with id %s with args: %s", fn.Name, fn.CallId, fn.Arguments)
	if funcCallResp, ok := a.toolset.DispatchTools(fn.Name, fn.CallId, fn.Arguments); ok {
		a.logger.Debug().Msgf("tool returned call id: %s, value: %s", funcCallResp.CallId, funcCallResp.Output)
		var conversationElements []json.RawMessage
		conversationElements = append(conversationElements, o)
		rawFuncCallResp, err := json.Marshal(funcCallResp)
		if err != nil {
			return nil, err
		}
		conversationElements = append(conversationElements, rawFuncCallResp)
		return conversationElements, nil
	} else {
		fmt.Printf("unknown function call: %s\n", fn.Name)
	}

	return nil, nil
}

func handleMessage(o json.RawMessage) ([]json.RawMessage, string, error) {
	var msg ConversationMessageOutput
	if err := json.Unmarshal(o, &msg); err != nil {
		return nil, "", err
	}

	if len(msg.Content) > 1 {
		fmt.Printf("Too many messages in response")
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

type ConversationFunctionOutput struct {
	Id        string `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CallId    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ConversationMessageOutput struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Status  string `json:"status"`
	Content []struct {
		Type        string        `json:"type"`
		Text        string        `json:"text"`
		Annotations []interface{} `json:"annotations"`
		Logprobs    []interface{} `json:"logprobs"`
	} `json:"content"`
	Phase string `json:"phase"`
}

type UserMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
