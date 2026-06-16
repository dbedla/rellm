package rellm

import (
	"encoding/json"
	"rellm/pkg/rellm/conversation_storage"

	"github.com/rs/zerolog"
)

type Toolset interface {
	BuildTools() []Tool
	DispatchTools(name string, callID string, arguments string) (FunctionCallResp, bool)
}

type FunctionCallResp struct {
	Type   string `json:"type"`
	CallId string `json:"call_id"`
	Output string `json:"output"`
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
	maxToolsIterationWithoutReturnMessage int
}

func (a *Agent) Ask(question string) (string, error) {

	nopProParamSet := func(r *ResponsesApiReq) {}
	msg, _, err := a.AskLikeAPro(question, nopProParamSet)

	return msg, err
}

type FuncLikeProSet func(*ResponsesApiReq)

func (a *Agent) AskLikeAPro(question string, proParameterSet FuncLikeProSet) (string, *ResponsesApiResp, error) {
	a.logger.Info().Msgf("question to agent: %s", question)
	defer a.logger.Info().Msg("question answered")

	conversation := a.CurrentConversation()

	userMsg, err := PromptMessageToConversation(question, "user")
	if err != nil {
		a.logger.Error().Err(err).Msgf("unable to build conversation %s", err.Error())
		return "", nil, err
	}
	conversation = append(conversation, userMsg)

	newConversation, msg, rawResp, err := a.Process(conversation, proParameterSet)

	if err != nil {
		a.logger.Error().Err(err).Msgf("unable to process conversation %s", err.Error())
		return "", nil, err
	}

	a.inMemoryConversation = newConversation

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
		return FunctionCallResp{Type: "function_call_output", CallId: callId, Output: "unable to execute function; " + err.Error()}
	}

	return FunctionCallResp{Type: "function_call_output", CallId: callId, Output: string(b)}
}
