package rellm

import (
	"encoding/json"
	"fmt"
)

// Prompt holds the instruction and parameters for a single LLM interaction.
// It is unexported — users can only obtain it via Agent.Prompt(), so the agent reference stays internal.
type Prompt struct {
	msg    string
	params promptParams
}

// WithTemperature sets sampling temperature for all calls in this prompt's loop.
func (p *Prompt) WithTemperature(t float32) *Prompt {
	p.params.Temperature = t
	return p
}

// WithReasoning sets reasoning effort level ("low", "medium", "high") for the entire loop.
func (p *Prompt) WithReasoning(effort string) *Prompt {
	if effort != "" {
		p.params.Reasoning = &ReasoningConfig{Effort: effort}
	}
	return p
}

// WithMaxOutputTokens sets max tokens for every call in this prompt's loop.
func (p *Prompt) WithMaxOutputTokens(n int) *Prompt {
	p.params.MaxOutputTokens = n
	return p
}

// WithTopP sets nucleus sampling parameter.
func (p *Prompt) WithTopP(t float32) *Prompt {
	p.params.TopP = t
	return p
}

// WithPresencePenalty sets presence penalty for every call in this prompt's loop.
func (p *Prompt) WithPresencePenalty(penalty float32) *Prompt {
	p.params.PresencePenalty = penalty
	return p
}

// WithFrequencyPenalty sets frequency penalty for every call in this prompt's loop.
func (p *Prompt) WithFrequencyPenalty(f float32) *Prompt {
	p.params.FrequencyPenalty = f
	return p
}

// WithSeed sets deterministic seed applied to every call in the loop.
func (p *Prompt) WithSeed(seed int64) *Prompt {
	s := seed
	p.params.Seed = &s
	return p
}

// WithLogprobs enables log probabilities output on every call in this prompt's loop.
func (p *Prompt) WithLogprobs(enabled bool) *Prompt {
	p.params.Logprobs = enabled
	return p
}

// WithTopLogprobs sets number of top log probabilities to return per call.
func (p *Prompt) WithTopLogprobs(n int) *Prompt {
	p.params.TopLogprobs = n
	return p
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
		params: defaultParams,
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

func (b *PromptBuilder) Build() *Prompt {
	return &Prompt{
		msg:    b.msg,
		params: b.params,
	}
}

// defaultParams is used as the baseline for all new builders/prompts.
var defaultParams = promptParams{}

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
