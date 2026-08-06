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

func TestPromptBuilder_Validation_AccumulatesMultipleErrors(t *testing.T) {
	_, err := rellm.NewPromptBuilder().
		WithMessage("").
		WithReasoning("").
		Build()

	assert.Error(t, err)
	assert.True(t, errors.Is(err, rellm.ErrEmptyPrompt), "ErrEmptyPrompt should be present in joined error")
	assert.True(t, errors.Is(err, rellm.ErrEmptyReasoningEffort), "ErrEmptyReasoningEffort should be present in joined error")

}

func TestPromptBuilder_Build_EmptyMessage(t *testing.T) {
	_, err := rellm.NewPromptBuilder().Build()
	assert.Error(t, err)
	assert.ErrorIs(t, err, rellm.ErrEmptyPrompt)
}

func TestAgentBuilder_Build(t *testing.T) {
	tmpDir := t.TempDir()

	buildProvider := func() rellm.Provider {
		p, err := rellm.NewLMStudioProvider(rellm.Model_LMS_Google_Gemma_4_26B_A4B, testBaseUrl, testPort)
		assert.NoError(t, err)
		return p
	}

	t.Run("Successful build", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithProvider(buildProvider()).
			WithWorkspaceDir(tmpDir).
			WithAgentName("TestAgent").
			WithStdoutLogger()

		agent, err := builder.Build()
		assert.NoError(t, err)
		assert.NotNil(t, agent)
	})

	t.Run("Missing workspaceDir", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithAgentName("TestAgent").
			WithNoOpLogger().
			WithProvider(buildProvider()).
			WithStdoutLogger()

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildNoWorkspaceDir)
	})

	t.Run("Missing provider", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithWorkspaceDir(tmpDir).
			WithStdoutLogger().
			WithAgentName("TestAgent")

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildNoProvider)
	})

	t.Run("Missing agent name", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithProvider(buildProvider()).
			WithWorkspaceDir(tmpDir).
			WithStdoutLogger()

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildNoAgentName)
	})

	t.Run("Multiple loggers configured", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithAgentName("TestAgent").
			WithProvider(buildProvider()).
			WithWorkspaceDir(tmpDir).
			WithStdoutLogger().
			WithNoOpLogger()

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildExactlyOneLogger)
	})

	t.Run("No loggers configured", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithProvider(buildProvider()).
			WithWorkspaceDir(tmpDir).
			WithAgentName("TestAgent")

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildExactlyOneLogger)
	})
}
