package rellm_test

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"rellm/pkg/rellm"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testBaseUrl              = "http://127.0.0.1"
	testPort                 = "1234"
	testResponsesApiEndpoint = "/v1/responses"
)

func TestAgent(t *testing.T) {
	agent, httpDo := buildTestAgent(t)
	defer httpDo.AssertExpectations(t)

	baseRequestMatch := func(req *http.Request) bool {

		if req.Body == nil {
			return false
		}

		b, err := io.ReadAll(req.Body)
		if err != nil {
			return false
		}
		req.Body = io.NopCloser(bytes.NewBuffer(b))

		//assert.Equal(t, givenApiReq, expectedApiReq)
		assert.JSONEq(t, string(rawResponsesApiReq_Hi), string(b))

		return req.URL.String() == testBaseUrl+":"+testPort+testResponsesApiEndpoint
	}

	// mock call of Do method of http client, do custom request matching for url and body field values
	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(rawResponsesApiResp_Hi)),
	}, nil).Once()

	respMsg, err := agent.Ask("Hi")
	assert.NoError(t, err, "failed to ask")
	assert.NotNil(t, respMsg, "response message should not be nil")
	//assert.JSONEq(t, rawResponsesApiResp_Hi, respMsg, "response message should match")
	assert.Equal(t, "Hello! How can I help you today? \n\nIf you have any questions about the weather, meteorology, climate patterns, or even how certain atmospheric phenomena work, feel free to ask!", respMsg, "response message should match")
}

func buildTestAgent(t *testing.T) (*rellm.Agent, *HttpDo) {

	workspace := inMemoryWorkspace(t)

	agentName := "TestAgent"

	mockHttp := new(HttpDo)

	lmsEndpoint := rellm.NewUniversalResponsesEndpoint(testBaseUrl, testPort, testResponsesApiEndpoint, nil)
	ep, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(lmsEndpoint).
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
		WithClientHttpDo(mockHttp).
		Build()
	assert.NoError(t, err, "failed to create endpoint")

	ta, err := rellm.NewAgentBuilder().
		WithEndpoint(ep).
		WithAgentName(agentName).
		WithWorkspaceDir(workspace).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithContinueConversation(false).
		WithSystemMessage("You are a helpful assistant with deep weather knowledge.").
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}

// inMemoryWorkspace create a directory for test in memory (in ram) return path
func inMemoryWorkspace(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "rellm-test-*")
	assert.NoError(t, err, "failed to create temp dir")
	return tempDir
}

type HttpDo struct {
	mock.Mock
}

func (h *HttpDo) Do(req *http.Request) (*http.Response, error) {
	args := h.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

const rawResponsesApiReq_Hi = `
{
    "model": "google/gemma-4-26b-a4b",
    "input": [
        {
            "role": "system",
            "content": "You are a helpful assistant with deep weather knowledge."
        },
        {
            "role": "user",
            "content": "Hi"
        }
    ]
}
`

const rawResponsesApiResp_Hi = `
{
    "id": "resp_ca4b31787e9d2c2ff33b9fba96c463c723764690750af951",
    "object": "response",
    "created_at": 1781873383,
    "model": "google/gemma-4-26b-a4b",
    "status": "completed",
    "completed_at": 1781873389,
    "output": [
        {
            "id": "rs_q92os6xjskh1oq74f888sfj",
            "type": "reasoning",
            "status": "completed",
            "summary": [],
            "content": [
                {
                    "type": "reasoning_text",
                    "text": "\n*   User says: \"Hi\"\n    *   Goal: Acknowledge the greeting and offer assistance, specifically leveraging my persona (helpful assistant with deep weather knowledge).\n\n    *   Helpful assistant.\n    *   Deep weather knowledge.\n\n    *   Greeting back (\"Hello!\", \"Hi there!\").\n    *   A prompt to engage with weather-related topics or any other questions.\n\n    *   \"Hi! How can I help you today? If you have any questions about the weather, meteorology, or climate, feel free to ask!\""
                }
            ]
        },
        {
            "id": "msg_64oviei8i6j5nubkgn8ztp",
            "type": "message",
            "role": "assistant",
            "status": "completed",
            "content": [
                {
                    "type": "output_text",
                    "text": "Hello! How can I help you today? \n\nIf you have any questions about the weather, meteorology, climate patterns, or even how certain atmospheric phenomena work, feel free to ask!",
                    "annotations": [],
                    "logprobs": []
                }
            ]
        }
    ],
    "error": null,
    "incomplete_details": null,
    "tools": [],
    "tool_choice": "auto",
    "parallel_tool_calls": true,
    "max_output_tokens": null,
    "temperature": 1,
    "top_p": 0.95,
    "presence_penalty": 0,
    "frequency_penalty": 1.1,
    "top_logprobs": 0,
    "max_tool_calls": null,
    "metadata": {},
    "background": false,
    "previous_response_id": null,
    "service_tier": "default",
    "truncation": "auto",
    "store": true,
    "instructions": null,
    "text": {
        "format": {
            "type": "text"
        }
    },
    "reasoning": {
        "effort": null,
        "summary": null
    },
    "safety_identifier": null,
    "prompt_cache_key": null,
    "user": null,
    "usage": {
        "input_tokens": 27,
        "input_tokens_details": {
            "cached_tokens": 22
        },
        "output_tokens": 157,
        "output_tokens_details": {
            "reasoning_tokens": 118
        },
        "total_tokens": 184,
        "cost": 0,
        "is_byok": false,
        "cost_details": {
            "upstream_inference_cost": 0,
            "upstream_inference_input_cost": 0,
            "upstream_inference_output_cost": 0
        }
    }
}
`
