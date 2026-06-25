package rellm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"rellm/pkg/rellm/conversation_storage"

	"github.com/rs/zerolog"
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

func (b *EndpointBuilder) WithResponsesApiEndpoint(endpoint ResponsesApiEndpoint) *EndpointBuilder {
	b.endpoint.rae = endpoint
	return b
}

func (b *EndpointBuilder) Build() (*Endpoint, error) {

	if b.endpoint.client == nil {
		b.endpoint.client = &http.Client{}
	}

	if b.endpoint.model == "" {
		return nil, fmt.Errorf("missing model for endpoint")

	}
	if b.endpoint.rae == nil {
		return nil, fmt.Errorf("missing response api endpoint for endpoint")
	}

	return &b.endpoint, nil
}

type AgentBuilder struct {
	agent              Agent
	useWorkspaceLogger bool
	useStdoutLogger    bool
	useNoOpLogger      bool
}

func NewAgentBuilder() *AgentBuilder {
	return &AgentBuilder{}
}

func (b *AgentBuilder) WithEndpoint(endpoint *Endpoint) *AgentBuilder {
	b.agent.endpoint = endpoint
	return b
}

func (b *AgentBuilder) WithToolset(toolset Toolset) *AgentBuilder {
	b.agent.toolset = toolset
	return b
}

func (b *AgentBuilder) WithConversationStorage(storage *conversation_storage.ConversationStorage) *AgentBuilder {
	b.agent.conversationStorage = storage
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

func (b *AgentBuilder) WithMaxToolsIterationWithoutReturnMessage(max uint64) *AgentBuilder {
	b.agent.maxToolsIterationWithoutReturnMessage = max
	return b
}

func (b *AgentBuilder) WithCustomLogger(logger *zerolog.Logger) *AgentBuilder {
	b.agent.logger = logger
	return b
}

func (b *AgentBuilder) WithWorkspaceLogger() *AgentBuilder {
	b.useWorkspaceLogger = true
	return b
}

func (b *AgentBuilder) WithStdoutLogger() *AgentBuilder {
	b.useStdoutLogger = true
	return b
}

func (b *AgentBuilder) WithNoOpLogger() *AgentBuilder {
	b.useNoOpLogger = true
	return b
}

func (b *AgentBuilder) Build() (*Agent, error) {

	if b.agent.workspaceDir == "" {
		return nil, fmt.Errorf("missing workspaceDir")
	}

	if b.agent.endpoint == nil {
		return nil, fmt.Errorf("missing endpoint")
	}

	if b.agent.agentName == "" {
		b.agent.agentName = "UnnamedAgent"
	}

	l, err := b.buildLogger()
	if err != nil {
		return nil, err
	}
	b.agent.logger = l

	if b.agent.maxToolsIterationWithoutReturnMessage == 0 {
		b.agent.maxToolsIterationWithoutReturnMessage = defaultMaxToolsIterationWithoutReturnMessage
	}

	if b.agent.conversationStorage == nil {
		b.agent.conversationStorage = conversation_storage.NewForAgent(b.agent.workspaceDir, b.agent.agentName)
	}

	b.agent.logger.Info().Msgf("===== New agent %s ready to action =====", b.agent.agentName)

	return &b.agent, nil
}

func (b *AgentBuilder) buildLogger() (*zerolog.Logger, error) {
	multipleLoggerConfig := !exactlyOneIsSet(b.useStdoutLogger, b.useWorkspaceLogger, b.useNoOpLogger, b.agent.logger != nil)
	if multipleLoggerConfig {
		return nil, fmt.Errorf("exactly one logger must be configured")
	}

	if b.useStdoutLogger {
		l := NewBaseStdOutLogger()
		return &l, nil
	}

	if b.useWorkspaceLogger {
		return b.buildWorkspaceLogger()
	}

	if b.useNoOpLogger {
		l := NewNoOpLogger()
		return &l, nil
	}

	return b.agent.logger, nil
}

func (b *AgentBuilder) buildWorkspaceLogger() (*zerolog.Logger, error) {
	if b.agent.workspaceDir == "" {
		return nil, fmt.Errorf("missing workspaceDir")
	}

	fp := filepath.Join(b.agent.workspaceDir, b.agent.agentName+".log")
	fl, err := NewBaseFileLogger(fp)
	if err != nil {
		return nil, fmt.Errorf("unable to create logger: %s", err.Error())
	}
	l := NewComponentLogger(fl, b.agent.agentName)
	return &l, nil
}

func exactlyOneIsSet(flags ...bool) bool {
	count := 0
	for _, f := range flags {
		if f {
			count++
		}
	}
	return count == 1
}

const (
	defaultMaxToolsIterationWithoutReturnMessage = 5
)
