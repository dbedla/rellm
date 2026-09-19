package rellm

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testUserMsg(text string) *UserMessage {
	return &UserMessage{MessageContent{Role: "user", Content: []MessagePart{{Type: "input_text", Text: text}}}}
}

func testAssistantMsg(text string) *AssistantMessage {
	return &AssistantMessage{MessageContent{Role: "assistant", Content: []MessagePart{{Type: "output_text", Text: text}}}}
}

func assertElementJSONEq(t *testing.T, expected string, actual ConversationElement) {
	t.Helper()
	b, err := json.Marshal(actual)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(b))
}

func TestInMemoryConversation_LoadEmpty(t *testing.T) {
	s := NewInMemoryConversation()
	ctx := context.Background()
	msgs, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestInMemoryConversation_AppendAndLoad(t *testing.T) {
	s := NewInMemoryConversation()
	delta := []ConversationElement{testUserMsg("hi")}
	ctx := context.Background()
	assert.NoError(t, s.Append(ctx, delta))
	msgs, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Len(t, msgs, 1)
	assertElementJSONEq(t, `{"kind":"user_message","role":"user","content":[{"type":"input_text","text":"hi"}]}`, msgs[0])
}

func TestInMemoryConversation_AppendMultiple(t *testing.T) {
	s := NewInMemoryConversation()
	first := []ConversationElement{testUserMsg("a")}
	second := []ConversationElement{testAssistantMsg("b")}
	ctx := context.Background()
	assert.NoError(t, s.Append(ctx, first))
	assert.NoError(t, s.Append(ctx, second))
	msgs, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Len(t, msgs, 2)
	assert.Equal(t, KindUserMessage, msgs[0].Kind())
	assert.Equal(t, KindAssistantMessage, msgs[1].Kind())
}

func TestInMemoryConversation_LoadDoesNotAlias(t *testing.T) {
	s := NewInMemoryConversation()
	ctx := context.Background()
	assert.NoError(t, s.Append(ctx, []ConversationElement{testUserMsg("x")}))

	loaded, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Len(t, loaded, 1)
	loaded = append(loaded, testAssistantMsg("y"))
	assert.Len(t, loaded, 2)

	again, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Len(t, again, 1)
}

func TestInMemoryConversation_LoadDoesNotAliasElements(t *testing.T) {
	s := NewInMemoryConversation()
	ctx := context.Background()
	assert.NoError(t, s.Append(ctx, []ConversationElement{testUserMsg("x")}))

	loaded, err := s.Load(ctx)
	assert.NoError(t, err)
	loaded[0].(*UserMessage).Content[0].Text = "mutated"

	again, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "x", again[0].(*UserMessage).Content[0].Text)
}

func TestInMemoryConversation_AppendEmpty(t *testing.T) {
	s := NewInMemoryConversation()
	ctx := context.Background()
	assert.NoError(t, s.Append(ctx, nil))
	msgs, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestFilesystemConversation_LoadMissingFile(t *testing.T) {
	s := NewFilesystemConversation(filepath.Join(t.TempDir(), "missing.jsonl"))
	ctx := context.Background()
	msgs, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestFilesystemConversation_AppendAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conv.jsonl")
	s := NewFilesystemConversation(path)
	delta := []ConversationElement{testUserMsg("hi"), testAssistantMsg("hello")}
	ctx := context.Background()
	err := s.Append(ctx, delta)
	assert.NoError(t, err)

	// Storage format: one canonical kind-tagged element per line (map marshal
	// sorts keys alphabetically).
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	expected := `{"content":[{"type":"input_text","text":"hi"}],"kind":"user_message","role":"user"}` + "\n" +
		`{"content":[{"type":"output_text","text":"hello"}],"kind":"assistant_message","role":"assistant"}` + "\n"
	assert.Equal(t, expected, string(content))

	msgs, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Len(t, msgs, 2)
	assert.Equal(t, delta[0].Kind(), msgs[0].Kind())
	assert.Equal(t, delta[1].Kind(), msgs[1].Kind())
}

func TestFilesystemConversation_AppendAndLoadUnknownElement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conv.jsonl")
	s := NewFilesystemConversation(path)
	raw := json.RawMessage(`{"type":"future_output","value":42}`)
	ctx := context.Background()

	err := s.Append(ctx, []ConversationElement{newUnknownElement(string(ProviderTagProviderOpenRouter), "future_output", "", raw)})
	assert.NoError(t, err)

	elements, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Len(t, elements, 1)
	unknown, ok := elements[0].(*UnknownElement)
	assert.True(t, ok)
	assert.Equal(t, string(ProviderTagProviderOpenRouter), unknown.Provider)
	assert.Equal(t, "future_output", unknown.Type)
	assert.JSONEq(t, string(raw), string(unknown.Raw))
}

func TestFilesystemConversation_LoadMalformedLineReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conv.jsonl")
	content := `{"kind":"user_message"}` + "\n" + `{broken` + "\n"
	assert.NoError(t, os.WriteFile(path, []byte(content), 0644))
	ctx := context.Background()
	s := NewFilesystemConversation(path)
	_, err := s.Load(ctx)
	assert.ErrorIs(t, err, ErrMalformedConversationLine)
}

func TestFilesystemConversation_LoadUnknownKindReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conv.jsonl")
	content := `{"kind":"time_travel","role":"user"}` + "\n"
	assert.NoError(t, os.WriteFile(path, []byte(content), 0644))
	s := NewFilesystemConversation(path)
	ctx := context.Background()
	_, err := s.Load(ctx)
	assert.ErrorIs(t, err, ErrUnknownTypeForConversationElement)
}

func TestFilesystemConversation_AppendEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conv.jsonl")
	s := NewFilesystemConversation(path)
	ctx := context.Background()
	assert.NoError(t, s.Append(ctx, nil))
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err))
	msgs, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestFilesystemConversation_AppendCreatesParentDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "conv.jsonl")
	s := NewFilesystemConversation(path)
	ctx := context.Background()
	err := s.Append(ctx, []ConversationElement{testUserMsg("hi")})
	assert.NoError(t, err)
	msgs, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Len(t, msgs, 1)
	assert.Equal(t, "hi", msgs[0].(*UserMessage).Content[0].Text)
}
