package rellm_test

import (
	"bytes"
	"context"
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
	agent, httpDo := buildTestImageAgent(t, testDefaultMaxAgentSteps, testImageGenerationHandler)
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

	ctx := context.Background()
	respMsg, err := agent.Ask(ctx, q)
	assert.NoError(t, err)

	assert.NotNil(t, respMsg, "response message should not be nil")

	conversation, err := agent.CurrentConversation(ctx)
	assert.NoError(t, err)
	lastMsg := conversation[len(conversation)-1]
	imageGeneration, ok := lastMsg.(*rellm.ImageGeneration)
	assert.True(t, ok, "expected *ImageGeneration, got %T", lastMsg)

	assert.Equal(t, "image-stored-under-this-id", imageGeneration.Result)
	assert.Equal(t, rellm.KindImageGeneration, imageGeneration.Kind())
	assert.Equal(t, "completed", imageGeneration.Status)
	assert.Equal(t, "ig_tmp_vqotwoa5eg", imageGeneration.ID)
}

func TestAgentPromptToGetImage_handlerErr(t *testing.T) {
	agent, httpDo := buildTestImageAgent(t, testDefaultMaxAgentSteps, testImageGenerationHandlerAlwaysErr)
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

	ctx := context.Background()
	respMsg, err := agent.Ask(ctx, q)
	assert.Error(t, err)
	assert.ErrorIs(t, err, rellm.ErrCustomImageHandlerFailed)

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "", respMsg, "response message should match")
}

func testImageGenerationHandler(_ context.Context, image *rellm.ImageGeneration) (string, error) {
	return "image-stored-under-this-id", nil
}

func testImageGenerationHandlerAlwaysErr(_ context.Context, image *rellm.ImageGeneration) (string, error) {
	return "", errors.New("test err in image handling error")
}

func buildTestImageAgent(t *testing.T, maxAgentSteps uint64,
	imageGenerationH rellm.HandleImageGeneration) (*rellm.Agent, *HTTPDoMock) {

	agentName := "TestImageAgent"
	mockHttp := new(HTTPDoMock)

	p, err := rellm.NewOpenRouterProviderWithHTTPClient("test-key", rellm.Model("x-ai/grok-imagine-image-quality"), mockHttp)
	assert.NoError(t, err, "failed to create provider")

	ta, err := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(maxAgentSteps).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage("You are a helpful assistant.").
		WithHandleImageGeneration(imageGenerationH).
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}
