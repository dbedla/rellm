package rellm

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptMessageToConversationMultimodal(t *testing.T) {
	prompt := "Hello with an image!"
	role := "user"
	msg, err := PromptMessageToConversation(prompt, role)
	assert.NoError(t, err)

	var userMsg UserMessage
	err = json.Unmarshal(msg, &userMsg)
	assert.NoError(t, err)
	assert.Equal(t, role, userMsg.Role)
	assert.Len(t, userMsg.Content, 1)
	assert.Equal(t, "input_text", userMsg.Content[0].Type)
	assert.Equal(t, prompt, userMsg.Content[0].Text)
}

func TestHandleMessageMultimodal(t *testing.T) {
	item := OutputItem{
		Type: "message",
		Content: []MessagePart{
			{Type: "text", Text: "Here is an image: "},
			{Type: "image_url", ImageURL: &ImageURL{URL: "http://example.com/image.png"}},
			{Type: "text", Text: "It is a nice image."},
		},
	}

	elements, msgResp, err := handleMessage(item)
	assert.NoError(t, err)
	assert.Len(t, elements, 1)

	var assistantMsg UserMessage
	err = json.Unmarshal(elements[0], &assistantMsg)
	assert.NoError(t, err)
	assert.Equal(t, "assistant", assistantMsg.Role)
	assert.Len(t, assistantMsg.Content, 3)
	assert.Equal(t, "Here is an image:  It is a nice image.", msgResp)
}
