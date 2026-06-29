package rellm_test

import (
	"testing"

	"rellm/pkg/rellm"

	"github.com/stretchr/testify/assert"
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
