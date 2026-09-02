package rellm_test

import (
	"errors"
	"testing"

	"rellm/pkg/rellm"

	"github.com/stretchr/testify/assert"
)

func TestPromptBuilder_Validation_RejectsEmptyReasoning(t *testing.T) {
	_, err := rellm.NewPromptBuilder().
		WithMessage("Hi").
		WithReasoning("").
		Build()

	assert.Error(t, err)
	assert.True(t, errors.Is(err, rellm.ErrEmptyReasoningEffort), "ErrEmptyReasoningEffort should be present in joined error")
}

func TestPromptBuilder_Validation_RejectsEmptyMessage(t *testing.T) {
	_, err := rellm.NewPromptBuilder().
		WithMessage("   ").
		Build()

	assert.Error(t, err)
	assert.ErrorIs(t, err, rellm.ErrEmptyPrompt)
}

func TestPromptBuilder_Build_EmptyMessage(t *testing.T) {
	_, err := rellm.NewPromptBuilder().Build()
	assert.Error(t, err)
	assert.ErrorIs(t, err, rellm.ErrEmptyPrompt)
}

func TestAgentBuilder_Build(t *testing.T) {

	buildProvider := func() rellm.Provider {
		p, err := rellm.NewLMStudioProvider("google/gemma-4-26b-a4b", testBaseUrl, testPort)
		assert.NoError(t, err)
		return p
	}

	t.Run("Successful build", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithProvider(buildProvider()).
			WithAgentName("TestAgent").
			WithConversationStorage(rellm.NewInMemoryStorage())

		agent, err := builder.Build()
		assert.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "TestAgent", agent.Name())
	})

	t.Run("Missing conversation storage", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithProvider(buildProvider()).
			WithAgentName("TestAgent")

		_, err := builder.Build()
		assert.ErrorIs(t, err, rellm.ErrBuildNoConversationStorage)
	})

	t.Run("Missing provider", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithAgentName("TestAgent")

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildNoProvider)
	})

	t.Run("Build twice does not alias the builder state", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithProvider(buildProvider()).
			WithAgentName("FirstAgent").
			WithConversationStorage(rellm.NewInMemoryStorage())

		agent1, err := builder.Build()
		assert.NoError(t, err)

		agent2, err := builder.WithAgentName("SecondAgent").Build()
		assert.NoError(t, err)

		assert.NotSame(t, agent1, agent2, "Build must return independent agents")
	})

}
