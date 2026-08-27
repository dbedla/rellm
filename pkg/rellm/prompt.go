package rellm

import (
	"fmt"
	"strings"
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

func (b *PromptBuilder) WithReasoning(effort ReasoningEffort) *PromptBuilder {

	b.params.Reasoning = &ReasoningConfig{Effort: effort}
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
	b.params.Seed = &seed
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

// Build returns the constructed prompt. It validates accumulated WithXxx calls;
// all validation errors are joined into a single error so callers see every problem at once.
func (b *PromptBuilder) Build() (*Prompt, error) {
	trimmedMsg := strings.TrimSpace(b.msg)
	if trimmedMsg == "" {
		return nil, ErrEmptyPrompt
	}

	if b.params.Reasoning != nil {
		if b.params.Reasoning.Effort == "" {
			return nil, ErrEmptyReasoningEffort
		}
	}
	return &Prompt{
		msg:    trimmedMsg,
		params: b.params,
	}, nil
}

func PromptMessageToConversation(prompt, role string) (ConversationElement, error) {
	content := []MessagePart{{Type: "input_text", Text: prompt}}
	switch role {
	case "user":
		return &UserMessage{MessageContent{Role: role, Content: content}}, nil
	case "system":
		return &SystemMessage{MessageContent{Role: role, Content: content}}, nil
	case "assistant":
		return &AssistantMessage{MessageContent{Role: role, Content: content}}, nil
	default:
		return nil, fmt.Errorf("promptMessageToConversation: unknown role %q", role)
	}
}
