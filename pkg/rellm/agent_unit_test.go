package rellm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptMessageToConversationMultimodal(t *testing.T) {
	prompt := "Hello with an image!"
	role := "user"
	msg, err := PromptMessageToConversation(prompt, role)
	assert.NoError(t, err)

	userMsg, ok := msg.(*UserMessage)
	assert.True(t, ok, "expected *UserMessage, got %T", msg)
	assert.Equal(t, KindUserMessage, userMsg.Kind())
	assert.Equal(t, role, userMsg.Role)
	assert.Len(t, userMsg.Content, 1)
	assert.Equal(t, "input_text", userMsg.Content[0].Type)
	assert.Equal(t, prompt, userMsg.Content[0].Text)
}
