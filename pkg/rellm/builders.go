package rellm

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"rellm/pkg/rellm/conversation_storage"
)

type EndpointBuilder struct {
	endpoint Endpoint
}

func NewEndpointBuilder() *EndpointBuilder {
	return &EndpointBuilder{}
}

func (b *EndpointBuilder) WithClientHttpDo(c ClientHttpDo) *EndpointBuilder {
	b.endpoint.client = c
	return b
}

func (b *EndpointBuilder) WithModel(model Model) *EndpointBuilder {
	b.endpoint.model = model
	return b
}

func (b *EndpointBuilder) WithResponseApiEndpoint(endpoint ResponseApiEndpoint) *EndpointBuilder {
	b.endpoint.rae = endpoint
	return b
}

func (b *EndpointBuilder) Build() *Endpoint {

	if b.endpoint.client == nil {
		b.endpoint.client = &http.Client{}
	}

	if b.endpoint.model == "" {
		panic("missing model for endpoint")
	}
	if b.endpoint.rae == nil {
		panic("missing response api endpoint for endpoint")
	}

	return &b.endpoint
}

// ToDo: should provide template for post req not own structure
func NewConversationParameterBuilder() *ConversationParameterBuilder {
	return &ConversationParameterBuilder{}
}

type ConversationParameterBuilder struct {
	cp ConversationParameters
}

func (b *ConversationParameterBuilder) WithToolset(tools []Tool) *ConversationParameterBuilder {
	b.cp.tools = tools
	return b
}

func (b *ConversationParameterBuilder) WithReasoning(effort ReasoningEffort) *ConversationParameterBuilder {
	if effort == "" {
		return b
	}

	r := ReasoningConfig{Effort: string(effort)}
	b.cp.Reasoning = &r

	return b
}

func (b *ConversationParameterBuilder) Build() *ConversationParameters {
	return &b.cp
}

type AgentBuilder struct {
	agent Agent
}

func NewAgentBuilder() *AgentBuilder {
	return &AgentBuilder{}
}

func (b *AgentBuilder) WithEndpoint(endpoint *Endpoint) *AgentBuilder {
	b.agent.endpoint = endpoint
	return b
}

func (b *AgentBuilder) WithConversationParameters(params *ConversationParameters) *AgentBuilder {
	b.agent.conversationParameters = params
	return b
}

// WithStdOutLogger set logger output as stdout, by default, the log is in the *.log file in the workspace
func (b *AgentBuilder) WithStdOutLogger() *AgentBuilder {
	if b.agent.logger != nil {
		panic("logger already set")
	}
	l := NewBaseStdOutLogger()
	b.agent.logger = &l
	return b
}

func (b *AgentBuilder) WithConversationStorage(storage *conversation_storage.ConversationStorage) *AgentBuilder {
	b.agent.conversationStorage = storage
	return b
}

func (b *AgentBuilder) WithToolsDispatcher(tools func(string, callID string, arguments string) (FunctionCallResp, bool)) *AgentBuilder {
	b.agent.toolDispatcher = tools
	return b
}

func (b *AgentBuilder) WithAgentName(name string) *AgentBuilder {
	b.agent.agentName = name
	return b
}

func (b *AgentBuilder) WithWorkspaceDir(dir string) *AgentBuilder {
	b.agent.workspaceDir = dir
	return b
}

func (b *AgentBuilder) WithSystemMessage(msg string) *AgentBuilder {
	b.agent.sysMsg = msg
	return b
}

func (b *AgentBuilder) WithInMemoryConversation(conv []json.RawMessage) *AgentBuilder {
	b.agent.inMemoryConversation = conv
	return b
}

func (b *AgentBuilder) WithContinueConversation(cont bool) *AgentBuilder {
	b.agent.continueConversation = cont
	return b
}

func (b *AgentBuilder) WithMaxToolsIterationWithoutReturnMessage(max int) *AgentBuilder {
	b.agent.maxToolsIterationWithoutReturnMessage = max
	return b
}

func (b *AgentBuilder) Build() *Agent {

	if b.agent.workspaceDir == "" {
		panic("missing workspaceDir")
	}

	if b.agent.endpoint == nil {
		panic("missing endpoint")
	}

	if b.agent.agentName == "" {
		b.agent.agentName = "UnnamedAgent"
	}

	if b.agent.logger == nil {
		fp := filepath.Join(b.agent.workspaceDir, b.agent.agentName+".log")
		fl := NewBaseFileLogger(fp)
		l := NewComponentLogger(fl, b.agent.agentName)
		b.agent.logger = &l
	}

	if b.agent.maxToolsIterationWithoutReturnMessage == 0 {
		b.agent.maxToolsIterationWithoutReturnMessage = 5
	}

	if b.agent.conversationStorage == nil {
		b.agent.conversationStorage = conversation_storage.NewForAgent(b.agent.workspaceDir, b.agent.agentName)
	}

	b.agent.logger.Info().Msgf("===== New agent %s ready to action =====", b.agent.agentName)

	return &b.agent
}
