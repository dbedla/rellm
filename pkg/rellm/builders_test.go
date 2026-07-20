package rellm_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"rellm/pkg/rellm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEndpointBuilder_Build(t *testing.T) {
	mockHttp := new(HttpDoMock)
	defer mockHttp.AssertExpectations(t)
	lmsEndpoint := rellm.NewUniversalResponsesEndpoint(testBaseUrl, testPort, testResponsesApiEndpoint, nil)

	t.Run("Successful build", func(t *testing.T) {
		builder := rellm.NewEndpointBuilder().
			WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
			WithResponsesApiEndpoint(lmsEndpoint).
			WithClientHttpDo(mockHttp)

		endpoint, err := builder.Build()
		assert.NoError(t, err)
		assert.NotNil(t, endpoint)
	})

	t.Run("Missing model", func(t *testing.T) {
		builder := rellm.NewEndpointBuilder().
			WithResponsesApiEndpoint(lmsEndpoint).
			WithClientHttpDo(mockHttp)

		agent, err := builder.Build()
		assert.Nil(t, agent)
		assert.Error(t, err)
		assert.Equal(t, rellm.ErrBuildNoModelName, err)
	})

	t.Run("Missing response api endpoint", func(t *testing.T) {
		builder := rellm.NewEndpointBuilder().
			WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
			WithClientHttpDo(mockHttp)

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildNoResponsesApiEndpoint)
	})

	t.Run("Missing http client", func(t *testing.T) {
		builder := rellm.NewEndpointBuilder().
			WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
			WithResponsesApiEndpoint(lmsEndpoint)

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildNoHttpClient)
	})
}

func TestPromptBuilder_AllFields(t *testing.T) {
	agent, httpDo := buildTestAgent(t)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			b, err := io.ReadAll(args.Get(0).(*http.Request).Body)
			assert.NoError(t, err)

			var req rellm.ResponsesApiReq
			err = json.Unmarshal(b, &req)
			assert.NoError(t, err)

			assert.Equal(t, float32(0.7), req.Temperature)
			assert.Equal(t, "high", req.Reasoning.Effort)
			assert.Equal(t, 512, req.MaxOutputTokens)
			assert.InDelta(t, 0.9, req.TopP, 0.001)
			assert.InDelta(t, 1.0, req.PresencePenalty, 0.001)
			assert.InDelta(t, 0.5, req.FrequencyPenalty, 0.001)
			require.NotNil(t, req.Seed)
			assert.Equal(t, int64(42), *req.Seed)
			assert.True(t, req.Logprobs)
			assert.Equal(t, 3, req.TopLogprobs)
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenRespHi)),
		}, nil)

	prompt, err := rellm.NewPromptBuilder().
		WithMessage("hi").
		WithTemperature(0.7).
		WithReasoning("high").
		WithMaxOutputTokens(512).
		WithTopP(0.9).
		WithPresencePenalty(1.0).
		WithFrequencyPenalty(0.5).
		WithSeed(42).
		WithLogprobs(true).
		WithTopLogprobs(3).
		Build()
	assert.NoError(t, err)

	_, err = agent.Execute(prompt)
	assert.NoError(t, err)
}

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
	mockHttp := new(HttpDoMock)
	defer mockHttp.AssertExpectations(t)
	lmsEndpoint := rellm.NewUniversalResponsesEndpoint(testBaseUrl, testPort, testResponsesApiEndpoint, nil)
	// Setup a valid endpoint for AgentBuilder tests
	endpoint, err := rellm.NewEndpointBuilder().
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
		WithResponsesApiEndpoint(lmsEndpoint).
		WithClientHttpDo(mockHttp).
		Build()
	assert.NoError(t, err)

	tmpDir := t.TempDir()

	t.Run("Successful build", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithEndpoint(endpoint).
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
			WithEndpoint(endpoint).
			WithStdoutLogger()

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildNoWorkspaceDir)
	})

	t.Run("Missing endpoint", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithWorkspaceDir(tmpDir).
			WithStdoutLogger().
			WithAgentName("TestAgent")

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildNoEndpoint)
	})

	t.Run("Missing agent name", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithEndpoint(endpoint).
			WithWorkspaceDir(tmpDir).
			WithStdoutLogger()

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildNoAgentName)
	})

	t.Run("Multiple loggers configured", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithAgentName("TestAgent").
			WithEndpoint(endpoint).
			WithWorkspaceDir(tmpDir).
			WithStdoutLogger().
			WithNoOpLogger()

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildExactlyOneLogger)
	})

	t.Run("No loggers configured", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithEndpoint(endpoint).
			WithWorkspaceDir(tmpDir).
			WithAgentName("TestAgent")

		_, err := builder.Build()
		assert.Error(t, err)
		assert.Equal(t, err, rellm.ErrBuildExactlyOneLogger)
	})
}
