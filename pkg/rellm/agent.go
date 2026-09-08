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

// Toolset groups the tools an agent can call.
// Tool names must match the provider's regex: OpenAI enforces
// ^[a-zA-Z0-9_-]+$, Meta enforces ^[a-zA-Z0-9_.-]+$.
// Prefer generating an implementation with an LLM; see
// pkg/agentsutils/limited_file_system.go for an example.
type Toolset interface {
	// Definitions returns the tool definitions advertised to the model: name,
	// arguments, and a description of what the tool does and when to use it.
	Definitions() []ToolDefinition

	// Dispatch executes a tool call. A returned error breaks the agentic loop;
	// put the error in ToolCallResult.Err to feed it back to the model instead.
	Dispatch(ctx context.Context, name string, arguments json.RawMessage) (ToolCallResult, error)
}

// ConversationStorage stores and loads conversation history.
type ConversationStorage interface {
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
// Available methods: Ask, Execute, CurrentConversation, Name.
type Agent struct {
	provider  Provider
	toolset   Toolset
	agentName string
	sysMsg    string

	conversationStorage ConversationStorage
	maxAgentSteps       uint64

	handleUnknownConversationElement HandleUnknownConversationElement
	handleImageGeneration            HandleImageGeneration
	inspectReq                       InspectEachRequest
	inspectResp                      InspectEachResponse
}

// HandleImageGeneration is called for image generation. The returned string is
// stored as the image identifier in place of the generated image.
type HandleImageGeneration func(ctx context.Context, image *ImageGeneration) (string, error)

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

// CurrentConversation returns the conversation history as stored.
func (a *Agent) CurrentConversation(ctx context.Context) ([]ConversationElement, error) {
	conversation, err := a.conversationStorage.Load(ctx)
	if err != nil {
		return nil, err
	}

	if len(conversation) != 0 {
		return conversation, nil
	}

	return []ConversationElement{}, nil
}

func funcResultToFunctionCallResp(callID string, funcResult any) FunctionCallResp {
	b, err := json.Marshal(funcResult)
	if err != nil {
		errorMsg := "unable to execute function; " + err.Error()
		return FunctionCallResp{Type: "function_call_output", CallID: callID, Output: errorMsg}
	}

	return FunctionCallResp{Type: "function_call_output", CallID: callID, Output: string(b)}
}

// Execute runs a prompt built with PromptBuilder, whose parameters control
// this interaction. A nil prompt or empty message returns ErrEmptyPrompt.
func (a *Agent) Execute(ctx context.Context, p *Prompt) (string, error) {
	if p == nil || p.msg == "" {
		return "", ErrEmptyPrompt
	}
	return a.run(ctx, p.msg, p.params)
}

// Ask is a minimal entry point for a plain-text question. An empty question
// returns ErrEmptyPrompt.
func (a *Agent) Ask(ctx context.Context, question string) (string, error) {
	prompt, err := NewPromptBuilder().WithMessage(question).Build()
	if err != nil {
		return "", err
	}

	return a.Execute(ctx, prompt)
}

func (a *Agent) Name() string {
	return a.agentName
}
