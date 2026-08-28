package rellm

type AgentBuilder struct {
	agent Agent
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

func (b *AgentBuilder) WithSystemMessage(msg string) *AgentBuilder {
	b.agent.sysMsg = msg
	return b
}

func (b *AgentBuilder) WithMaxAgentSteps(max uint64) *AgentBuilder {
	b.agent.maxAgentSteps = max
	return b
}

// HandleImageGeneration
func (b *AgentBuilder) WithHandleImageGeneration(handleImage HandleImageGeneration) *AgentBuilder {
	b.agent.handleImageGeneration = handleImage
	return b
}

func (b *AgentBuilder) WithInspectEachRequest(inspect InspectEachRequest) *AgentBuilder {
	b.agent.inspectReq = inspect
	return b
}

func (b *AgentBuilder) WithInspectEachResponse(inspect InspectEachResponse) *AgentBuilder {
	b.agent.inspectResp = inspect
	return b
}

func (b *AgentBuilder) WithUnknownConversationHandler(handler HandleUnknownConversationElement) *AgentBuilder {
	b.agent.handleUnknownConversationElement = handler
	return b
}

func (b *AgentBuilder) Build() (*Agent, error) {
	if b.agent.provider == nil {
		return nil, ErrBuildNoProvider
	}

	if b.agent.agentName == "" {
		return nil, ErrBuildNoAgentName
	}

	if b.agent.maxAgentSteps == 0 {
		b.agent.maxAgentSteps = DefaultMaxAgentSteps
	}

	if b.agent.conversationStorage == nil {
		return nil, ErrBuildNoConversationStorage
	}

	agent := b.agent // copy: builder stays reusable without mutating built agents
	return &agent, nil
}

const (
	DefaultMaxAgentSteps = 5
)
