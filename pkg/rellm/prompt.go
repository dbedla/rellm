package rellm

import (
	"fmt"
	"strings"
)

// Prompt holds the instruction and parameters for a single LLM interaction.
// interaction can consist of multiple message exchange agent <-> llm (usualy tool cals)
// Prompt should be build via PromptBuilder
type Prompt struct {
	msg    string
	params promptParams
}

// promptParams holds inference parameters configured via WithXxx methods.
type promptParams struct {
	Temperature      *float32
	Reasoning        *ReasoningConfig
	MaxOutputTokens  int
	TopP             *float32
	PresencePenalty  *float32
	FrequencyPenalty *float32
	Seed             *int64
	Logprobs         bool
	TopLogprobs      int
}

// PromptBuilder is used to construct a prompt with specific parameters.
type PromptBuilder struct {
	msg    string
	params promptParams
}

// NewPromptBuilder creates a new PromptBuilder with default parameters. first method in chain
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		params: promptParams{},
	}
}

// WithMessage sets the message for the prompt. mandatory field
func (b *PromptBuilder) WithMessage(msg string) *PromptBuilder {
	b.msg = msg
	return b
}

// WithTemperature sets the temperature for the prompt. optional
// todo: add description what is temperature responsible for
func (b *PromptBuilder) WithTemperature(t float32) *PromptBuilder {
	b.params.Temperature = &t
	return b
}

// WithReasoning sets the reasoning configuration for the prompt. optional
// todo: add description what is reasoning responsible for
func (b *PromptBuilder) WithReasoning(effort ReasoningEffort) *PromptBuilder {

	b.params.Reasoning = &ReasoningConfig{Effort: effort}
	return b
}

// WithMaxOutputTokens sets the maximum number of output tokens for the prompt. optional
// Maximum number of output tokens controls the maximum length of the generated text. Higher values make the output longer, while lower values make it shorter. value is set for each request so it means that if during prompt execution agent will call llm each request will get same provided value
// todo: add better description
func (b *PromptBuilder) WithMaxOutputTokens(n int) *PromptBuilder {
	b.params.MaxOutputTokens = n
	return b
}

// WithTopP sets TopP for the prompt, optional,
// todo: add better description what is for this parameter
func (b *PromptBuilder) WithTopP(t float32) *PromptBuilder {
	b.params.TopP = &t
	return b
}

// WithPresencePenalty sets PresencePenalty for the prompt, optional,
// todo: add better description what is for this parameter
func (b *PromptBuilder) WithPresencePenalty(p float32) *PromptBuilder {
	b.params.PresencePenalty = &p
	return b
}

// WithFrequencyPenalty sets FrequencyPenalty for the prompt, optional,
// todo: add better description what is for this parameter
func (b *PromptBuilder) WithFrequencyPenalty(f float32) *PromptBuilder {
	b.params.FrequencyPenalty = &f
	return b
}

// WithSeed sets Seed for the prompt, optional,
// todo: add better description what is for this parameter
func (b *PromptBuilder) WithSeed(seed int64) *PromptBuilder {
	b.params.Seed = &seed
	return b
}

// WithLogprobs sets Logprobs for the prompt, optional,
// todo: add better description what is for this parameter
func (b *PromptBuilder) WithLogprobs(enabled bool) *PromptBuilder {
	b.params.Logprobs = enabled
	return b
}

// WithTopLogprobs sets TopLogprobs for the prompt, optional,
// todo: add better description what is for this parameter
func (b *PromptBuilder) WithTopLogprobs(n int) *PromptBuilder {
	b.params.TopLogprobs = n
	return b
}

// Build returns the constructed prompt. It validates accumulated WithXxx calls;
// last method in chain, message cannot be empty, message with only whitespaces is consider empty
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

func promptMessageToConversation(prompt, role string) (ConversationElement, error) {
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
