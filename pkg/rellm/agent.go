package rellm

import (
	"context"
	"encoding/json"
	"net/http"
)

type ReasoningEffort string

const (
	ReasoningEffortNone   ReasoningEffort = "none"
	ReasoningEffortLow    ReasoningEffort = "low"
	ReasoningEffortHigh   ReasoningEffort = "high"
	ReasoningEffortMedium ReasoningEffort = "medium"
	ReasoningEffortXHigh  ReasoningEffort = "xhigh"
)

// Model
// value for models can be found:
//   - For openrouter: https://openrouter.ai/models (curl --request GET --url 'https://openrouter.ai/api/v1/models?limit=10' | jq)
//   - For lmstudio: https://lmstudio.ai/models
//
// names used by openrouter and lmstudio are not interchangeable:
//   - lms: "google/gemma-4-26b-a4b"
//   - openrouter: "google/gemma-4-26b-a4b-it"
type Model string

// ToolCallResult is used for holding value or error returned by tool
// error stored does not break the agentic loop
type ToolCallResult struct {
	Value any
	Err   error
}

// Toolset
// OpenAI endpoints enforce tool name to match with regex: ^[a-zA-Z0-9_-]+$
// Meta endpoint enforce tool name to match with regex: ^[a-zA-Z0-9_.-]+$
// do not implement this interface by yourself
// use llm to do it for you, as example point to
// pkg/agentsutils/limited_file_system.go
// pkg/agentsutils/limited_file_system_toolset.go
// Toolset can be chained example will be provided
type Toolset interface {
	// Definitions return a list of tool definitions with information such as name arguments, and
	// description is used by llm to gain knowledge about how a given tool works and when to use it
	Definitions() []ToolDefinition

	// Dispatch dispatches tool call, direct return of error will brake the agentic loop
	// error placed inside ToolCallResult will not brake the agentic loop and will be returned into llm as a result of tool call
	Dispatch(ctx context.Context, name string, arguments json.RawMessage) (ToolCallResult, error)
}

// ConversationStorage interface used to store and load conversation history
type ConversationStorage interface {
	// Load return a list of conversation elements, error will break the agentic loop, Context can be used to cancel the operation
	Load(context.Context) ([]ConversationElement, error)

	// Append append a list of conversation elements, error will break the agentic loop, Context can be used to cancel the operation
	// Append can be called multiple times during execution of one Prompt
	Append(context.Context, []ConversationElement) error
}

// HTTPClient interface used to make http requests
// by providing own implementation it is possible to add more functionality and security
// e.g.: rate limiting, redirections, and other
type HTTPClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// Agent struct used to manage conversation and provide agentic loop
// Should be created with builder AgentBuilder pkg/rellm/builders.go
// available methods: Ask, Execute, CurrentConversation, Name
// heart of Agent is an incorporated loop that allows dispatching function calls and manage conversation
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

// HandleImageGeneration used as a callback for image generation
// returned string will be used as image identifier and stored instead of original image content
type HandleImageGeneration func(ctx context.Context, image *ImageGeneration) (string, error)

// InspectEachRequest used to inspect each request before it is sent to provider, last chance to modify or log data
type InspectEachRequest func(*ResponsesAPIReq)

// InspectEachResponse used to inspect each response before it is sent to provider, last chance to modify or log data
type InspectEachResponse func(resp *ResponsesAPIResp)

// HandleUnknownConversationElement is called for each provider output item rellm
// cannot classify. The returned slice replaces the unknown element's slot in
// the conversation:
//
//   - nil or empty: drop the element
//   - []ConversationElement{el}: keep it unchanged
//   - any other slice: replace it with those elements
type HandleUnknownConversationElement func(ctx context.Context, el *UnknownElement) ([]ConversationElement, error)

// CurrentConversation list of standard element of conversation
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

// Execute builds the user message, packages inference params
// prompt with empty message is not allowed
// usage of prompt allows to inject more parameters for given interaction
func (a *Agent) Execute(ctx context.Context, p *Prompt) (string, error) {
	if p == nil || p.msg == "" {
		return "", ErrEmptyPrompt
	}
	return a.run(ctx, p.msg, p.params)
}

// Ask is the simple entry point for quick questions — no builder needed
// easy-to-use minimal effort
// empty message is not allowed
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
