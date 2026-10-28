package rellm

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// pngHeader is a minimal PNG signature; http.DetectContentType maps it to
// image/png.
var pngHeader = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

func TestPromptBuilder_Build_ImageOnlyPrompt(t *testing.T) {
	prompt, err := NewPromptBuilder().
		WithImageBytes(pngHeader, ImageOption{}).
		Build()

	assert.NoError(t, err)
	assert.Empty(t, prompt.msg)     // no message, images only — still valid
	assert.Len(t, prompt.images, 1) // and the image made it through
	assert.Equal(t, "input_image", prompt.images[0].Type)
	assert.Equal(t, "data:image/png;base64,iVBORw0KGgo=", *prompt.images[0].ImageURL)
	assert.Equal(t, ImageDetailAuto, *prompt.images[0].Detail)
}

func TestPromptBuilder_Build_NeitherMessageNorImage(t *testing.T) {
	_, err := NewPromptBuilder().Build()
	assert.ErrorIs(t, err, ErrEmptyPrompt)

	_, err = NewPromptBuilder().WithMessage("   ").Build()
	assert.ErrorIs(t, err, ErrEmptyPrompt)
}

func TestPromptBuilder_WithImageURL_Validation(t *testing.T) {
	tests := []struct {
		name string
		url  string
		err  error
	}{
		{"empty", "  ", ErrEmptyImageURL},
		{"data url misrouted", "data:image/png;base64,iVBORw0KGgo=", ErrImageURLIsDataURL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPromptBuilder().
				WithMessage("hi").
				WithImageURL(tt.url, ImageOption{}).
				Build()
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestPromptBuilder_WithImageDataURL_Validation(t *testing.T) {
	_, err := NewPromptBuilder().
		WithMessage("hi").
		WithImageDataURL("https://example.com/img.png", ImageOption{}).
		Build()
	assert.ErrorIs(t, err, ErrInvalidImageDataURL)

	_, err = NewPromptBuilder().
		WithMessage("hi").
		WithImageDataURL("data:image/png;base64,iVBORw0KGgo=", ImageOption{}).
		Build()
	assert.NoError(t, err)
}

func TestPromptBuilder_WithImageBytes_Validation(t *testing.T) {
	_, err := NewPromptBuilder().
		WithMessage("hi").
		WithImageBytes([]byte("just text"), ImageOption{}).
		Build()
	assert.ErrorIs(t, err, ErrInvalidImageData)

	_, err = NewPromptBuilder().
		WithMessage("hi").
		WithImageBytes(nil, ImageOption{}).
		Build()
	assert.ErrorIs(t, err, ErrInvalidImageData)
}

func TestPromptBuilder_WithImageFile_Validation(t *testing.T) {
	_, err := NewPromptBuilder().
		WithMessage("hi").
		WithImageFile(filepath.Join(t.TempDir(), "missing.png"), ImageOption{}).
		Build()
	assert.ErrorIs(t, err, ErrImageFileUnreadable)

	path := filepath.Join(t.TempDir(), "img.png")
	assert.NoError(t, os.WriteFile(path, pngHeader, 0o600))
	prompt, err := NewPromptBuilder().
		WithMessage("hi").
		WithImageFile(path, ImageOption{}).
		Build()
	assert.NoError(t, err)
	assert.Equal(t, "data:image/png;base64,iVBORw0KGgo=", *prompt.images[0].ImageURL)
}

func TestPromptBuilder_ImageDetail(t *testing.T) {
	low := ImageDetailLow
	prompt, err := NewPromptBuilder().
		WithMessage("hi").
		WithImageBytes(pngHeader, ImageOption{Detail: &low}).
		Build()
	assert.NoError(t, err)
	assert.Equal(t, ImageDetailLow, *prompt.images[0].Detail)

	invalid := ImageDetail("ultra")
	_, err = NewPromptBuilder().
		WithMessage("hi").
		WithImageBytes(pngHeader, ImageOption{Detail: &invalid}).
		Build()
	assert.ErrorIs(t, err, ErrInvalidImageDetail)
}

func TestPromptBuilder_ImagesAccumulateInOrder(t *testing.T) {
	prompt, err := NewPromptBuilder().
		WithMessage("compare these").
		WithImageURL("https://example.com/a.png", ImageOption{}).
		WithImageDataURL("data:image/png;base64,iVBORw0KGgo=", ImageOption{}).
		Build()

	assert.NoError(t, err)
	assert.Equal(t, "compare these", prompt.msg)
	assert.Len(t, prompt.images, 2)
	assert.Equal(t, "input_image", prompt.images[0].Type)
	assert.Equal(t, "https://example.com/a.png", *prompt.images[0].ImageURL)
	assert.Equal(t, ImageDetailAuto, *prompt.images[0].Detail)
	assert.Equal(t, "data:image/png;base64,iVBORw0KGgo=", *prompt.images[1].ImageURL)
}

// TestStdToConversationElements_InputImageEcho is the regression for the
// wire fix: providers echo input_image parts with a string image_url; the
// old object-typed field silently parsed them to empty content.
func TestStdToConversationElements_InputImageEcho(t *testing.T) {
	item := json.RawMessage(`{
		"type": "message",
		"role": "user",
		"content": [
			{"type": "input_text", "text": "what is in this image?"},
			{"type": "input_image", "image_url": "https://example.com/cat.png", "detail": "low"}
		]
	}`)

	p := &OpenRouterProvider{}
	elements, err := p.ToConversationElements([]json.RawMessage{item})
	assert.NoError(t, err)
	assert.Len(t, elements, 1)

	userMsg, ok := elements[0].(*UserMessage)
	assert.True(t, ok, "expected *UserMessage, got %T", elements[0])
	assert.Len(t, userMsg.Content, 2)

	imagePart := userMsg.Content[1]
	assert.Equal(t, "input_image", imagePart.Type)
	assert.Equal(t, "https://example.com/cat.png", *imagePart.ImageURL)
	assert.Equal(t, ImageDetailLow, *imagePart.Detail)
}

func TestStdToProviderRepresentation_InputImageWire(t *testing.T) {
	url := "https://example.com/cat.png"
	detail := ImageDetailHigh
	msg := UserMessage{MessageContent{
		Role: "user",
		Content: []MessagePart{
			{Type: "input_text", Text: "what is in this image?"},
			{Type: "input_image", ImageURL: &url, Detail: &detail},
		},
	}}

	raw, err := StdToProviderRepresentation([]ConversationElement{&msg})
	assert.NoError(t, err)

	var wire struct {
		Content []json.RawMessage `json:"content"`
	}
	assert.NoError(t, json.Unmarshal(raw[0], &wire))
	assert.Len(t, wire.Content, 2)

	var imagePart map[string]any
	assert.NoError(t, json.Unmarshal(wire.Content[1], &imagePart))
	assert.Equal(t, "input_image", imagePart["type"])
	// image_url stays a plain string, never an object (Responses API shape).
	assert.Equal(t, url, imagePart["image_url"])
	assert.Equal(t, "high", imagePart["detail"])
}

func TestPromptBuilder_WithImageFileID(t *testing.T) {
	prompt, err := NewPromptBuilder().
		WithMessage("hi").
		WithImageFileID("file-123", ImageOption{}).
		Build()
	assert.NoError(t, err)
	assert.Len(t, prompt.images, 1)
	assert.Nil(t, prompt.images[0].ImageURL) // file_id part, no url
	assert.Equal(t, "file-123", *prompt.images[0].FileID)
	assert.Equal(t, ImageDetailAuto, *prompt.images[0].Detail)

	_, err = NewPromptBuilder().
		WithMessage("hi").
		WithImageFileID("  ", ImageOption{}).
		Build()
	assert.ErrorIs(t, err, ErrEmptyImageFileID)
}

func TestPromptBuilder_FileIDOnOtherDoorsRejected(t *testing.T) {
	fid := "file-123"
	_, err := NewPromptBuilder().
		WithMessage("hi").
		WithImageURL("https://example.com/a.png", ImageOption{FileID: &fid}).
		Build()
	assert.ErrorIs(t, err, ErrImageURLAndFileID)

	_, err = NewPromptBuilder().
		WithMessage("hi").
		WithImageDataURL("data:image/png;base64,iVBORw0KGgo=", ImageOption{FileID: &fid}).
		Build()
	assert.ErrorIs(t, err, ErrImageURLAndFileID)

	_, err = NewPromptBuilder().
		WithMessage("hi").
		WithImageBytes(pngHeader, ImageOption{FileID: &fid}).
		Build()
	assert.ErrorIs(t, err, ErrImageURLAndFileID)

	_, err = NewPromptBuilder().
		WithMessage("hi").
		WithImageFile(filepath.Join(t.TempDir(), "x.png"), ImageOption{FileID: &fid}).
		Build()
	assert.ErrorIs(t, err, ErrImageURLAndFileID)
}

func TestPromptBuilder_CacheBreakpoint(t *testing.T) {
	bp := PromptCacheBreakpoint{TTL: PromptCacheTTL30m} // Mode empty → explicit
	prompt, err := NewPromptBuilder().
		WithMessage("hi").
		WithImageURL("https://example.com/a.png", ImageOption{CacheBreakpoint: &bp}).
		Build()
	assert.NoError(t, err)
	assert.NotNil(t, prompt.images[0].CacheBreakpoint)
	assert.Equal(t, PromptCacheBreakpointModeExplicit, prompt.images[0].CacheBreakpoint.Mode)
	assert.Equal(t, PromptCacheTTL30m, prompt.images[0].CacheBreakpoint.TTL)

	badMode := PromptCacheBreakpoint{Mode: "auto"}
	_, err = NewPromptBuilder().
		WithMessage("hi").
		WithImageURL("https://example.com/a.png", ImageOption{CacheBreakpoint: &badMode}).
		Build()
	assert.ErrorIs(t, err, ErrInvalidPromptCacheBreakpointMode)

	badTTL := PromptCacheBreakpoint{TTL: "1h"}
	_, err = NewPromptBuilder().
		WithMessage("hi").
		WithImageURL("https://example.com/a.png", ImageOption{CacheBreakpoint: &badTTL}).
		Build()
	assert.ErrorIs(t, err, ErrInvalidPromptCacheBreakpointTTL)
}

func TestStdToProviderRepresentation_FileIDWire(t *testing.T) {
	fid := "file-123"
	detail := ImageDetailLow
	bp := PromptCacheBreakpoint{Mode: PromptCacheBreakpointModeExplicit, TTL: PromptCacheTTL12h}
	msg := UserMessage{MessageContent{
		Role: "user",
		Content: []MessagePart{
			{Type: "input_image", FileID: &fid, Detail: &detail, CacheBreakpoint: &bp},
		},
	}}

	raw, err := StdToProviderRepresentation([]ConversationElement{&msg})
	assert.NoError(t, err)

	var wire struct {
		Content []json.RawMessage `json:"content"`
	}
	assert.NoError(t, json.Unmarshal(raw[0], &wire))
	assert.Len(t, wire.Content, 1)

	var part map[string]any
	assert.NoError(t, json.Unmarshal(wire.Content[0], &part))
	assert.Equal(t, "input_image", part["type"])
	assert.Equal(t, "file-123", part["file_id"])
	assert.Equal(t, "low", part["detail"])
	assert.NotContains(t, part, "image_url") // file_id part, no url
	breakpoint, ok := part["prompt_cache_breakpoint"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "explicit", breakpoint["mode"])
	assert.Equal(t, "12h", breakpoint["ttl"])
}

func TestFilesystemConversation_ImagePartRoundTrip(t *testing.T) {
	url := "data:image/png;base64,iVBORw0KGgo="
	msg := UserMessage{MessageContent{
		Role: "user",
		Content: []MessagePart{
			{Type: "input_text", Text: "what is in this image?"},
			{Type: "input_image", ImageURL: &url},
		},
	}}

	conv := NewFilesystemConversation(filepath.Join(t.TempDir(), "conversation.jsonl"))
	assert.NoError(t, conv.Append(context.Background(), []ConversationElement{&msg}))

	loaded, err := conv.Load(context.Background())
	assert.NoError(t, err)
	assert.Len(t, loaded, 1)

	userMsg, ok := loaded[0].(*UserMessage)
	assert.True(t, ok, "expected *UserMessage, got %T", loaded[0])
	assert.Equal(t, KindUserMessage, userMsg.Kind())
	assert.Len(t, userMsg.Content, 2)
	assert.Equal(t, "input_image", userMsg.Content[1].Type)
	assert.Equal(t, url, *userMsg.Content[1].ImageURL)
}
