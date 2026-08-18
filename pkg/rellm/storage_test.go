package rellm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage_LoadEmpty(t *testing.T) {
	s := NewInMemoryStorage()
	msgs, err := s.Load()
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestInMemoryStorage_AppendAndLoad(t *testing.T) {
	s := NewInMemoryStorage()
	delta := []json.RawMessage{
		json.RawMessage(`{"role":"user","content":"hi"}`),
		json.RawMessage(`{"role":"assistant","content":"hello"}`),
	}
	err := s.Append(delta)
	assert.NoError(t, err)
	msgs, err := s.Load()
	assert.NoError(t, err)
	assert.Equal(t, delta, msgs)
}

func TestInMemoryStorage_AppendMultiple(t *testing.T) {
	s := NewInMemoryStorage()
	first := []json.RawMessage{json.RawMessage(`{"role":"user","content":"a"}`)}
	second := []json.RawMessage{json.RawMessage(`{"role":"assistant","content":"b"}`)}
	assert.NoError(t, s.Append(first))
	assert.NoError(t, s.Append(second))
	msgs, err := s.Load()
	assert.NoError(t, err)
	assert.Equal(t, append(first, second...), msgs)
}

func TestInMemoryStorage_LoadDoesNotAlias(t *testing.T) {
	s := NewInMemoryStorage()
	assert.NoError(t, s.Append([]json.RawMessage{json.RawMessage(`{"role":"user","content":"x"}`)}))
	loaded, err := s.Load()
	assert.NoError(t, err)
	assert.Len(t, loaded, 1)
	loaded = append(loaded, json.RawMessage(`{"role":"assistant","content":"y"}`))
	again, err := s.Load()
	assert.Len(t, loaded, 2)
	assert.NoError(t, err)
	assert.Len(t, again, 1)
}

func TestInMemoryStorage_AppendEmpty(t *testing.T) {
	s := NewInMemoryStorage()
	assert.NoError(t, s.Append(nil))
	msgs, err := s.Load()
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestFilesystemStorage_LoadMissingFile(t *testing.T) {
	s := NewFilesystemStorage(filepath.Join(t.TempDir(), "missing.jsonl"))
	msgs, err := s.Load()
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestFilesystemStorage_AppendAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conv.jsonl")
	s := NewFilesystemStorage(path)
	delta := []json.RawMessage{
		json.RawMessage(`{"role":"user","content":"hi"}`),
		json.RawMessage(`{"role":"assistant","content":"hello"}`),
	}
	err := s.Append(delta)
	assert.NoError(t, err)
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	expected := `{"role":"user","content":"hi"}` + "\n" + `{"role":"assistant","content":"hello"}` + "\n"
	assert.Equal(t, expected, string(content))
	msgs, err := s.Load()
	assert.NoError(t, err)
	assert.Equal(t, delta, msgs)
}

func TestFilesystemStorage_AppendMultiple(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conv.jsonl")
	s := NewFilesystemStorage(path)
	first := []json.RawMessage{json.RawMessage(`{"role":"user","content":"a"}`)}
	second := []json.RawMessage{json.RawMessage(`{"role":"assistant","content":"b"}`)}
	assert.NoError(t, s.Append(first))
	assert.NoError(t, s.Append(second))
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	expected := `{"role":"user","content":"a"}` + "\n" + `{"role":"assistant","content":"b"}` + "\n"
	assert.Equal(t, expected, string(content))
	msgs, err := s.Load()
	assert.NoError(t, err)
	assert.Equal(t, append(first, second...), msgs)
}

func TestFilesystemStorage_LoadSkipsMalformedLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conv.jsonl")
	content := `{"role":"user","content":"hi"}` + "\n" + `{broken` + "\n"
	assert.NoError(t, os.WriteFile(path, []byte(content), 0644))
	s := NewFilesystemStorage(path)
	msgs, err := s.Load()
	assert.NoError(t, err)
	assert.Len(t, msgs, 1)
	assert.Equal(t, json.RawMessage(`{"role":"user","content":"hi"}`), msgs[0])
}

func TestFilesystemStorage_AppendEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conv.jsonl")
	s := NewFilesystemStorage(path)
	assert.NoError(t, s.Append(nil))
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err))
	msgs, err := s.Load()
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestFilesystemStorage_AppendMultilineJSONCompacts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conv.jsonl")
	s := NewFilesystemStorage(path)
	delta := []json.RawMessage{
		json.RawMessage("{\n  \"role\": \"user\",\n  \"content\": \"hello world\"\n}"),
	}
	err := s.Append(delta)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	expected := `{"role":"user","content":"hello world"}` + "\n"
	assert.Equal(t, expected, string(content))

	msgs, err := s.Load()
	assert.NoError(t, err)
	assert.Len(t, msgs, 1)
	assert.JSONEq(t, `{"role":"user","content":"hello world"}`, string(msgs[0]))
}

func TestFilesystemStorage_AppendCreatesParentDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "conv.jsonl")
	s := NewFilesystemStorage(path)
	delta := []json.RawMessage{json.RawMessage(`{"role":"user","content":"hi"}`)}
	err := s.Append(delta)
	assert.NoError(t, err)
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, `{"role":"user","content":"hi"}`+"\n", string(content))
	msgs, err := s.Load()
	assert.NoError(t, err)
	assert.Equal(t, delta, msgs)
}
