package rellm

import (
	"context"
	"encoding/json"
)

type ToolCallResult struct {
	Value any
	Err   error
}

type Toolset interface {
	Definitions() []ToolDefinition
	Dispatch(ctx context.Context, name string, arguments json.RawMessage) (ToolCallResult, error)
}

type ConversationStorage interface {
	Load(context.Context) ([]ConversationElement, error)
	Append(context.Context, []ConversationElement) error
}

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
type InspectEachRequest func(*ResponsesAPIReq)
type InspectEachResponse func(resp *ResponsesAPIResp)

// HandleUnknownConversationElement is called for each provider output item rellm
// cannot classify. The returned slice replaces the unknown element's slot in
// the conversation:
//
//   - nil or empty: drop the element
//   - []ConversationElement{el}: keep it unchanged
//   - any other slice: replace it with those elements
type HandleUnknownConversationElement func(ctx context.Context, el *UnknownElement) ([]ConversationElement, error)

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

// Execute builds the user message, packages inference params, and hands both to
// agent.run() which owns conversation history, HTTP req assembly, and tool-loop.
func (a *Agent) Execute(ctx context.Context, p *Prompt) (string, error) {
	if p == nil || p.msg == "" {
		return "", ErrEmptyPrompt
	}
	return a.run(ctx, p.msg, p.params)
}

// Ask is the simple entry point for quick questions — no builder needed.
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
