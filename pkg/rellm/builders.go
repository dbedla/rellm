package rellm

import "context"

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

// WithUnknownConversationElementDrop configures the agent to drop unknown
// conversation elements returned by the provider.
func (b *AgentBuilder) WithUnknownConversationElementDrop() *AgentBuilder {
	b.agent.handleUnknownConversationElement = func(ctx context.Context, el *UnknownElement) ([]ConversationElement, error) {
		return nil, nil
	}
	return b
}

// WithUnknownConversationElementKeepInTheLoop configures the agent to keep
// unknown conversation elements unchanged in the conversation loop.
func (b *AgentBuilder) WithUnknownConversationElementKeepInTheLoop() *AgentBuilder {
	b.agent.handleUnknownConversationElement = func(ctx context.Context, el *UnknownElement) ([]ConversationElement, error) {
		return []ConversationElement{el}, nil
	}
	return b
}

// WithUnknownConversationElementHandler sets a custom handler for unknown conversation
// elements. The returned slice replaces the unknown element's slot in the
// conversation; nil or an empty slice drops it.
func (b *AgentBuilder) WithUnknownConversationElementHandler(handler HandleUnknownConversationElement) *AgentBuilder {
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

	if b.agent.handleUnknownConversationElement == nil {
		b.agent.handleUnknownConversationElement = func(ctx context.Context, el *UnknownElement) ([]ConversationElement, error) {
			return nil, ErrNoUnknownConversationElementHandler
		}
	}

	agent := b.agent // copy: builder stays reusable without mutating built agents
	return &agent, nil
}

const (
	DefaultMaxAgentSteps = 5
)
