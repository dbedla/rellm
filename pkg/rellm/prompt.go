package rellm

import (
	"encoding/json"
	"fmt"
)

// Prompt holds the instruction and parameters for a single LLM interaction.
type Prompt struct {
	msg    string
	params promptParams
}

// promptParams holds inference parameters configured via WithXxx methods.
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

// PromptBuilder is used to construct a prompt with specific parameters.
type PromptBuilder struct {
	msg    string
	params promptParams
}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		params: promptParams{},
	}
}

func (b *PromptBuilder) WithMessage(msg string) *PromptBuilder {
	b.msg = msg
	return b
}

func (b *PromptBuilder) WithTemperature(t float32) *PromptBuilder {
	b.params.Temperature = t
	return b
}

func (b *PromptBuilder) WithReasoning(effort string) *PromptBuilder {
	if effort != "" {
		b.params.Reasoning = &ReasoningConfig{Effort: effort}
	}
	return b
}

func (b *PromptBuilder) WithMaxOutputTokens(n int) *PromptBuilder {
	b.params.MaxOutputTokens = n
	return b
}

func (b *PromptBuilder) WithTopP(t float32) *PromptBuilder {
	b.params.TopP = t
	return b
}

func (b *PromptBuilder) WithPresencePenalty(p float32) *PromptBuilder {
	b.params.PresencePenalty = p
	return b
}

func (b *PromptBuilder) WithFrequencyPenalty(f float32) *PromptBuilder {
	b.params.FrequencyPenalty = f
	return b
}

func (b *PromptBuilder) WithSeed(seed int64) *PromptBuilder {
	s := seed
	b.params.Seed = &s
	return b
}

func (b *PromptBuilder) WithLogprobs(enabled bool) *PromptBuilder {
	b.params.Logprobs = enabled
	return b
}

func (b *PromptBuilder) WithTopLogprobs(n int) *PromptBuilder {
	b.params.TopLogprobs = n
	return b
}

// Build returns the constructed prompt. It validates that a message was set;
// callers receive ErrEmptyPrompt if they try to execute an unconfigured builder.
func (b *PromptBuilder) Build() (*Prompt, error) {
	if b.msg == "" {
		return nil, ErrEmptyPrompt
	}
	return &Prompt{
		msg:    b.msg,
		params: b.params,
	}, nil
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
