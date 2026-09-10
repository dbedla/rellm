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
	finalReport, err := agent.Ask(ctx, q)
	assert.NoError(t, err)
	assert.Nil(t, finalReport.Messages)
	if assert.NotNil(t, finalReport.Image) {
		assert.Equal(t, "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAA", finalReport.Image.Original.Result)
		assert.Equal(t, "completed", finalReport.Image.Original.Status)
		assert.Equal(t, "ig_tmp_vqotwoa5eg", finalReport.Image.Original.ID)
		if assert.Len(t, finalReport.Image.PolicyOutput, 1) {
			policyImage, ok := finalReport.Image.PolicyOutput[0].(*rellm.ImageGeneration)
			if assert.True(t, ok) {
				assert.Equal(t, "image-stored-under-this-id", policyImage.Result)
			}
		}
	}

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
	finalReport, err := agent.Ask(ctx, q)
	assert.Error(t, err)
	assert.ErrorIs(t, err, rellm.ErrCustomImageHandlerFailed)

	assert.Nil(t, finalReport.Messages)
	assert.NotNil(t, finalReport.Image)
	assert.Equal(t, "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAA", finalReport.Image.Original.Result)
	assert.Empty(t, finalReport.Image.PolicyOutput)
}

func TestAgentPromptToGetImagePolicies(t *testing.T) {
	tests := []struct {
		name             string
		configure        func(*rellm.AgentBuilder)
		conversationSize int
		policyOutputSize int
	}{
		{
			name: "drop",
			configure: func(builder *rellm.AgentBuilder) {
				builder.WithImageGenerationDrop()
			},
			conversationSize: 2,
			policyOutputSize: 0,
		},
		{
			name: "keep in the loop",
			configure: func(builder *rellm.AgentBuilder) {
				builder.WithImageGenerationKeepInTheLoop()
			},
			conversationSize: 3,
			policyOutputSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, httpDo := buildTestImageAgentWithPolicy(t, testDefaultMaxAgentSteps, tt.configure)
			defer httpDo.AssertExpectations(t)

			httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
				Once().
				Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(goldenImageResp)),
				}, nil)

			finalReport, err := agent.Ask(context.Background(), "generate an image")
			assert.NoError(t, err)
			assert.Nil(t, finalReport.Messages)
			if assert.NotNil(t, finalReport.Image) {
				assert.Equal(t, "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAA", finalReport.Image.Original.Result)
				assert.Len(t, finalReport.Image.PolicyOutput, tt.policyOutputSize)
			}

			conversation, err := agent.CurrentConversation(context.Background())
			assert.NoError(t, err)
			assert.Len(t, conversation, tt.conversationSize)
		})
	}
}

func testImageGenerationHandler(_ context.Context, image *rellm.ImageGeneration) ([]rellm.ConversationElement, error) {
	image.Result = "image-stored-under-this-id"
	return []rellm.ConversationElement{image}, nil
}

func testImageGenerationHandlerAlwaysErr(_ context.Context, image *rellm.ImageGeneration) ([]rellm.ConversationElement, error) {
	return nil, errors.New("test err in image handling error")
}

func buildTestImageAgent(t *testing.T, maxAgentSteps uint64,
	imageGenerationH rellm.HandleImageGeneration) (*rellm.Agent, *HTTPDoMock) {
	return buildTestImageAgentWithPolicy(t, maxAgentSteps, func(builder *rellm.AgentBuilder) {
		builder.WithImageGenerationHandler(imageGenerationH)
	})
}

func buildTestImageAgentWithPolicy(t *testing.T, maxAgentSteps uint64,
	configure func(*rellm.AgentBuilder)) (*rellm.Agent, *HTTPDoMock) {

	agentName := "TestImageAgent"
	mockHttp := new(HTTPDoMock)

	p, err := rellm.NewOpenRouterProviderWithHTTPClient("test-key", rellm.Model("x-ai/grok-imagine-image-quality"), mockHttp)
	assert.NoError(t, err, "failed to create provider")

	builder := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(maxAgentSteps).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage("You are a helpful assistant.")
	configure(builder)
	ta, err := builder.Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}
