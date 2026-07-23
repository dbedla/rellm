package rellm_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"rellm/pkg/rellm"
	"strings"
	"testing"

	_ "embed"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//go:embed testdata/image_req.json
var goldenImageReq string

//go:embed testdata/image_resp.json
var goldenImageResp string

func TestAgentPromptToGetImage(t *testing.T) {
	agent, httpDo := buildTestImageAgent(t, TestDefaultMaxToolsIterationWithoutReturnMessage, testImageHandler)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenImageReq, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenImageResp)),
		}, nil)

	q := "A clean, minimalist flat vector illustration of a tic-tac-toe board. White background, bold black grid lines. Three bright blue \"O\" symbols are aligned horizontally in the middle row, indicating a win. Minimalist aesthetic, high contrast, simple and modern graphic design."

	respMsg, err := agent.Ask(q)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "[system <for user visibility only>] image generated, handler returned: image-stored-under-this-id", respMsg, "response message should match")

	conversation := agent.CurrentConversation()
	lastMsg := conversation[len(conversation)-1]
	var imageConversationRepresentation rellm.ImageGenerationConversationPlaceholder
	err = json.Unmarshal(lastMsg, &imageConversationRepresentation)
	assert.NoError(t, err, "failed to unmarshal last message")

	assert.Equal(t, "image-stored-under-this-id", imageConversationRepresentation.Result)
	assert.Equal(t, "image_generation_call", imageConversationRepresentation.Type)
	assert.Equal(t, "completed", imageConversationRepresentation.Status)
	assert.Equal(t, "ig_tmp_vqotwoa5eg", imageConversationRepresentation.Id)
}

func TestAgentPromptToGetImage_handlerErr(t *testing.T) {
	agent, httpDo := buildTestImageAgent(t, TestDefaultMaxToolsIterationWithoutReturnMessage, testImageHandlerAlwaysErr)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenImageReq, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenImageResp)),
		}, nil)

	q := "A clean, minimalist flat vector illustration of a tic-tac-toe board. White background, bold black grid lines. Three bright blue \"O\" symbols are aligned horizontally in the middle row, indicating a win. Minimalist aesthetic, high contrast, simple and modern graphic design."

	respMsg, err := agent.Ask(q)
	assert.Error(t, err)
	assert.ErrorIs(t, err, rellm.ErrCustomImageHandlerFailed)

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "", respMsg, "response message should match")
}

func testImageHandler(image rellm.OutputItem) (string, error) {
	return "image-stored-under-this-id", nil
}

func testImageHandlerAlwaysErr(image rellm.OutputItem) (string, error) {
	return "", errors.New("test err in image handling error")
}

func buildTestImageAgent(t *testing.T, maxToolsIterationWithoutReturnMessage uint64, imageH rellm.HandleImage) (*rellm.Agent, *HttpDoMock) {

	agentName := "TestImageAgent"
	workspace := t.TempDir()
	mockHttp := new(HttpDoMock)

	lmsEndpoint := rellm.NewUniversalResponsesEndpoint(testBaseUrl, testPort, testResponsesApiEndpoint, nil)
	ep, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(lmsEndpoint).
		WithProvider(rellm.Provider_OpenRouter).
		WithModel(rellm.Model("x-ai/grok-imagine-image-quality")).
		WithClientHttpDo(mockHttp).
		Build()
	assert.NoError(t, err, "failed to create endpoint")

	ta, err := rellm.NewAgentBuilder().
		WithEndpoint(ep).
		WithAgentName(agentName).
		WithWorkspaceDir(workspace).
		WithMaxToolsIterationWithoutReturnMessage(maxToolsIterationWithoutReturnMessage).
		WithContinueConversation(false).
		WithSystemMessage("You are a helpful assistant.").
		WithHandleImage(imageH).
		WithNoOpLogger().
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}
