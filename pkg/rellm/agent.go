package rellm

import (
	"context"
	"encoding/json"
)

type Toolset interface {
	BuildTools() []Tool
	DispatchTools(ctx context.Context, name string, callID string, arguments json.RawMessage) (FunctionCallResp, bool)
}

type ConversationStorage interface {
	Load() ([]ConversationElement, error)
	Append([]ConversationElement) error
}

type Agent struct {
	provider  Provider
	toolset   Toolset
	agentName string
	sysMsg    string

	conversationStorage ConversationStorage
	maxAgentSteps       uint64

	handleImageGeneration HandleImageGeneration
	inspectReq            InspectEachRequest
	inspectResp           InspectEachResponse
}

// HandleImageGeneration used as a callback for image generation
// returned string will be used as image identifier and stored instead of original image content
type HandleImageGeneration func(ctx context.Context, image *ImageGeneration) (string, error)
type InspectEachRequest func(*ResponsesAPIReq)
type InspectEachResponse func(resp *ResponsesAPIResp)

func (a *Agent) CurrentConversation() ([]ConversationElement, error) {
	conversation, err := a.conversationStorage.Load()
	if err != nil {
		return nil, err
	}

	if len(conversation) != 0 {
		return conversation, nil
	}

	systemMessage, err := PromptMessageToConversation(a.sysMsg, "system")
	if err != nil {
		return nil, err
	}
	err = a.conversationStorage.Append([]ConversationElement{systemMessage})
	if err != nil {
		return nil, err
	}
	return a.conversationStorage.Load()
}

func FuncResultToFunctionCallResp(callID string, funcResult any) FunctionCallResp {
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
