package rellm

import (
	"encoding/json"

	"github.com/rs/zerolog"
)

type Toolset interface {
	BuildTools() []Tool
	DispatchTools(name string, callID string, arguments json.RawMessage) (FunctionCallResp, bool)
}

type Agent struct {
	provider     Provider
	toolset      Toolset
	logger       *zerolog.Logger
	agentName    string
	workspaceDir string
	sysMsg       string

	conversationStorage                   ConversationStorage
	maxToolsIterationWithoutReturnMessage uint64

	handleImageGeneration HandleImageGeneration
	inspectReq            InspectEachRequest
	inspectResp           InspectEachResponse
}

// HandleImageGeneration used as a callback for image generation
// returned string will be used as image identifier and stored instead of original image content
type HandleImageGeneration func(image *ImageGeneration) (string, error)
type InspectEachRequest func(*ResponsesApiReq)
type InspectEachResponse func(resp *ResponsesApiResp)

func (a *Agent) CurrentConversation() ([]json.RawMessage, error) {
	conversation, err := a.conversationStorage.Load()
	if err != nil || len(conversation) != 0 {
		return conversation, err
	}

	systemMessage, err := PromptMessageToConversation(a.sysMsg, "system")
	if err != nil {
		return nil, err
	}
	err = a.conversationStorage.Append([]json.RawMessage{systemMessage})
	if err != nil {
		return nil, err
	}
	return a.conversationStorage.Load()
}

func FuncResultToFunctionCallResp(callId string, funcResult any) FunctionCallResp {
	b, err := json.Marshal(funcResult)
	if err != nil {
		errorMsg := "unable to execute function; " + err.Error()
		return FunctionCallResp{Type: "function_call_output", CallId: callId, Output: errorMsg}
	}

	return FunctionCallResp{Type: "function_call_output", CallId: callId, Output: string(b)}
}

// Execute builds the user message, packages inference params, and hands both to
// agent.run() which owns conversation history, HTTP req assembly, and tool-loop.
func (a *Agent) Execute(p *Prompt) (string, error) {
	if p == nil || p.msg == "" {
		return "", ErrEmptyPrompt
	}
	return a.run(p.msg, p.params)
}

// Ask is the simple entry point for quick questions — no builder needed.
func (a *Agent) Ask(question string) (string, error) {
	prompt, err := NewPromptBuilder().WithMessage(question).Build()
	if err != nil {
		return "", err
	}

	return a.Execute(prompt)
}
