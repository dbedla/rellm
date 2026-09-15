package rellm

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptMessageToConversationMultimodal(t *testing.T) {
	prompt := "Hello with an image!"
	role := "user"
	msg, err := promptMessageToConversation(prompt, role)
	assert.NoError(t, err)

	userMsg, ok := msg.(*UserMessage)
	assert.True(t, ok, "expected *UserMessage, got %T", msg)
	assert.Equal(t, KindUserMessage, userMsg.Kind())
	assert.Equal(t, role, userMsg.Role)
	assert.Len(t, userMsg.Content, 1)
	assert.Equal(t, "input_text", userMsg.Content[0].Type)
	assert.Equal(t, prompt, userMsg.Content[0].Text)
}

func TestCloseWithError(t *testing.T) {
	funcWithNamedErr := func() (finalErr error) {
		cer := &closWithErr{}
		defer closeWithError(&finalErr, cer)
		return nil
	}
	err := funcWithNamedErr()
	assert.ErrorIs(t, err, errTestCloseErr)
}

func TestCloseWithErrorAndFunctionError(t *testing.T) {
	funcWithNamedErr := func() (finalErr error) {
		cer := &closWithErr{}
		defer closeWithError(&finalErr, cer)
		return errTestErr
	}
	err := funcWithNamedErr()
	assert.ErrorIs(t, err, errTestCloseErr)
	assert.ErrorIs(t, err, errTestErr)
}

func TestCloseWithErrorErrArgIsNil(t *testing.T) {
	funcWithNamedErr := func() (finalErr error) {
		cer := &closWithErr{}
		defer closeWithError(nil, cer)
		return errTestErr
	}
	err := funcWithNamedErr()
	assert.ErrorIs(t, err, errTestErr)
}

func TestCloseWithErrorCloserArgIsNil(t *testing.T) {
	funcWithNamedErr := func() (finalErr error) {
		defer closeWithError(&finalErr, nil)
		return errTestErr
	}
	err := funcWithNamedErr()
	assert.ErrorIs(t, err, errTestErr)
}

func TestCloseNoError(t *testing.T) {
	funcWithNamedErr := func() (finalErr error) {
		cer := &closNoErr{}
		defer closeWithError(&finalErr, cer)
		return nil
	}
	err := funcWithNamedErr()
	assert.NoError(t, err)
}

var errTestCloseErr = fmt.Errorf("error closing")
var errTestErr = fmt.Errorf("error test")

type closWithErr struct {
}

func (c *closWithErr) Close() error {
	return errTestCloseErr
}

type closNoErr struct {
}

func (c *closNoErr) Close() error {
	return nil
}

func TestFuncResultToFunctionCallRespSerializesOutputAsString(t *testing.T) {
	resp := funcResultToFunctionCallResp("call_123", int64(42))

	jsonResp, err := json.Marshal(resp)
	assert.NoError(t, err, "failed to marshal function call response")
	assert.JSONEq(t, `{"type":"function_call_output","call_id":"call_123","output":"42"}`, string(jsonResp))
}

func TestCutBodySnippet(t *testing.T) {
	emoji := "😂"
	assert.Len(t, emoji, 4)
	stringInput := "important string data and "
	extras := " more data"
	combined := stringInput + emoji + extras
	lenBasedOnChar := len(stringInput) + 1

	trimmed := bodySnippet([]byte(combined), lenBasedOnChar)
	assert.Equal(t, stringInput+emoji+"...", trimmed)
}
