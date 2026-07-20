package rellm

import (
	"encoding/json"
	"errors"
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

type ImageGenerationConversationPlaceholder struct {
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

	handleImage HandleImage
	inspectReq  InspectEachRequest
	inspectResp InspectEachResponse
}

// HandleImage used as a callback for image generation
// returned string will be used as image identifier in conversation
type HandleImage func(image OutputItem) (string, error)
type InspectEachRequest func(*ResponsesApiReq)
type InspectEachResponse func(resp *ResponsesApiResp)

func (a *Agent) run(msg string, params promptParams) (string, error) {
	a.logger.Info().Msgf("question to agent: %s", string(msg))
	defer a.logger.Info().Msg("question answered")

	conversation := a.CurrentConversation()

	userMsg, err := PromptMessageToConversation(msg, "user")
	if err != nil {
		return "", errors.Join(ErrUserMsgConversionFailed, err)
	}

	conversation = append(conversation, userMsg)

	req := toBaseResponsesApiReq(params, a.endpoint.model, conversation)

	if a.toolset != nil {
		req.Tools = a.toolset.BuildTools()
	}

	newConversation, msgRespFromLLM, _, err := a.process(req)

	if len(newConversation) > 0 {
		a.inMemoryConversation = newConversation
	}

	if err != nil {
		a.logger.Error().Err(err).Msgf("unable to process conversation %s", err.Error())
		return "", err
	}

	a.logger.Info().Msgf("message: %s", msgRespFromLLM)
	return msgRespFromLLM, nil
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
	return a.Execute(&Prompt{msg: question, params: defaultParams})
}

func toBaseResponsesApiReq(params promptParams, model Model, conversation []json.RawMessage) *ResponsesApiReq {
	return &ResponsesApiReq{
		Model:            string(model),
		Input:            conversation,
		Temperature:      params.Temperature,
		Reasoning:        params.Reasoning,
		MaxOutputTokens:  params.MaxOutputTokens,
		TopP:             params.TopP,
		PresencePenalty:  params.PresencePenalty,
		FrequencyPenalty: params.FrequencyPenalty,
		Seed:             params.Seed,
		Logprobs:         params.Logprobs,
		TopLogprobs:      params.TopLogprobs,
	}
}
