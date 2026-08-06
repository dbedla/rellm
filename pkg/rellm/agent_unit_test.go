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
