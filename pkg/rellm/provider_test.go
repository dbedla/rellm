package rellm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextFromContent(t *testing.T) {
	parts := []MessagePart{
		{Type: "input_text", Text: "hello"},
		{Type: "input_text", Text: ""},
		{Type: "input_text", Text: "world"},
	}
	assert.Equal(t, "hello world", TextFromContent(parts))

	assert.Empty(t, TextFromContent(nil))
}
