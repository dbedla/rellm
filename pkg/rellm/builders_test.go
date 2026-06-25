package rellm_test

import (
	"testing"

	"rellm/pkg/rellm"

	"github.com/stretchr/testify/assert"
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
		require.NoError(t, err)
		assert.NotNil(t, endpoint)
	})

	t.Run("Missing model", func(t *testing.T) {
		builder := rellm.NewEndpointBuilder().
			WithResponsesApiEndpoint(lmsEndpoint).
			WithClientHttpDo(mockHttp)

		_, err := builder.Build()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing model")
	})

	t.Run("Missing response api endpoint", func(t *testing.T) {
		builder := rellm.NewEndpointBuilder().
			WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
			WithClientHttpDo(mockHttp)

		_, err := builder.Build()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing response api endpoint")
	})
}

func TestAgentBuilder_Build(t *testing.T) {
	mockHttp := new(HttpDoMock)
	defer mockHttp.AssertExpectations(t)
	lmsEndpoint := rellm.NewUniversalResponsesEndpoint(testBaseUrl, testPort, testResponsesApiEndpoint, nil)
	// Setup a valid endpoint for AgentBuilder tests
	endpoint, _ := rellm.NewEndpointBuilder().
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
		WithResponsesApiEndpoint(lmsEndpoint).
		Build()

	tmpDir := t.TempDir()

	t.Run("Successful build", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithEndpoint(endpoint).
			WithWorkspaceDir(tmpDir).
			WithAgentName("TestAgent").
			WithStdoutLogger()

		agent, err := builder.Build()
		require.NoError(t, err)
		assert.NotNil(t, agent)
	})

	t.Run("Missing workspaceDir", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithEndpoint(endpoint).
			WithStdoutLogger()

		_, err := builder.Build()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing workspaceDir")
	})

	t.Run("Missing endpoint", func(t *testing.T) {
		builder := rellm.NewAgentBuilder().
			WithWorkspaceDir(tmpDir).
			WithStdoutLogger()

		_, err := builder.Build()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing endpoint")
	})
}
