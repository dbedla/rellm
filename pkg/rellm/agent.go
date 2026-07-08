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

// Prompt returns a new prompt for this agent. Use fluent methods to configure
// inference params, then call Execute() to run the full conversation loop.
func (a *Agent) Prompt(question string) *prompt {
	return newPrompt(question, a)
}

type prompt struct {
	msg    string
	agent  *Agent
	params promptParams
}

func newPrompt(msg string, agent *Agent) *prompt {
	return &prompt{msg: msg, agent: agent}
}

// WithTemperature sets sampling temperature for all calls in this prompt's loop.
func (p *prompt) WithTemperature(t float32) *prompt {
	p.params.Temperature = t
	return p
}

// WithReasoning sets reasoning effort level ("low", "medium", "high") for the entire loop.
func (p *prompt) WithReasoning(effort string) *prompt {
	if effort != "" {
		p.params.Reasoning = &ReasoningConfig{Effort: effort}
	}
	return p
}

// WithMaxOutputTokens sets max tokens for every call in this prompt's loop.
func (p *prompt) WithMaxOutputTokens(n int) *prompt {
	p.params.MaxOutputTokens = n
	return p
}

// WithTopP sets nucleus sampling parameter.
func (p *prompt) WithTopP(t float32) *prompt {
	p.params.TopP = t
	return p
}

// WithPresencePenalty sets presence penalty for every call in this prompt's loop.
func (p *prompt) WithPresencePenalty(penalty float32) *prompt {
	p.params.PresencePenalty = penalty
	return p
}

// WithFrequencyPenalty sets frequency penalty for every call in this prompt's loop.
func (p *prompt) WithFrequencyPenalty(f float32) *prompt {
	p.params.FrequencyPenalty = f
	return p
}

// WithSeed sets deterministic seed applied to every call in the loop.
func (p *prompt) WithSeed(seed int64) *prompt {
	s := seed
	p.params.Seed = &s
	return p
}

// WithLogprobs enables log probabilities output on every call in this prompt's loop.
func (p *prompt) WithLogprobs(enabled bool) *prompt {
	p.params.Logprobs = enabled
	return p
}

// WithTopLogprobs sets number of top log probabilities to return per call.
func (p *prompt) WithTopLogprobs(n int) *prompt {
	p.params.TopLogprobs = n
	return p
}

// Execute builds the user message, packages inference params, and hands both to
// agent.run() which owns conversation history, HTTP req assembly, and tool-loop.
func (p *prompt) Execute() (string, error) {
	defer p.expire()
	if p.isExpired() {
		return "", ErrPromptIsExpired
	}

	return p.agent.run(p.msg, p.params)
}

func (p *prompt) expire() {
	p.agent = nil
	p.msg = ""
}

func (p *prompt) isExpired() bool {
	return p.agent == nil || len(p.msg) == 0
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
