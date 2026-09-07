package rellm

import "context"

// AgentBuilder is a builder for Agent
// it is recomended way to construct agent, provide basic validation
// example of usage:
// todo: add some example of usage
type AgentBuilder struct {
	agent Agent
}

// NewAgentBuilder method should be used as start of chain during agent building
func NewAgentBuilder() *AgentBuilder {
	return &AgentBuilder{}
}

// WithProvider method sets provider for agent, mandatory
func (b *AgentBuilder) WithProvider(provider Provider) *AgentBuilder {
	b.agent.provider = provider
	return b
}

// WithToolset method sets toolset for agent, optional
func (b *AgentBuilder) WithToolset(toolset Toolset) *AgentBuilder {
	b.agent.toolset = toolset
	return b
}

// WithConversationStorage method sets conversation storage for agent, mandatory
func (b *AgentBuilder) WithConversationStorage(storage ConversationStorage) *AgentBuilder {
	b.agent.conversationStorage = storage
	return b
}

// WithAgentName method sets agent name, optional,
// can be used to identify agent
func (b *AgentBuilder) WithAgentName(name string) *AgentBuilder {
	b.agent.agentName = name
	return b
}

// WithSystemMessage method sets system message for agent, optional
// system message should contain all context for agent required to perform tasks
// todo: elaborate more about system message
func (b *AgentBuilder) WithSystemMessage(msg string) *AgentBuilder {
	b.agent.sysMsg = msg
	return b
}

// WithMaxAgentSteps method sets maximum number of steps agent can take, optional, defoult to DefaultMaxAgentSteps = 5
// it means that agent cycle at most 5 times before return, if limit reached will return error ErrMaxAgentStepsReached
// cycle means sending request to llm with tool call result, it is possible to handle more than one tool call result in one cycle
func (b *AgentBuilder) WithMaxAgentSteps(max uint64) *AgentBuilder {
	b.agent.maxAgentSteps = max
	return b
}

// HandleImageGeneration sets te user defined image handler for more detail see HandleImageGeneration
func (b *AgentBuilder) WithHandleImageGeneration(handleImage HandleImageGeneration) *AgentBuilder {
	b.agent.handleImageGeneration = handleImage
	return b
}

// WithInspectEachRequest sets the user defined request inspector for more detail see InspectEachRequest
func (b *AgentBuilder) WithInspectEachRequest(inspect InspectEachRequest) *AgentBuilder {
	b.agent.inspectReq = inspect
	return b
}

// WithInspectEachResponse sets the user defined response inspector for more detail see InspectEachResponse
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

// Build last in chain during agent builder, returns agent or error, enforce all mandatory fields:
// todo: provide names of builders method to set mandatory fields
// return the copy, so build can be call multiple times and each time fresh agent will be returned
// todo: what about provider oand other fileds hide under interfaces is it ok or maybe builder should be oneshot and after return set to nil agent
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
	// DefaultMaxAgentSteps is the default maximum number of steps an agent can take before it is considered stuck.
	DefaultMaxAgentSteps = 5
)
