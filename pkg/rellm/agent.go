package rellm

import (
	"encoding/json"
	"rellm/pkg/rellm/conversation_storage"

	"github.com/rs/zerolog"
)

type Toolset interface {
	BuildTools() []Tool
	DispatchTools(name string, callID string, arguments json.RawMessage) (FunctionCallResp, bool)
}

type FunctionCallResp struct {
	Type   string `json:"type"`
	CallId string `json:"call_id"`
	Output string `json:"output"`
}

type ImageGenerationResp struct {
	Type   string `json:"type"`
	Id     string `json:"id"`
	Status string `json:"status"`
	Result string `json:"result"`
}

type Agent struct {
	endpoint                              *Endpoint
	toolset                               Toolset
	logger                                *zerolog.Logger
	conversationStorage                   *conversation_storage.ConversationStorage
	agentName                             string
	workspaceDir                          string
	sysMsg                                string
	inMemoryConversation                  []json.RawMessage
	continueConversation                  bool
	maxToolsIterationWithoutReturnMessage uint64
}

func (a *Agent) Ask(question string) (string, error) {

	msg, _, err := a.AskLikeAPro(question, nil, nil)

	return msg, err
}

type HandleImage func(image OutputItem) (string, error)
type InspectEachRequest func(*ResponsesApiReq)
type InspectEachResponse func(resp *ResponsesApiResp)

func (a *Agent) AskLikeAPro(question string, inspectReq InspectEachRequest, inspectResp InspectEachResponse) (string, *ResponsesApiResp, error) {
	a.logger.Info().Msgf("question to agent: %s", question)
	defer a.logger.Info().Msg("question answered")

	conversation := a.CurrentConversation()

	userMsg, err := PromptMessageToConversation(question, "user")
	if err != nil {
		a.logger.Error().Err(err).Msgf("unable to build conversation %s", err.Error())
		return "", nil, err
	}
	conversation = append(conversation, userMsg)

	interactions := requestedInteractions{}
	interactions.inspectReq = inspectReq
	interactions.inspectResp = inspectResp

	newConversation, msg, rawResp, err := a.process(conversation, interactions)

	if len(newConversation) > 0 {
		a.inMemoryConversation = newConversation
	}

	if err != nil {
		a.logger.Error().Err(err).Msgf("unable to process conversation %s", err.Error())
		return "", rawResp, err
	}

	a.logger.Info().Msgf("message: %s", msg)

	return msg, rawResp, nil
}

func (a *Agent) CurrentConversation() []json.RawMessage {
	conversation := a.inMemoryConversation

	if len(conversation) == 0 && a.continueConversation {
		restoredConversation, err := a.conversationStorage.RestoreLatest()
		if err != nil {
			a.logger.Error().Err(err).Msg("unable to restore conversation from workspace")
		}
		a.logger.Info().Msgf("from workspace conv len: %d", len(restoredConversation))
		conversation = restoredConversation
	}
	if len(conversation) == 0 {
		brandNewConversation, err := PromptMessageToConversation(a.sysMsg, "system")
		if err != nil {
			a.logger.Error().Err(err).Msg("unable to build conversation")
		}
		a.logger.Info().Msgf("brand new conversation started")
		conversation = []json.RawMessage{brandNewConversation}
	}

	return conversation
}

func (a *Agent) StoreConversation() error {
	err := a.conversationStorage.Store(a.inMemoryConversation)
	if err != nil {
		a.logger.Error().Err(err).Msg("unable to store conversation")
		return err
	}
	a.logger.Debug().Msgf("conversation stored, check workspace folder: %s", a.workspaceDir)
	return nil
}

func FuncResultToFunctionCallResp(callId string, funcResult any) FunctionCallResp {
	b, err := json.Marshal(funcResult)
	if err != nil {
		errorMsg := "unable to execute function; " + err.Error()
		return FunctionCallResp{Type: "function_call_output", CallId: callId, Output: errorMsg}
	}

	return FunctionCallResp{Type: "function_call_output", CallId: callId, Output: string(b)}
}

func (a *Agent) Prompt(question string) *Prompt {
	return newPrompt(question, a)
}

func (a *Agent) promptWithInteractions(question string, interactions requestedInteractions) (string, *ResponsesApiResp, error) {
	a.logger.Info().Msgf("question to agent: %s", question)
	defer a.logger.Info().Msg("question answered")

	conversation := a.CurrentConversation()

	userMsg, err := PromptMessageToConversation(question, "user")
	if err != nil {
		a.logger.Error().Err(err).Msgf("unable to build conversation %s", err.Error())
		return "", nil, err
	}
	conversation = append(conversation, userMsg)

	newConversation, msg, rawResp, err := a.process(conversation, interactions)

	if len(newConversation) > 0 {
		a.inMemoryConversation = newConversation
	}

	if err != nil {
		a.logger.Error().Err(err).Msgf("unable to process conversation %s", err.Error())
		return "", rawResp, err
	}

	a.logger.Info().Msgf("message: %s", msg)

	return msg, rawResp, nil
}

type requestedInteractions struct {
	inspectReq  InspectEachRequest
	inspectResp InspectEachResponse
	handleImage HandleImage
}
type Prompt struct {
	msg   string
	agent *Agent

	interactions requestedInteractions
}

func newPrompt(msg string, agent *Agent) *Prompt {
	return &Prompt{msg: msg, agent: agent}
}
func (p *Prompt) WithInspectReq(inspectReq InspectEachRequest) *Prompt {
	p.interactions.inspectReq = inspectReq
	return p
}
func (p *Prompt) WithInspectResp(inspectResp InspectEachResponse) *Prompt {
	p.interactions.inspectResp = inspectResp
	return p
}
func (p *Prompt) WithHandleImage(handleImage HandleImage) *Prompt {
	p.interactions.handleImage = handleImage
	return p
}

func (p *Prompt) Execute() (string, *ResponsesApiResp, error) {
	return p.agent.promptWithInteractions(p.msg, p.interactions)
}
