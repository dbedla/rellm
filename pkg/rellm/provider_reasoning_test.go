package rellm

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReasoningSummaryAndEncryptedContentRoundTrip(t *testing.T) {
	providers := []struct {
		name     string
		provider Provider
	}{
		{name: "OpenRouter", provider: &OpenRouterProvider{}},
		{name: "OpenAI", provider: &OpenAIProvider{}},
		{name: "LMStudio", provider: &LMStudioProvider{}},
	}
	cases := []struct {
		name   string
		raw    string
		status string
	}{
		{
			name:   "structured summary",
			status: "completed",
			raw:    `{"id":"rs-1","type":"reasoning","status":"completed","summary":[{"type":"summary_text","text":"List the files."},{"type":"summary_text","text":"Summarize their contents."}]}`,
		},
		{
			name:   "summary and encrypted content",
			status: "completed",
			raw:    `{"id":"rs-1","type":"reasoning","status":"completed","summary":[{"type":"summary_text","text":"List the files."}],"encrypted_content":"opaque-replay-state","format":"openai-responses-v1"}`,
		},
		{
			name:   "encrypted content only",
			status: "completed",
			raw:    `{"id":"rs-1","type":"reasoning","status":"completed","summary":[],"encrypted_content":"opaque-replay-state","format":"openai-responses-v1"}`,
		},
		{
			// OpenAI omits the status field entirely on reasoning items.
			name: "no status",
			raw:  `{"id":"rs-1","type":"reasoning","summary":[{"type":"summary_text","text":"List the files."}],"encrypted_content":"opaque-replay-state","format":"openai-responses-v1"}`,
		},
	}
	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					elements, err := p.provider.ToConversationElements([]json.RawMessage{json.RawMessage(tc.raw)})
					require.NoError(t, err)
					require.Len(t, elements, 1)
					require.IsType(t, &Reasoning{}, elements[0])
					reasoning, ok := elements[0].(*Reasoning)
					assert.True(t, ok)
					assert.Equal(t, tc.status, reasoning.Status)

					wire, err := p.provider.ToProviderRepresentation(elements)
					require.NoError(t, err)
					require.Len(t, wire, 1)
					assert.JSONEq(t, tc.raw, string(wire[0]))
				})
			}
		})
	}
}

//go:embed testdata/openrouter/file_summary_02_luna_resp.json
var goldenReasoningFileSummaryResponse []byte

func TestOpenRouterReasoningRoundTripFromFileSummary(t *testing.T) {
	var response struct {
		Output []json.RawMessage `json:"output"`
	}
	require.NoError(t, json.Unmarshal(goldenReasoningFileSummaryResponse, &response))
	require.NotEmpty(t, response.Output)

	p := &OpenRouterProvider{}
	elements, err := p.ToConversationElements(response.Output[:1])
	require.NoError(t, err)
	require.Len(t, elements, 1)
	require.IsType(t, &Reasoning{}, elements[0])
	reasoning := elements[0].(*Reasoning)
	require.Len(t, reasoning.Summary, 1)
	assert.Equal(t, "summary_text", reasoning.Summary[0].Type)
	assert.NotEmpty(t, reasoning.Summary[0].Text)
	assert.NotEmpty(t, reasoning.EncryptedContent)
	assert.Equal(t, "openai-responses-v1", reasoning.Format)
	assert.Empty(t, reasoning.Signature)

	wire, err := p.ToProviderRepresentation(elements)
	require.NoError(t, err)
	require.Len(t, wire, 1)
	assert.JSONEq(t, string(response.Output[0]), string(wire[0]))
}
