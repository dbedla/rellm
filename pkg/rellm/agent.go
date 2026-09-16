package rellm

import (
	"context"
	"encoding/json"
	"net/http"
)

// ReasoningEffort controls how much effort the model spends on reasoning
// before answering.
type ReasoningEffort string

const (
	// ReasoningEffortNone disables reasoning.
	ReasoningEffortNone ReasoningEffort = "none"
	// ReasoningEffortLow requests minimal reasoning.
	ReasoningEffortLow ReasoningEffort = "low"
	// ReasoningEffortHigh requests above-average reasoning.
	ReasoningEffortHigh ReasoningEffort = "high"
	// ReasoningEffortMedium requests moderate reasoning.
	ReasoningEffortMedium ReasoningEffort = "medium"
	// ReasoningEffortXHigh requests maximum reasoning.
	ReasoningEffortXHigh ReasoningEffort = "xhigh"
)

// Model identifies a model by its provider-specific name. Names are not
// interchangeable between providers:
//
//	LM Studio:  "google/gemma-4-26b-a4b"    (https://lmstudio.ai/models)
//	OpenRouter: "google/gemma-4-26b-a4b-it" (https://openrouter.ai/models)
type Model string

// ToolCallResult holds the value or error produced by a tool call.
// Err is reported back to the model and does not break the agentic loop.
type ToolCallResult struct {
	Value any
	Err   error
}

// ParallelToolCallsMode controls whether the model may return multiple tool
// calls in a single response.
type ParallelToolCallsMode string

const (
	// ParallelToolCallsDefaultForProvider omits the field from the request;
	// the provider's default behavior applies. It is the zero value.
	ParallelToolCallsDefaultForProvider ParallelToolCallsMode = ""
	// ParallelToolCallsEnable allows the model to call tools in parallel.
	// Determinism-pinning: most providers already default to this.
	ParallelToolCallsEnable ParallelToolCallsMode = "enable"
	// ParallelToolCallsDisable makes the model call one tool per response, for
	// toolsets with ordering-dependent side effects.
	ParallelToolCallsDisable ParallelToolCallsMode = "disable"
)

// Toolset groups the tools an agent can call.
// Tool names placed in definition must match the provider's regex: OpenAI enforces
// ^[a-zA-Z0-9_-]+$, Meta enforces ^[a-zA-Z0-9_.-]+$.
// Prefer generating an implementation with an LLM; see
// pkg/toolsets/calculatortoolset.go for an example.
// See https://openrouter.ai/docs/api_reference/responses/tool-calling
type Toolset interface {
	// Definitions returns the tool definitions advertised to the model: name,
	// arguments, and a description of what the tool does and when to use it.
	Definitions() []ToolDefinition

	// Dispatch executes a tool call. A returned error breaks the agentic loop;
	// put the error in ToolCallResult.Err to feed it back to the model instead.
	Dispatch(ctx context.Context, name string, arguments json.RawMessage) (ToolCallResult, error)
}

// Conversation stores and loads conversation history.
type Conversation interface {
	// Load returns the stored conversation elements. A non-nil error breaks
	// the agentic loop; use ctx to cancel the read.
	Load(context.Context) ([]ConversationElement, error)

	// Append adds elements to the history and may be called multiple times
	// during one prompt. A non-nil error breaks the agentic loop.
	Append(context.Context, []ConversationElement) error
}

// HTTPClient performs HTTP requests. Provide a custom implementation to add
// rate limiting, redirect handling, or other transport policy.
type HTTPClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// Agent manages a conversation and runs the agentic loop: it assembles
// requests, dispatches function calls, and returns a final answer. Create it
// with AgentBuilder.
//
// Available methods: Ask, Execute, Conversation, Name, SystemMessage.
type Agent struct {
	conversation      Conversation
	provider          Provider
	toolset           Toolset
	textFormat        *TextFormat
	parallelToolCalls ParallelToolCallsMode
	agentName         string
	sysMsg            string
	maxAgentSteps     uint64

	handleUnknownConversationElement HandleUnknownConversationElement
	handleImageGeneration            HandleImageGeneration
	inspectReq                       InspectEachRequest
	inspectResp                      InspectEachResponse
}

// HandleImageGeneration is called for each generated image. The returned slice
// replaces the image's slot in the conversation:
//
//   - nil or empty: drop the image from the conversation
//   - []ConversationElement{image}: keep it unchanged
//   - any other slice: replace it with those elements
//
// The original image and the policy output are returned in Report.Images.
type HandleImageGeneration func(ctx context.Context, image *ImageGeneration) ([]ConversationElement, error)

// InspectEachRequest inspects each request before it is sent to the provider.
// Last chance to modify or log it.
type InspectEachRequest func(*ResponsesAPIReq)

// InspectEachResponse inspects each provider response before it is processed.
// Last chance to modify or log it.
type InspectEachResponse func(resp *ResponsesAPIResp)

// HandleUnknownConversationElement is called for each provider output item rellm
// cannot classify. The returned slice replaces the unknown element's slot in
// the conversation:
//
//   - nil or empty: drop the element
//   - []ConversationElement{el}: keep it unchanged
//   - any other slice: replace it with those elements
type HandleUnknownConversationElement func(ctx context.Context, el *UnknownElement) ([]ConversationElement, error)

// Report is the result of one agent run (Ask or Execute). It carries the
// final assistant message, every image the run produced, and one usage stat
// per provider call.
//
// A Report is generated exclusively for a single Ask or Execute call: it
// never carries information from previous calls. What does persist across
// calls is the conversation history (see Conversation and Agent.Conversation);
// a Report only reflects what its own run received
// from the provider.
//
// When a run ends in error, the Report returned alongside the error holds
// everything gathered so far: stats from the provider calls that succeeded
// (even the one that failed, if it carried usage), and images from earlier
// steps. Message is empty unless the failing step also produced text.
type Report struct {
	// Message is the text of the assistant message from the last
	// successful step. It is empty when the run produced no text (for
	// example an image-only or tool-only response) or ended in error.
	Message string
	// Images holds one entry per generated image, accumulated across all
	// steps of the run.
	Images []ImageReport
	// StepsStats holds one entry per provider call, in call order.
	StepsStats []StepStat
}

// ImageReport describes one generated image: Original is the provider payload
// as received, and PolicyOutput is an independent copy of the elements the
// image policy substituted into the conversation.
type ImageReport struct {
	Original     *ImageGeneration
	PolicyOutput []ConversationElement
}

// StepStat reports the API usage of a single provider call within one
// agent run. Providers that omit usage are recorded with zero values.
type StepStat struct {
	APIUsage ResponsesAPIUsage
}

// Execute runs a prompt built with PromptBuilder, whose parameters control
// this interaction. A nil prompt or empty message returns ErrEmptyPrompt.
func (a *Agent) Execute(ctx context.Context, p *Prompt) (Report, error) {
	if p == nil || p.msg == "" {
		return Report{}, ErrEmptyPrompt
	}
	return a.run(ctx, p.msg, p.params)
}

// Ask is a minimal entry point for a plain-text question. An empty question
// returns ErrEmptyPrompt.
func (a *Agent) Ask(ctx context.Context, question string) (Report, error) {
	prompt, err := NewPromptBuilder().WithMessage(question).Build()
	if err != nil {
		return Report{}, err
	}

	return a.Execute(ctx, prompt)
}

// Name returns the agent's name, an identifier set via WithAgentName.
// It is empty when WithAgentName was not called.
func (a *Agent) Name() string {
	return a.agentName
}

// Conversation returns the agent's conversation. Use it to read the history
// (Load) and append elements (Append); the implementation cannot be swapped
// for another one after Build.
func (a *Agent) Conversation() Conversation {
	return a.conversation
}

// SystemMessage returns the agent's system message, the standing
// instructions sent with every request this agent makes. It is empty when
// WithSystemMessage was not called.
func (a *Agent) SystemMessage() string {
	return a.sysMsg
}
