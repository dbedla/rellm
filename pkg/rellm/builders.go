package rellm

import (
	"errors"
	"path/filepath"

	"github.com/rs/zerolog"
)

type AgentBuilder struct {
	agent              Agent
	useWorkspaceLogger bool
	useStdoutLogger    bool
	useNoOpLogger      bool
	inspectReq         InspectEachRequest
	inspectResp        InspectEachResponse
}

func NewAgentBuilder() *AgentBuilder {
	return &AgentBuilder{}
}

func (b *AgentBuilder) WithProvider(provider Provider) *AgentBuilder {
	b.agent.provider = provider
	return b
}

func (b *AgentBuilder) WithToolset(toolset Toolset) *AgentBuilder {
	b.agent.toolset = toolset
	return b
}

func (b *AgentBuilder) WithConversationStorage(storage ConversationStorage) *AgentBuilder {
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

func (b *AgentBuilder) WithMaxToolsIterationWithoutReturnMessage(max uint64) *AgentBuilder {
	b.agent.maxToolsIterationWithoutReturnMessage = max
	return b
}

// HandleImageGeneration
func (b *AgentBuilder) WithHandleImageGeneration(handleImage HandleImageGeneration) *AgentBuilder {
	b.agent.handleImageGeneration = handleImage
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

func (b *AgentBuilder) WithInspectEachRequest(inspect InspectEachRequest) *AgentBuilder {
	b.inspectReq = inspect
	return b
}

func (b *AgentBuilder) WithInspectEachResponse(inspect InspectEachResponse) *AgentBuilder {
	b.inspectResp = inspect
	return b
}

func (b *AgentBuilder) Build() (*Agent, error) {

	if b.agent.workspaceDir == "" {
		return nil, ErrBuildNoWorkspaceDir
	}

	if b.agent.provider == nil {
		return nil, ErrBuildNoProvider
	}

	if b.agent.agentName == "" {
		return nil, ErrBuildNoAgentName
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
		return nil, ErrBuildNoConversationStorage
	}

	if b.inspectReq != nil {
		b.agent.inspectReq = b.inspectReq
	}

	if b.inspectResp != nil {
		b.agent.inspectResp = b.inspectResp
	}

	b.agent.logger.Info().Msgf("===== New agent %s ready to action =====", b.agent.agentName)

	return &b.agent, nil
}

func (b *AgentBuilder) buildLogger() (*zerolog.Logger, error) {
	multipleLoggerConfig := !exactlyOneIsSet(b.useStdoutLogger, b.useWorkspaceLogger, b.useNoOpLogger, b.agent.logger != nil)
	if multipleLoggerConfig {
		return nil, ErrBuildExactlyOneLogger
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
		return nil, ErrBuildNoWorkspaceDir
	}

	fp := filepath.Join(b.agent.workspaceDir, b.agent.agentName+".log")
	fl, err := NewBaseFileLogger(fp)
	if err != nil {
		return nil, errors.Join(ErrBuildLogger, err)
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
	//todo: max steps rename
	defaultMaxToolsIterationWithoutReturnMessage = 5
)
