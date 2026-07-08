package rellm

import (
	"encoding/json"
	"fmt"
)

// promptParams holds inference parameters configured via WithXxx methods.
// It is the only data passed from prompt.Execute() to agent.run().
type promptParams struct {
	Temperature      float32
	Reasoning        *ReasoningConfig
	MaxOutputTokens  int
	TopP             float32
	PresencePenalty  float32
	FrequencyPenalty float32
	Seed             *int64
	Logprobs         bool
	TopLogprobs      int
}

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
