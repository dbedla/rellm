package rellm

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Non-empty struct: zero-size allocations all share one address, which would
// make the pointer-distinct assertions below meaningless.
type cloneTestCustomElement struct{ marker int }

func (*cloneTestCustomElement) Kind() ElementKind { return ElementKind("custom") }

func (*cloneTestCustomElement) Clone() ConversationElement { return &cloneTestCustomElement{} }

func TestCloneConversationElementImageGeneration(t *testing.T) {
	orig := &ImageGeneration{ID: "ig_1", Type: "image_generation_call", Status: "completed", Result: "data:image/jpeg;base64,AAA"}

	clone := orig.Clone().(*ImageGeneration)

	assert.NotSame(t, orig, clone)
	assert.Equal(t, orig, clone)

	clone.Result = "MUTATED"
	assert.Equal(t, "data:image/jpeg;base64,AAA", orig.Result)
}

func TestCloneConversationElementFunctionCallResp(t *testing.T) {
	orig := &FunctionCallResp{ID: "id_1", Type: "function_call_output", CallID: "call_1", Output: "42"}

	clone := orig.Clone().(*FunctionCallResp)

	assert.NotSame(t, orig, clone)
	assert.Equal(t, orig, clone)

	clone.Output = "MUTATED"
	assert.Equal(t, "42", orig.Output)
}

func TestCloneConversationElementFunctionCall(t *testing.T) {
	orig := &FunctionCall{ID: "id_1", Name: "get_data", Args: json.RawMessage(`{"a":1}`), CallID: "call_1"}

	clone := orig.Clone().(*FunctionCall)

	assert.NotSame(t, orig, clone)
	assert.Equal(t, orig, clone)

	clone.Args[0] = 'M'
	assert.Equal(t, json.RawMessage(`{"a":1}`), orig.Args)
}

func TestCloneConversationElementReasoning(t *testing.T) {
	orig := &Reasoning{
		ID:               "r1",
		Status:           "completed",
		Summary:          []ReasoningSummaryPart{{Type: "summary_text", Text: "thought"}},
		Text:             "reasoning text",
		Signature:        "sig",
		EncryptedContent: "enc",
		Format:           "text",
	}

	clone := orig.Clone().(*Reasoning)

	assert.NotSame(t, orig, clone)
	assert.Equal(t, orig, clone)

	clone.Summary[0].Text = "MUTATED"
	assert.Equal(t, "thought", orig.Summary[0].Text)
}

func TestCloneConversationElementMessages(t *testing.T) {
	mkContent := func() []MessagePart {
		return []MessagePart{{
			Type:     "input_text",
			Text:     "hello",
			ImageURL: &ImageURL{URL: "http://example.com/i.png"},
		}}
	}

	t.Run("user", func(t *testing.T) {
		orig := &UserMessage{MessageContent{ID: "u1", Role: "user", Status: "in_progress", Content: mkContent()}}
		clone := orig.Clone().(*UserMessage)
		assert.NotSame(t, orig, clone)
		assert.Equal(t, orig, clone)
		assertMessageContentClone(t, &orig.MessageContent, &clone.MessageContent)
	})

	t.Run("assistant", func(t *testing.T) {
		orig := &AssistantMessage{MessageContent{ID: "a1", Role: "assistant", Content: mkContent()}}
		clone := orig.Clone().(*AssistantMessage)
		assert.NotSame(t, orig, clone)
		assert.Equal(t, orig, clone)
		assertMessageContentClone(t, &orig.MessageContent, &clone.MessageContent)
	})

	t.Run("system", func(t *testing.T) {
		orig := &SystemMessage{MessageContent{ID: "s1", Role: "system", Content: mkContent()}}
		clone := orig.Clone().(*SystemMessage)
		assert.NotSame(t, orig, clone)
		assert.Equal(t, orig, clone)
		assertMessageContentClone(t, &orig.MessageContent, &clone.MessageContent)
	})
}

func assertMessageContentClone(t *testing.T, orig, clone *MessageContent) {
	t.Helper()

	assert.NotSame(t, orig.Content[0].ImageURL, clone.Content[0].ImageURL)

	clone.Content[0].Text = "MUTATED"
	clone.Content[0].ImageURL.URL = "MUTATED"

	assert.Equal(t, "hello", orig.Content[0].Text)
	assert.Equal(t, "http://example.com/i.png", orig.Content[0].ImageURL.URL)
}

func TestCloneConversationElementUnknownElement(t *testing.T) {
	orig := &UnknownElement{Provider: "openrouter", Type: "custom", Role: "assistant", Raw: json.RawMessage(`{"x":1}`)}

	clone := orig.Clone().(*UnknownElement)

	assert.NotSame(t, orig, clone)
	assert.Equal(t, orig, clone)

	clone.Raw[0] = 'M'
	assert.Equal(t, json.RawMessage(`{"x":1}`), orig.Raw)
}

func TestCloneConversationElements(t *testing.T) {
	assert.Len(t, cloneConversationElements(nil), 0)

	orig := []ConversationElement{
		&ImageGeneration{ID: "ig_1", Result: "r1"},
		&FunctionCallResp{CallID: "c1", Output: "o1"},
		&cloneTestCustomElement{},
	}

	clones := cloneConversationElements(orig)

	assert.Len(t, clones, len(orig))
	assert.NotSame(t, orig[0], clones[0])
	assert.NotSame(t, orig[1], clones[1])
	assert.NotSame(t, orig[2], clones[2]) // custom types are cloned via their own Clone

	clones[0].(*ImageGeneration).Result = "MUTATED"
	assert.Equal(t, "r1", orig[0].(*ImageGeneration).Result)
}
