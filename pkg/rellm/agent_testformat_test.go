package rellm_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"rellm/pkg/rellm"
	"strings"
	"testing"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//go:embed testdata/textformat/lms_tf_gemma_req.json
var goldenTextFormatLMSGemmaReq string

//go:embed testdata/textformat/lms_tf_gemma_resp.json
var goldenTextFormatLMSGemmaResp string

//go:embed testdata/textformat/or_tf_gemini_req.json
var goldenTextFormatOrGeminiReq string

//go:embed testdata/textformat/or_tf_gemini_resp.json
var goldenTextFormatOrGeminiResp string

//go:embed testdata/textformat/or_tf_luna_req.json
var goldenTextFormatOrLunaReq string

//go:embed testdata/textformat/or_tf_luna_resp.json
var goldenTextFormatOrLunaResp string

func TestAgentTestFormat(t *testing.T) {

	lmsAgent, lmsHTTPDo := buildTestFormatOutputLMSAgent(t, testDefaultMaxAgentSteps)
	orGeminiAgent, orGeminiHTTPDo := buildTestFormatOpenRouterAgent(
		t,
		rellm.Model("google/gemini-3.1-flash-lite"),
		testDefaultMaxAgentSteps)
	orLunaAgent, orLunaHTTPDo := buildTestFormatOpenRouterAgent(
		t,
		rellm.Model("openai/gpt-5.6-luna"),
		testDefaultMaxAgentSteps)

	tests := []struct {
		tname    string
		agent    *rellm.Agent
		httpMock *HTTPDoMock
		req      string
		resp     string
	}{
		{
			tname:    "TestFormatOutputLMSAgent",
			agent:    lmsAgent,
			httpMock: lmsHTTPDo,
			req:      goldenTextFormatLMSGemmaReq,
			resp:     goldenTextFormatLMSGemmaResp,
		},
		{
			tname:    "TestFormatOutputOrGeminiAgent",
			agent:    orGeminiAgent,
			httpMock: orGeminiHTTPDo,
			req:      goldenTextFormatOrGeminiReq,
			resp:     goldenTextFormatOrGeminiResp,
		},
		{
			tname:    "TestFormatOutputOrLunaAgent",
			agent:    orLunaAgent,
			httpMock: orLunaHTTPDo,
			req:      goldenTextFormatOrLunaReq,
			resp:     goldenTextFormatOrLunaResp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.tname, func(t *testing.T) {
			tt.httpMock.On("Do", mock.MatchedBy(baseRequestMatch)).
				Once().
				Run(func(args mock.Arguments) {
					req := args.Get(0).(*http.Request)

					b, err := io.ReadAll(req.Body)
					assert.NoError(t, err, "failed to read request body")

					req.Body = io.NopCloser(bytes.NewBuffer(b))

					assert.JSONEq(t, tt.req, string(b))
				}).
				Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(tt.resp)),
				}, nil)

			textFormat := &rellm.TextFormat{
				Type:   "json_schema",
				Name:   "person",
				Strict: true,
				Schema: (&jsonschema.Reflector{DoNotReference: true}).Reflect(&Person{}),
			}

			msg := "I am John Snow from Winterfel, I have 100 years..."
			prompt, err := rellm.NewPromptBuilder().
				WithMessage(msg).
				WithTextFormat(textFormat).
				Build()
			assert.NoError(t, err)

			resp, err := tt.agent.Execute(context.Background(), prompt)
			assert.NoError(t, err)

			p := Person{}
			err = json.Unmarshal([]byte(resp), &p)
			assert.NoError(t, err)

			assert.Equal(t, "John Snow", p.Name)
			assert.Equal(t, 100, p.Age)
			assert.Equal(t, "Winterfel", p.City)
		})
	}

}

type Person struct {
	Name string `json:"name" jsonschema:"description=Full name of the person"`
	Age  int    `json:"age"  jsonschema:"description=Age in years"`
	City string `json:"city" jsonschema:"description=City of residence"`
}

const structuredOutputSysPrompt = `You are an assistant that extracts structured information from user text and returns only valid JSON matching the requested schema.`

func buildTestFormatOpenRouterAgent(t *testing.T, model rellm.Model, maxAgentSteps uint64) (*rellm.Agent, *HTTPDoMock) {
	agentName := "TestFormatOpenRouterAgent"
	mockHttp := new(HTTPDoMock)

	p, err := rellm.NewOpenRouterProviderWithHTTPClient("test-key", model, mockHttp)
	assert.NoError(t, err)

	ta, err := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(maxAgentSteps).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage(structuredOutputSysPrompt).
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}

func buildTestFormatOutputLMSAgent(t *testing.T, maxAgentSteps uint64) (*rellm.Agent, *HTTPDoMock) {

	agentName := "FormatOutputLMSAgent"
	mockHttp := new(HTTPDoMock)

	p, err := rellm.NewLMStudioProviderWithHTTPClient("google/gemma-4-26b-a4b", testBaseUrl, testPort, mockHttp)
	assert.NoError(t, err, "failed to create provider")

	ta, err := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(maxAgentSteps).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage(structuredOutputSysPrompt).
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}
