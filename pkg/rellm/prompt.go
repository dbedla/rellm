package rellm

import (
	"fmt"
	"strings"
)

// Prompt holds the instruction and inference parameters for one interaction
// with the agent. An interaction may involve multiple provider round-trips
// (e.g. tool calls). Build prompts with PromptBuilder.
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

// PromptBuilder constructs a Prompt, validating its fields on Build.
//
// Usage example:
//
//	prompt, err := rellm.NewPromptBuilder().
//		WithMessage("Write a haiku about tide pools.").
//		WithTemperature(0.7).
//		WithMaxOutputTokens(100).
//		Build()
//	if err != nil {
//		log.Fatal(err)
//	}
//	answer, err := agent.Execute(ctx, prompt)
//
// The WithXxx methods are optional except WithMessage, which is required.
type PromptBuilder struct {
	msg    string
	params promptParams
}

// NewPromptBuilder starts a prompt builder chain with default parameters.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		params: promptParams{},
	}
}

// WithMessage sets the user message. Required.
func (b *PromptBuilder) WithMessage(msg string) *PromptBuilder {
	b.msg = msg
	return b
}

// WithTemperature sets the sampling temperature (usually 0..2). Higher values
// make output more random; lower values more deterministic. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#temperature
func (b *PromptBuilder) WithTemperature(t float32) *PromptBuilder {
	b.params.Temperature = &t
	return b
}

// WithReasoning sets the reasoning effort for models that support it. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#reasoning
func (b *PromptBuilder) WithReasoning(effort ReasoningEffort) *PromptBuilder {

	b.params.Reasoning = &ReasoningConfig{Effort: effort}
	return b
}

// WithMaxOutputTokens caps the output length of every provider call in this
// prompt. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#max-tokens
func (b *PromptBuilder) WithMaxOutputTokens(n int) *PromptBuilder {
	b.params.MaxOutputTokens = n
	return b
}

// WithTopP sets nucleus sampling: only tokens within the top-p probability
// mass are considered. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#top-p
func (b *PromptBuilder) WithTopP(t float32) *PromptBuilder {
	b.params.TopP = &t
	return b
}

// WithPresencePenalty discourages repeating already-used tokens (range -2..2).
// Optional.
// See https://openrouter.ai/docs/api_reference/parameters#presence-penalty
func (b *PromptBuilder) WithPresencePenalty(p float32) *PromptBuilder {
	b.params.PresencePenalty = &p
	return b
}

// WithFrequencyPenalty discourages frequent tokens proportionally to their
// repetition (range -2..2). Optional.
// See https://openrouter.ai/docs/api_reference/parameters#frequency-penalty
func (b *PromptBuilder) WithFrequencyPenalty(f float32) *PromptBuilder {
	b.params.FrequencyPenalty = &f
	return b
}

// WithSeed sets the random seed for reproducible output. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#seed
func (b *PromptBuilder) WithSeed(seed int64) *PromptBuilder {
	b.params.Seed = &seed
	return b
}

// WithLogprobs enables returning log probabilities for output tokens. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#logprobs
func (b *PromptBuilder) WithLogprobs(enabled bool) *PromptBuilder {
	b.params.Logprobs = enabled
	return b
}

// WithTopLogprobs sets how many of the most likely tokens are returned per
// output position when WithLogprobs is enabled. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#top-logprobs
func (b *PromptBuilder) WithTopLogprobs(n int) *PromptBuilder {
	b.params.TopLogprobs = n
	return b
}

// Build validates the accumulated WithXxx calls and returns the Prompt. The
// message is required and must not be empty or whitespace-only.
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
