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
	handleImage                           HandleImage
}

func (a *Agent) Ask(question string) (string, error) {

	msg, _, err := a.AskLikeAPro(question, nil, nil)

	return msg, err
}

// HandleImage used as a callback for image generation
// returned string will be used as image identifier in conversation
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

	req := &ResponsesApiReq{
		Model: string(a.endpoint.model),
		Input: conversation,
	}
	if a.toolset != nil {
		req.Tools = a.toolset.BuildTools()
	}

	interactions := requestedInteractions{}
	interactions.inspectReq = inspectReq
	interactions.inspectResp = inspectResp

	newConversation, msg, rawResp, err := a.process(req, interactions)

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

// Prompt returns a new prompt for this agent. Use fluent methods to configure
// inference params, then call Execute() to run the full conversation loop.
func (a *Agent) Prompt(question string) *prompt {
	return newPrompt(question, a)
}

type requestedInteractions struct {
	inspectReq  InspectEachRequest
	inspectResp InspectEachResponse
	handleImage HandleImage
}

// prompt is the private handle for building and executing a single logical
// conversation turn. It holds an internal partial ResponsesApiReq that gets
// filled incrementally via WithXxx methods; at Execute time only Input is set
// (conversation-specific) before passing to process().
type prompt struct {
	msg   string
	agent *Agent

	// Internal partial request, built up by WithXxx calls. Only Model + Tools are
	// set at init; inference params are added as the user chains fluent methods.
	req *ResponsesApiReq
}

func newPrompt(msg string, agent *Agent) *prompt {
	p := prompt{msg: msg, agent: agent}
	p.req = &ResponsesApiReq{
		Model: string(p.agent.endpoint.model),
	}

	if p.agent.toolset != nil {
		p.req.Tools = p.agent.toolset.BuildTools()
	}
	return &p
}

// WithTemperature sets sampling temperature for all calls in this prompt's loop.
func (p *prompt) WithTemperature(t float32) *prompt {
	p.req.Temperature = t
	return p
}

// WithReasoning sets reasoning effort level ("low", "medium", "high") for the entire loop.
func (p *prompt) WithReasoning(effort string) *prompt {
	if effort != "" {
		p.req.Reasoning = &ReasoningConfig{Effort: effort}
	}
	return p
}

// WithMaxOutputTokens sets max tokens for every call in this prompt's loop.
func (p *prompt) WithMaxOutputTokens(n int) *prompt {
	p.req.MaxOutputTokens = n
	return p
}

// WithTopP sets nucleus sampling parameter.
func (p *prompt) WithTopP(t float32) *prompt {
	p.req.TopP = t
	return p
}

// WithPresencePenalty sets presence penalty for every call in this prompt's loop.
func (p *prompt) WithPresencePenalty(penalty float32) *prompt {
	p.req.PresencePenalty = penalty
	return p
}

// WithFrequencyPenalty sets frequency penalty for every call in this prompt's loop.
func (p *prompt) WithFrequencyPenalty(f float32) *prompt {
	p.req.FrequencyPenalty = f
	return p
}

// WithSeed sets deterministic seed applied to every call in the loop.
func (p *prompt) WithSeed(seed int64) *prompt {
	s := seed
	p.req.Seed = &s
	return p
}

// WithLogprobs enables log probabilities output on every call in this prompt's loop.
func (p *prompt) WithLogprobs(enabled bool) *prompt {
	p.req.Logprobs = enabled
	return p
}

// WithTopLogprobs sets number of top log probabilities to return per call.
func (p *prompt) WithTopLogprobs(n int) *prompt {
	p.req.TopLogprobs = n
	return p
}

// Execute runs the full conversation loop: sets Input on internal request,
// sends to API, handles tool calls in a loop until a final message or max iterations.
func (p *prompt) Execute() (string, error) {
	defer p.expire()
	if p.isExpired() {
		return "", ErrPromptIsExpired
	}

	p.agent.logger.Info().Msgf("question to agent: %s", p.msg)
	defer p.agent.logger.Info().Msg("question answered")

	conversation := p.agent.CurrentConversation()

	userMsg, err := PromptMessageToConversation(p.msg, "user")
	if err != nil {
		p.agent.logger.Error().Err(err).Msgf("unable to build conversation %s", err.Error())
		return "", err
	}
	conversation = append(conversation, userMsg)

	// Set Input (conversation-specific) on internal request; same pointer
	// passed through the entire loop — only Input mutates between iterations.
	p.req.Input = conversation

	interactions := requestedInteractions{
		handleImage: p.agent.handleImage,
	}

	newConversation, msg, _, err := p.agent.process(p.req, interactions)

	if len(newConversation) > 0 {
		p.agent.inMemoryConversation = newConversation
	}

	if err != nil {
		p.agent.logger.Error().Err(err).Msgf("unable to process conversation %s", err.Error())
		return "", err
	}

	p.agent.logger.Info().Msgf("message: %s", msg)

	return msg, nil
}

func (p *prompt) expire() {
	p.agent = nil
	p.msg = ""
	p.req = nil
}

func (p *prompt) isExpired() bool {
	return p.agent == nil || len(p.msg) == 0
}
