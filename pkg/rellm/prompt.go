package rellm

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// ImageDetail controls the fidelity at which the model processes an input
// image. One of ImageDetailAuto, ImageDetailLow, ImageDetailHigh, or
// ImageDetailOriginal. Defaults to auto; always sent on the wire because
// OpenRouter requires the field.
type ImageDetail string

const (
	// ImageDetailAuto lets the provider choose the fidelity (its default).
	ImageDetailAuto ImageDetail = "auto"
	// ImageDetailLow processes the image at low fidelity (cheaper tokens).
	ImageDetailLow ImageDetail = "low"
	// ImageDetailHigh processes the image at high fidelity.
	ImageDetailHigh ImageDetail = "high"
	// ImageDetailOriginal processes the image at its native resolution.
	ImageDetailOriginal ImageDetail = "original"
)

// PromptCacheTTL is the minimum lifetime of a prompt cache breakpoint.
type PromptCacheTTL string

const (
	// PromptCacheTTL30m caches the breakpoint prefix for 30 minutes (default).
	PromptCacheTTL30m PromptCacheTTL = "30m"
	// PromptCacheTTL12h caches the breakpoint prefix for 12 hours (billed at
	// the higher cache-write rate).
	PromptCacheTTL12h PromptCacheTTL = "12h"
)

// PromptCacheBreakpointModeExplicit is the only breakpoint mode; the wire
// field is required, so an empty Mode is normalized to it at Build.
const PromptCacheBreakpointModeExplicit = "explicit"

// PromptCacheBreakpoint marks an explicit prompt-cache boundary on one
// content block: everything through the block carrying it is a candidate
// cached prefix. Natively supported by OpenAI GPT-5.6+; OpenRouter converts
// the marker to Anthropic-style cache_control for providers that use it.
// TTL empty inherits the request-level TTL.
type PromptCacheBreakpoint struct {
	// Mode is the breakpoint mode; only "explicit" exists, so an empty
	// value defaults to explicit at Build.
	Mode string `json:"mode"`
	// TTL is the minimum cache lifetime: PromptCacheTTL30m or
	// PromptCacheTTL12h; empty inherits the request-level TTL.
	TTL PromptCacheTTL `json:"ttl,omitempty"`
}

// ImageOption carries the optional fields of one image input for the
// WithImage* methods. Every field is a pointer: nil means the default.
// The zero value is valid and means "all defaults".
type ImageOption struct {
	// Detail sets the image detail level; nil defaults to ImageDetailAuto.
	Detail *ImageDetail `json:"detail,omitempty"`
	// FileID references an image uploaded via the provider's Files API
	// (input_image.file_id) instead of inline data or a URL. Only
	// meaningful with WithImageFileID; rejected on the other doors because
	// input_image carries either image_url or file_id, never both.
	FileID *string `json:"file_id,omitempty"`
	// CacheBreakpoint marks a prompt-cache boundary through this image.
	CacheBreakpoint *PromptCacheBreakpoint `json:"prompt_cache_breakpoint,omitempty"`
}

// imageInputKind discriminates raw image inputs before Build resolves them.
type imageInputKind uint8

const (
	imageInputURL imageInputKind = iota
	imageInputDataURL
	imageInputBytes
	imageInputFile
	imageInputFileID
)

// promptImageInput stores one raw WithImage* call until Build converts it.
type promptImageInput struct {
	kind imageInputKind
	url  string // url, data url, or file path
	data []byte // imageInputBytes
	opts ImageOption
}

// Prompt holds the instruction and inference parameters for one interaction
// with the agent. An interaction may involve multiple provider round-trips
// (e.g. tool calls). Build prompts with PromptBuilder.
type Prompt struct {
	msg    string
	images []MessagePart // resolved input_image parts, sent with the message
	params promptParams
}

// empty reports whether the prompt carries no message and no images.
func (p *Prompt) empty() bool {
	return p.msg == "" && len(p.images) == 0
}

// promptParams holds inference parameters configured via WithXxx methods.
type promptParams struct {
	Temperature      *float64
	Reasoning        *ReasoningConfig
	MaxOutputTokens  int
	TopP             *float64
	PresencePenalty  *float64
	FrequencyPenalty *float64
	TopLogprobs      int
}

// PromptBuilder constructs a Prompt, validating its fields on Build.
//
// Usage example:
//
//	prompt, err := rellm.NewPromptBuilder().
//		WithMessage("Write a haiku about tide pools.").
//		WithTemperature(0.7).
//		WithMaxOutputTokens(100).
//		Build()
//	if err != nil {
//		log.Fatal(err)
//	}
//	finalReport, err := agent.Execute(ctx, prompt)
//	answer := finalReport.Message
//
// The WithXxx methods are optional; a message (WithMessage) or at least
// one image (WithImage*) is required — image-only prompts are valid.
type PromptBuilder struct {
	msg    string
	images []promptImageInput
	params promptParams
}

// NewPromptBuilder starts a prompt builder chain with default parameters.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		params: promptParams{},
	}
}

// WithMessage sets the user message. Required.
func (b *PromptBuilder) WithMessage(msg string) *PromptBuilder {
	b.msg = msg
	return b
}

// WithImageURL adds an image the provider fetches from a web URL. The URL
// is sent as-is; a data URL here is rejected at Build (use WithImageDataURL).
// Repeat calls accumulate images in call order. Optional fields go in opts;
// the zero ImageOption is valid.
// See https://developers.openai.com/api/docs/guides/images-vision
func (b *PromptBuilder) WithImageURL(url string, opts ImageOption) *PromptBuilder {
	b.images = append(b.images, promptImageInput{kind: imageInputURL, url: url, opts: opts})
	return b
}

// WithImageDataURL adds an image supplied as a data URL
// ("data:image/png;base64,..."). The data URL is sent as-is after a prefix
// check at Build. Repeat calls accumulate images in call order.
func (b *PromptBuilder) WithImageDataURL(dataURL string, opts ImageOption) *PromptBuilder {
	b.images = append(b.images, promptImageInput{kind: imageInputDataURL, url: dataURL, opts: opts})
	return b
}

// WithImageBytes adds an image supplied as raw bytes. The MIME type is
// sniffed (http.DetectContentType) and the bytes are base64-encoded into a
// data URL at Build. Repeat calls accumulate images in call order.
func (b *PromptBuilder) WithImageBytes(data []byte, opts ImageOption) *PromptBuilder {
	b.images = append(b.images, promptImageInput{kind: imageInputBytes, data: data, opts: opts})
	return b
}

// WithImageFile adds an image read from a file path. The file is read at
// Build (With* methods do not fail), its MIME type sniffed, and its bytes
// base64-encoded into a data URL. Repeat calls accumulate images in call order.
func (b *PromptBuilder) WithImageFile(path string, opts ImageOption) *PromptBuilder {
	b.images = append(b.images, promptImageInput{kind: imageInputFile, url: path, opts: opts})
	return b
}

// WithImageFileID adds an image uploaded via the provider's Files API,
// referenced by its file ID: the part carries input_image.file_id instead
// of image_url. opts.FileID is redundant here (the door's own ID wins) and
// rejected on every other door. Repeat calls accumulate images in call order.
func (b *PromptBuilder) WithImageFileID(fileID string, opts ImageOption) *PromptBuilder {
	b.images = append(b.images, promptImageInput{kind: imageInputFileID, url: fileID, opts: opts})
	return b
}

// WithTemperature sets the sampling temperature (usually 0..2). Higher values
// make output more random; lower values more deterministic. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#temperature
func (b *PromptBuilder) WithTemperature(t float64) *PromptBuilder {
	b.params.Temperature = &t
	return b
}

// WithReasoning sets the reasoning effort for models that support it. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#reasoning
func (b *PromptBuilder) WithReasoning(effort ReasoningEffort) *PromptBuilder {

	b.params.Reasoning = &ReasoningConfig{Effort: effort}
	return b
}

// WithMaxOutputTokens caps the output length of every provider call in this
// prompt. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#max-tokens
func (b *PromptBuilder) WithMaxOutputTokens(n int) *PromptBuilder {
	b.params.MaxOutputTokens = n
	return b
}

// WithTopP sets nucleus sampling: only tokens within the top-p probability
// mass are considered. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#top-p
func (b *PromptBuilder) WithTopP(t float64) *PromptBuilder {
	b.params.TopP = &t
	return b
}

// WithPresencePenalty discourages repeating already-used tokens (range -2..2).
// Optional.
// See https://openrouter.ai/docs/api_reference/parameters#presence-penalty
func (b *PromptBuilder) WithPresencePenalty(p float64) *PromptBuilder {
	b.params.PresencePenalty = &p
	return b
}

// WithFrequencyPenalty discourages frequent tokens proportionally to their
// repetition (range -2..2). Optional.
// See https://openrouter.ai/docs/api_reference/parameters#frequency-penalty
func (b *PromptBuilder) WithFrequencyPenalty(f float64) *PromptBuilder {
	b.params.FrequencyPenalty = &f
	return b
}

// WithTopLogprobs sets how many of the most likely tokens are returned per
// output position when logprobs are enabled. Optional.
// See https://openrouter.ai/docs/api_reference/parameters#top-logprobs
func (b *PromptBuilder) WithTopLogprobs(n int) *PromptBuilder {
	b.params.TopLogprobs = n
	return b
}

// Build validates the accumulated WithXxx calls and returns the Prompt. A
// message or at least one image is required (message-only, image-only, and
// mixed prompts are all valid); the message must not be whitespace-only.
// Images are resolved here: web and data URLs are validated, bytes and files
// are sniffed and encoded as data URLs, and each image's detail defaults to
// ImageDetailAuto.
func (b *PromptBuilder) Build() (*Prompt, error) {
	trimmedMsg := strings.TrimSpace(b.msg)

	images, err := resolveImages(b.images)
	if err != nil {
		return nil, err
	}

	if trimmedMsg == "" && len(images) == 0 {
		return nil, ErrEmptyPrompt
	}

	if b.params.Reasoning != nil {
		if b.params.Reasoning.Effort == "" {
			return nil, ErrEmptyReasoningEffort
		}
	}
	return &Prompt{
		msg:    trimmedMsg,
		images: images,
		params: b.params,
	}, nil
}

// resolveImages converts raw image inputs into input_image message parts,
// validating each one and normalizing its detail level and cache breakpoint.
func resolveImages(inputs []promptImageInput) ([]MessagePart, error) {
	parts := make([]MessagePart, 0, len(inputs))
	for _, in := range inputs {
		detail, err := in.opts.resolveDetail()
		if err != nil {
			return nil, err
		}
		breakpoint, err := in.opts.resolveCacheBreakpoint()
		if err != nil {
			return nil, err
		}

		part := MessagePart{
			Type:            "input_image",
			Detail:          detail,
			CacheBreakpoint: breakpoint,
		}

		if in.kind == imageInputFileID {
			fileID := strings.TrimSpace(in.url)
			if fileID == "" {
				return nil, ErrEmptyImageFileID
			}
			part.FileID = &fileID
			parts = append(parts, part)
			continue
		}

		// The remaining doors all carry an image_url; a file id alongside
		// one would send both, and the wire part accepts either, never both.
		if in.opts.FileID != nil {
			return nil, errors.Join(ErrImageURLAndFileID, fmt.Errorf("file id set on a url/data door; use WithImageFileID"))
		}

		var url string
		switch in.kind {
		case imageInputURL:
			url = strings.TrimSpace(in.url)
			if url == "" {
				return nil, ErrEmptyImageURL
			}
			if strings.HasPrefix(strings.ToLower(url), "data:") {
				return nil, errors.Join(ErrImageURLIsDataURL, fmt.Errorf("url: %s", in.url))
			}

		case imageInputDataURL:
			if !strings.HasPrefix(strings.ToLower(in.url), "data:image/") {
				return nil, errors.Join(ErrInvalidImageDataURL, fmt.Errorf("data url must start with data:image/, got: %s", in.url))
			}
			url = in.url

		case imageInputBytes:
			dataURL, err := imageDataURL(in.data)
			if err != nil {
				return nil, err
			}
			url = dataURL

		case imageInputFile:
			data, err := os.ReadFile(in.url)
			if err != nil {
				return nil, errors.Join(ErrImageFileUnreadable, err)
			}
			dataURL, err := imageDataURL(data)
			if err != nil {
				return nil, err
			}
			url = dataURL
		}

		part.ImageURL = &url
		parts = append(parts, part)
	}
	return parts, nil
}

// resolveDetail validates the option's detail level and defaults to auto.
func (o ImageOption) resolveDetail() (*ImageDetail, error) {
	if o.Detail == nil {
		auto := ImageDetailAuto
		return &auto, nil
	}
	switch *o.Detail {
	case ImageDetailAuto, ImageDetailLow, ImageDetailHigh, ImageDetailOriginal:
		detail := *o.Detail
		return &detail, nil
	default:
		return nil, errors.Join(ErrInvalidImageDetail, fmt.Errorf("unknown image detail %q, want one of auto, low, high, original", *o.Detail))
	}
}

// resolveCacheBreakpoint validates and normalizes the option's cache
// breakpoint; nil stays nil (no breakpoint on the wire).
func (o ImageOption) resolveCacheBreakpoint() (*PromptCacheBreakpoint, error) {
	if o.CacheBreakpoint == nil {
		return nil, nil
	}
	bp := *o.CacheBreakpoint
	bp.Mode = strings.TrimSpace(bp.Mode)
	if bp.Mode == "" {
		bp.Mode = PromptCacheBreakpointModeExplicit
	}
	if bp.Mode != PromptCacheBreakpointModeExplicit {
		return nil, errors.Join(ErrInvalidPromptCacheBreakpointMode, fmt.Errorf("mode %q, want %q", bp.Mode, PromptCacheBreakpointModeExplicit))
	}
	bp.TTL = PromptCacheTTL(strings.TrimSpace(string(bp.TTL)))
	if bp.TTL != "" && bp.TTL != PromptCacheTTL30m && bp.TTL != PromptCacheTTL12h {
		return nil, errors.Join(ErrInvalidPromptCacheBreakpointTTL, fmt.Errorf("ttl %q, want %q or %q", bp.TTL, PromptCacheTTL30m, PromptCacheTTL12h))
	}
	return &bp, nil
}

// imageDataURL sniffs the content type of data and encodes it as a data URL.
// Anything http.DetectContentType does not recognize as image/* is rejected.
func imageDataURL(data []byte) (string, error) {
	mime := http.DetectContentType(data)
	if !strings.HasPrefix(mime, "image/") {
		return "", errors.Join(ErrInvalidImageData, fmt.Errorf("sniffed content type %q, want image/*", mime))
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

func promptMessageToConversation(text, role string, images []MessagePart) (ConversationElement, error) {
	content := make([]MessagePart, 0, len(images)+1)
	if text != "" {
		content = append(content, MessagePart{Type: "input_text", Text: text})
	}
	content = append(content, images...)
	switch role {
	case "user":
		return &UserMessage{MessageContent{Role: role, Content: content}}, nil
	case "system":
		return &SystemMessage{MessageContent{Role: role, Content: content}}, nil
	case "assistant":
		return &AssistantMessage{MessageContent{Role: role, Content: content}}, nil
	default:
		return nil, fmt.Errorf("promptMessageToConversation: unknown role %q", role)
	}
}
