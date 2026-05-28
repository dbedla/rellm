package rellm

import (
	"encoding/json"
	"fmt"
	"log"
)

type OutputType struct {
	Type string `json:"type"`
}

func (a *Agent) Process(conversation []json.RawMessage) ([]json.RawMessage, string) {
	for range a.maxToolsIterationWithoutReturnMessage {
		conversationResponse, err := a.endpoint.Post(conversation, a.conversationParameters)
		if err != nil {
			panic(err)
		}

		if conversationResponse.Error != nil {
			a.logger.Error().Msgf("error in conversation response: %v", conversationResponse.Error.Message)
			panic(conversationResponse.Error.Message)
		}

		functionResultAsConversation, msgRespFromLLM := a.dispatchFunctionOutput(conversationResponse.Output)
		conversation = append(conversation, functionResultAsConversation...)

		// only msg
		if len(functionResultAsConversation) == 1 && msgRespFromLLM != "" {
			return conversation, msgRespFromLLM
		}
	}

	a.logger.Warn().Msg("too many function call iterations without return message")
	return conversation, "Warn too many function call iterations without return message"
}

func (a *Agent) dispatchFunctionOutput(output []json.RawMessage) ([]json.RawMessage, string) {
	var conversationElements []json.RawMessage
	var msgRespFromLLM string

	for _, o := range output {
		elements, msg := a.processOutputItem(o)
		conversationElements = append(conversationElements, elements...)
		if msg != "" {
			msgRespFromLLM = msg
		}
	}

	return conversationElements, msgRespFromLLM
}

func (a *Agent) processOutputItem(o json.RawMessage) ([]json.RawMessage, string) {
	var outputType OutputType
	if err := json.Unmarshal(o, &outputType); err != nil {
		log.Println("cannot detect type:", err)
		return nil, ""
	}

	switch outputType.Type {
	case "function_call":
		return a.handleFunctionCall(o), ""
	case "message":
		return handleMessage(o)
	}

	return nil, ""
}

func (a *Agent) handleFunctionCall(o json.RawMessage) []json.RawMessage {
	var fn ConversationFunctionOutput
	if err := json.Unmarshal(o, &fn); err != nil {
		log.Println(err)
		return nil
	}

	if a.toolDispatcher == nil {
		a.logger.Warn().Msgf("tool call (%s) but no tools provider)", fn.Name)
		return nil
	}

	a.logger.Debug().Msgf("tool call %s with id %s with args: %s", fn.Name, fn.CallId, fn.Arguments)
	if funcCallResp, ok := a.toolDispatcher(fn.Name, fn.CallId, fn.Arguments); ok {
		a.logger.Debug().Msgf("tool returned call id: %s, value: %s", funcCallResp.CallId, funcCallResp.Output)
		var conversationElements []json.RawMessage
		conversationElements = append(conversationElements, o)
		conversationElements = append(conversationElements, ToRawJsonMsg(funcCallResp))
		return conversationElements
	} else {
		fmt.Printf("unknown function call: %s\n", fn.Name)
	}

	return nil
}

func handleMessage(o json.RawMessage) ([]json.RawMessage, string) {
	var msg ConversationMessageOutput
	if err := json.Unmarshal(o, &msg); err != nil {
		log.Println(err)
		return nil, ""
	}

	if len(msg.Content) > 1 {
		fmt.Printf("Too many messages in response")
		panic("Too many messages in response")
	}

	var conversationElements []json.RawMessage
	var msgRespFromLLM string
	for _, content := range msg.Content {
		assistantMsg := UserMessage{Role: "assistant", Content: content.Text}
		conversationElements = append(conversationElements, ToRawJsonMsg(assistantMsg))
		msgRespFromLLM = assistantMsg.Content
	}

	return conversationElements, msgRespFromLLM
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

func ToRawJsonMsg(fcr any) json.RawMessage {
	b, err := json.Marshal(fcr)
	if err != nil {
		panic(err)
	}
	return b
}
