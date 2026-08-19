package rellm

type AgentBuilder struct {
	agent       Agent
	inspectReq  InspectEachRequest
	inspectResp InspectEachResponse
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

	return &b.agent, nil
}

const (
	//todo: max steps rename
	defaultMaxToolsIterationWithoutReturnMessage = 5
)
