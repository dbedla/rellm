package rellm

import "encoding/json"

// ResponsesAPIReq is the request body sent to the Responses API.
type ResponsesAPIReq struct {
	// Required
	Model string `json:"model"`

	// Input: raw conversation items in the provider's wire format
	Input []json.RawMessage `json:"input,omitempty"`

	// Output controls
	MaxOutputTokens  int      `json:"max_output_tokens,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	// Number of most likely tokens to return per output position
	TopLogprobs int `json:"top_logprobs,omitempty"`

	// Tools and function calling
	Tools []ToolDefinition `json:"tools,omitempty"`

	// Reasoning models configuration (e.g., o3-family)
	Reasoning *ReasoningConfig `json:"reasoning,omitempty"`

	// Structured output via the Responses API text.format field
	Text *TextConfig `json:"text,omitempty"`
}

// ToolDefinition defines a tool the model can call.
type ToolDefinition struct {
	Type        string      `json:"type"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
	Strict      bool        `json:"strict,omitempty"`
}

// ReasoningConfig controls behavior for reasoning-capable models.
type ReasoningConfig struct {
	Effort ReasoningEffort `json:"effort,omitempty"`
}

// TextConfig configures the text output of the Responses API.
type TextConfig struct {
	Format *TextFormat `json:"format,omitempty"`
}

// TextFormat requests structured output via text.format. Only the fields
// relevant to Type should be set: json_schema uses Name, Schema, and Strict.
type TextFormat struct {
	Type   string      `json:"type"`
	Name   string      `json:"name,omitempty"`
	Schema interface{} `json:"schema,omitempty"`
	Strict bool        `json:"strict,omitempty"`
}

type ResponsesAPIUsage struct {
	InputTokens        int `json:"input_tokens"`
	InputTokensDetails struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"input_tokens_details"`
	OutputTokens        int `json:"output_tokens"`
	OutputTokensDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"output_tokens_details"`
	TotalTokens int                  `json:"total_tokens"`
	Cost        *float64             `json:"cost,omitempty"`
	IsBYOK      *bool                `json:"is_byok,omitempty"`
	CostDetails *ResponseCostDetails `json:"cost_details,omitempty"`
}

type ResponseCostDetails struct {
	UpstreamInferenceCost       float64 `json:"upstream_inference_cost"`
	UpstreamInferenceInputCost  float64 `json:"upstream_inference_input_cost"`
	UpstreamInferenceOutputCost float64 `json:"upstream_inference_output_cost"`
}

// ResponseTool echoes a function tool definition sent in the request.
type ResponseTool struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
	Strict      bool   `json:"strict"`
}

// IncompleteDetails explains why a response is incomplete.
type IncompleteDetails struct {
	Reason string `json:"reason"`
}

// ResponsesAPIResp is the response body returned by the Responses API.
// It carries the fields mandatory per the OpenAI Responses API spec
// (id, object, created_at, model, instructions, output, error,
// incomplete_details, tools, tool_choice, parallel_tool_calls, temperature,
// top_p, metadata) plus the extras the library consumes. Unknown fields are
// ignored by encoding/json.
type ResponsesAPIResp struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"`
	CreatedAt         int                `json:"created_at"`
	Model             string             `json:"model"`
	Instructions      json.RawMessage    `json:"instructions"`
	Output            []json.RawMessage  `json:"output"`
	Error             *LLMError          `json:"error"`
	IncompleteDetails *IncompleteDetails `json:"incomplete_details"`
	Tools             []ResponseTool     `json:"tools"`
	ToolChoice        json.RawMessage    `json:"tool_choice"`
	ParallelToolCalls bool               `json:"parallel_tool_calls"`
	Temperature       *float64           `json:"temperature"`
	TopP              *float64           `json:"top_p"`
	Metadata          map[string]string  `json:"metadata"`
	Usage             ResponsesAPIUsage  `json:"usage"`
}

// LLMError is the error object carried in a failed response. On the
// Responses API, code is a string (e.g. "server_error"); code and message
// are required per the OpenAI spec whenever error is present.
type LLMError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
