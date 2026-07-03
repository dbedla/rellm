package rellm

import (
	"encoding/json"
	"fmt"
)

func PromptMessageToConversation(prompt, role string) (json.RawMessage, error) {
	input := UserMessage{
		Role: role,
		Content: []MessagePart{
			{Type: "input_text", Text: prompt},
		},
	}

	jsonInput, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("promptMessageToConversation: %w", err)
	}

	return jsonInput, nil
}

func BuildStartOfConversation(systemPrompt, userPrompt string) ([]json.RawMessage, error) {
	conversation := []json.RawMessage{}

	if systemPrompt != "" {
		sp, err := PromptMessageToConversation(systemPrompt, "system")
		if err != nil {
			return nil, err
		}
		conversation = append(conversation, sp)
	}

	if userPrompt != "" {
		up, err := PromptMessageToConversation(userPrompt, "user")
		if err != nil {
			return nil, err
		}
		conversation = append(conversation, up)
	}

	return conversation, nil
}
