package rellm

import "context"

// AgentBuilder constructs an Agent. It is the recommended way to create an
// agent; Build validates mandatory fields before returning it.
//
//	agent, err := rellm.NewAgentBuilder().
//		WithProvider(provider).
//		WithConversation(rellm.NewInMemoryConversation()).
//		WithUnknownConversationElementKeepInTheLoop(). // other policies: WithUnknownConversationElementDrop, WithUnknownConversationElementHandler
//		WithImageGenerationKeepInTheLoop().           // other policies: WithImageGenerationDrop, WithImageGenerationHandler
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

// WithTextFormat requests structured output via the Responses API text.format
// field for every request this agent makes, in the same way the toolset
// applies to every request. Optional. Calling it again replaces the previous
// format.
// See https://developers.openai.com/api/docs/guides/structured-outputs?api-mode=responses
func (b *AgentBuilder) WithTextFormat(f TextFormat) *AgentBuilder {
	b.agent.textFormat = &f
	return b
}

// WithConversation sets the conversation used to persist and load the
// conversation history. Required.
func (b *AgentBuilder) WithConversation(storage Conversation) *AgentBuilder {
	b.agent.conversation = storage
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

// WithImageGenerationDrop configures the agent to drop generated images from
// the conversation loop. The original image is still returned in
// Report.Images[n].Original.
func (b *AgentBuilder) WithImageGenerationDrop() *AgentBuilder {
	b.agent.handleImageGeneration = func(ctx context.Context, image *ImageGeneration) ([]ConversationElement, error) {
		return nil, nil
	}
	return b
}

// WithImageGenerationKeepInTheLoop configures the agent to keep generated
// images unchanged in the conversation loop.
func (b *AgentBuilder) WithImageGenerationKeepInTheLoop() *AgentBuilder {
	b.agent.handleImageGeneration = func(ctx context.Context, image *ImageGeneration) ([]ConversationElement, error) {
		return []ConversationElement{image}, nil
	}
	return b
}

// WithImageGenerationHandler sets a custom image policy. The returned slice
// replaces the generated image's slot in the conversation; nil or an empty
// slice drops it.
func (b *AgentBuilder) WithImageGenerationHandler(handler HandleImageGeneration) *AgentBuilder {
	b.agent.handleImageGeneration = handler
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
// fields are set via WithProvider and WithConversation; optional fields
// fall back to defaults. Build returns a copy, so the builder can be reused to
// create another agent.
func (b *AgentBuilder) Build() (*Agent, error) {
	if b.agent.textFormat != nil {
		err := validateTextFormat(b.agent.textFormat)
		if err != nil {
			return nil, err
		}
	}

	if b.agent.provider == nil {
		return nil, ErrBuildNoProvider
	}

	if b.agent.maxAgentSteps == 0 {
		b.agent.maxAgentSteps = DefaultMaxAgentSteps
	}

	if b.agent.conversation == nil {
		return nil, ErrBuildNoConversation
	}

	if b.agent.handleUnknownConversationElement == nil {
		return nil, ErrNoUnknownConversationElementHandler
	}

	if b.agent.handleImageGeneration == nil {
		return nil, ErrNoImageHandler
	}

	agent := b.agent // copy: builder stays reusable without mutating built agents
	return &agent, nil
}

const (
	// DefaultMaxAgentSteps is the default maximum number of steps an agent may
	// take per prompt before reporting ErrMaxAgentStepsReached.
	DefaultMaxAgentSteps = 5
)

// validate checks the fields relevant to the text format type.
func validateTextFormat(f *TextFormat) error {
	switch f.Type {
	case "text", "json_object":
		return nil
	case "json_schema":
		if f.Name == "" {
			return ErrEmptyTextFormatName
		}
		if f.Schema == nil {
			return ErrEmptyTextFormatSchema
		}
		return nil
	case "":
		return ErrEmptyTextFormat
	default:
		return ErrUnknownTextFormatType
	}
}
