package rellm

import "context"

// AgentBuilder constructs an Agent. It is the recommended way to create an
// agent; Build validates mandatory fields before returning it.
//
//	agent, err := rellm.NewAgentBuilder().
//		WithProvider(provider).
//		WithConversationStorage(rellm.NewInMemoryStorage()).
//		WithUnknownConversationElementKeepInTheLoop(). // other policies: WithUnknownConversationElementDrop, WithUnknownConversationElementHandler
//		Build()
type AgentBuilder struct {
	agent Agent
}

// NewAgentBuilder starts a builder chain for constructing an Agent.
func NewAgentBuilder() *AgentBuilder {
	return &AgentBuilder{}
}

// WithProvider sets the provider the agent will call. Required.
func (b *AgentBuilder) WithProvider(provider Provider) *AgentBuilder {
	b.agent.provider = provider
	return b
}

// WithToolset sets the toolset used to dispatch tool calls. Optional.
func (b *AgentBuilder) WithToolset(toolset Toolset) *AgentBuilder {
	b.agent.toolset = toolset
	return b
}

// WithConversationStorage sets the storage used to persist and load the
// conversation history. Required.
func (b *AgentBuilder) WithConversationStorage(storage ConversationStorage) *AgentBuilder {
	b.agent.conversationStorage = storage
	return b
}

// WithAgentName sets the agent name, used for identification. Optional.
func (b *AgentBuilder) WithAgentName(name string) *AgentBuilder {
	b.agent.agentName = name
	return b
}

// WithSystemMessage sets the system message, which carries standing
// instructions and context for every request this agent makes. Optional.
func (b *AgentBuilder) WithSystemMessage(msg string) *AgentBuilder {
	b.agent.sysMsg = msg
	return b
}

// WithMaxAgentSteps sets the maximum number of steps per prompt. Optional;
// defaults to DefaultMaxAgentSteps. A step is one provider request plus its
// tool call results; reaching the limit returns ErrMaxAgentStepsReached.
func (b *AgentBuilder) WithMaxAgentSteps(max uint64) *AgentBuilder {
	b.agent.maxAgentSteps = max
	return b
}

// WithHandleImageGeneration sets the image handler; see HandleImageGeneration.
func (b *AgentBuilder) WithHandleImageGeneration(handleImage HandleImageGeneration) *AgentBuilder {
	b.agent.handleImageGeneration = handleImage
	return b
}

// WithInspectEachRequest sets the request inspector; see InspectEachRequest.
func (b *AgentBuilder) WithInspectEachRequest(inspect InspectEachRequest) *AgentBuilder {
	b.agent.inspectReq = inspect
	return b
}

// WithInspectEachResponse sets the response inspector; see InspectEachResponse.
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

// Build finishes the chain and returns a ready Agent, or an error. Required
// fields are set via WithProvider and WithConversationStorage; optional fields
// fall back to defaults. Build returns a copy, so the builder can be reused to
// create another agent.
func (b *AgentBuilder) Build() (*Agent, error) {
	if b.agent.provider == nil {
		return nil, ErrBuildNoProvider
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
	// DefaultMaxAgentSteps is the default maximum number of steps an agent may
	// take per prompt before reporting ErrMaxAgentStepsReached.
	DefaultMaxAgentSteps = 5
)
