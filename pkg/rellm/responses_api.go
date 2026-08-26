package rellm

import "encoding/json"

type ResponsesApiReq struct {
	// Required
	Model string `json:"model"`

	// Input can be:
	// - string (simple text)
	// - []ContentPart (multimodal/text parts for the Responses API)
	// - json.RawMessage (if you want complete control)
	Input []json.RawMessage `json:"input,omitempty"`

	// Optional high-level instruction (system-style prompt)
	Instructions string `json:"instructions,omitempty"`

	// Output controls
	MaxOutputTokens  int     `json:"max_output_tokens,omitempty"`
	Temperature      float32 `json:"temperature,omitempty"`
	TopP             float32 `json:"top_p,omitempty"`
	PresencePenalty  float32 `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`

	// Stop controls
	// NOTE: Some examples show "stop"; others may reference "stop_sequences".
	// Use Stop for modern Responses API.
	Stop []string `json:"stop,omitempty"`

	// Log probabilities (if supported by the selected model)
	Logprobs    bool `json:"logprobs,omitempty"`
	TopLogprobs int  `json:"top_logprobs,omitempty"`

	// Determinism
	Seed *int64 `json:"seed,omitempty"`

	// Tools and function calling
	Tools []Tool `json:"tools,omitempty"`
	// ToolChoice can be:
	// - string: "auto", "none", or "required"
	// - ToolChoiceOption: { "type": "function", "function": { "name": "..." } }
	ToolChoice interface{} `json:"tool_choice,omitempty"`

	// Reasoning models configuration (e.g., o3-family)
	Reasoning *ReasoningConfig `json:"reasoning,omitempty"`

	// Structured output and JSON schema
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// Streaming (SSE) toggle
	Stream bool `json:"stream,omitempty"`

	// Modalities and audio output (text-to-speech, etc.)
	// Example: Modalities: ["text"] or ["text","audio"]
	Modalities []string           `json:"modalities,omitempty"`
	Audio      *AudioOutputConfig `json:"audio,omitempty"`

	// File attachments for tools like file_search / code_interpreter
	Attachments []Attachment `json:"attachments,omitempty"`

	// Arbitrary metadata you want to associate with the request
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ContentPart represents a single item in the "input" array for multimodal or
// structured inputs. Use Type: "input_text", "input_image", "input_audio", etc.
// Only fill the fields relevant to the chosen Type; the rest remain empty.
type ContentPart struct {
	Type string `json:"type"`

	// For type == "input_text"
	Text string `json:"text,omitempty"`

	// For type == "input_image"
	// Depending on the API variant, you might provide a URL or embedded data.
	// Keep both optional to support either style.
	ImageURL string     `json:"image_url,omitempty"`
	Image    *ImageData `json:"image,omitempty"`

	// For type == "input_audio"
	Audio *InputAudioData `json:"audio,omitempty"`
}

// ImageData holds embedded image bytes for input_image variants.
type ImageData struct {
	// Base64-encoded image bytes
	Data string `json:"data"`
	// Example: "png", "jpeg", "webp"
	Format string `json:"format"`
}

// InputAudioData holds embedded audio for input_audio variants.
type InputAudioData struct {
	// Base64-encoded audio bytes
	Data string `json:"data"`
	// Example: "wav", "mp3", "pcm16"
	Format string `json:"format"`
}

// Tool defines a tool the model can call.
type Tool struct {
	Type        string      `json:"type"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
	Strict      bool        `json:"strict,omitempty"`
}

// ToolChoiceOption allows forcing a specific function.
type ToolChoiceOption struct {
	Type     string                    `json:"type"` // e.g., "function"
	Function *ToolChoiceFunctionTarget `json:"function,omitempty"`
}

type ToolChoiceFunctionTarget struct {
	Name string `json:"name"`
}

// ReasoningConfig controls behavior for reasoning-capable models.
type ReasoningConfig struct {
	Effort ReasoningEffort `json:"effort,omitempty"`
}

// ResponseFormat lets you request structured output.
type ResponseFormat struct {
	// "text" | "json_object" | "json_schema"
	Type       string                `json:"type"`
	JSONSchema *JSONSchemaDefinition `json:"json_schema,omitempty"`
}

// JSONSchemaDefinition is used when Type == "json_schema".
type JSONSchemaDefinition struct {
	// A friendly name for the schema
	Name string `json:"name"`
	// Enforce the exact schema if supported by the model
	Strict bool `json:"strict,omitempty"`
	// The actual JSON Schema (use a Go struct, map, or raw JSON)
	Schema interface{} `json:"schema"`
}

// AudioOutputConfig controls audio generation when you include "audio" in Modalities.
type AudioOutputConfig struct {
	// e.g., "alloy"
	Voice string `json:"voice,omitempty"`
	// e.g., "wav", "mp3", "pcm16"
	Format string `json:"format,omitempty"`
	// Optional sample rate (e.g., 24000)
	SampleRate int `json:"sample_rate,omitempty"`
}

// Attachment references files available to tools.
type Attachment struct {
	FileID string           `json:"file_id"`
	Tools  []AttachmentTool `json:"tools,omitempty"`
}

// AttachmentTool declares which tool(s) can access the attached file.
type AttachmentTool struct {
	// e.g., "file_search", "code_interpreter"
	Type string `json:"type"`
}

type ResponsesApiUsage struct {
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

type ResponseTool struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
	Strict      bool   `json:"strict"`
}

type ResponsesApiResp struct {
	ID                string            `json:"id"`
	Object            string            `json:"object"`
	CreatedAt         int               `json:"created_at"`
	Model             string            `json:"model"`
	Status            string            `json:"status"`
	CompletedAt       int               `json:"completed_at"`
	Output            []json.RawMessage `json:"output"`
	Error             *ErrorLLM         `json:"error"`
	IncompleteDetails interface{}       `json:"incomplete_details"`
	Tools             []ResponseTool    `json:"tools"`
	ToolChoice        string            `json:"tool_choice"`
	ParallelToolCalls bool              `json:"parallel_tool_calls"`
	MaxOutputTokens   interface{}       `json:"max_output_tokens"`
	Temperature       float64           `json:"temperature"`
	TopP              float64           `json:"top_p"`
	PresencePenalty   float64           `json:"presence_penalty"`
	FrequencyPenalty  float64           `json:"frequency_penalty"`
	TopLogprobs       float64           `json:"top_logprobs"`
	MaxToolCalls      interface{}       `json:"max_tool_calls"`
	Metadata          struct {
	} `json:"metadata"`
	Background         bool        `json:"background"`
	PreviousResponseID interface{} `json:"previous_response_id"`
	ServiceTier        string      `json:"service_tier"`
	Truncation         string      `json:"truncation"`
	Store              bool        `json:"store"`
	Instructions       interface{} `json:"instructions"`
	Text               struct {
		Format struct {
			Type string `json:"type"`
		} `json:"format"`
	} `json:"text"`
	Reasoning        interface{}       `json:"reasoning"`
	SafetyIdentifier interface{}       `json:"safety_identifier"`
	PromptCacheKey   interface{}       `json:"prompt_cache_key"`
	User             json.RawMessage   `json:"user,omitempty"`
	Usage            ResponsesApiUsage `json:"usage"`
}

type ErrorLLM struct {
	Message  string `json:"message"`
	Code     any    `json:"code"`
	Metadata struct {
		ProviderName interface{} `json:"provider_name"`
	} `json:"metadata"`
}
