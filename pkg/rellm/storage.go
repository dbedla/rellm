package rellm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// InMemoryStorage stores conversation elements in memory. Not safe for
// concurrent use; data is lost when the process exits.
type InMemoryStorage struct {
	messages []ConversationElement
}

// NewInMemoryStorage creates a new instance of InMemoryStorage.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{}
}

// Load returns a copy of the stored conversation elements. ctx is ignored.
func (s *InMemoryStorage) Load(_ context.Context) ([]ConversationElement, error) {
	cp := make([]ConversationElement, len(s.messages))
	copy(cp, s.messages)
	return cp, nil
}

// Append adds elements to the in-memory history. ctx is ignored.
func (s *InMemoryStorage) Append(_ context.Context, delta []ConversationElement) error {
	s.messages = append(s.messages, delta...)
	return nil
}

// FilesystemStorage persists conversation elements as JSON lines in a file.
// Not safe for concurrent use. A corrupted line makes Load fail.
type FilesystemStorage struct {
	path string
}

// NewFilesystemStorage returns a storage using the file at path.
func NewFilesystemStorage(path string) *FilesystemStorage {
	return &FilesystemStorage{path: path}
}

// Load reads conversation elements from the file. A missing file yields an
// empty conversation. ctx is ignored.
func (s *FilesystemStorage) Load(_ context.Context) ([]ConversationElement, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []ConversationElement{}, nil
		}
		return nil, err
	}
	elements := []ConversationElement{}
	for lineNum, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if !json.Valid(line) {
			return nil, errors.Join(ErrMalformedConversationStorage, fmt.Errorf("file: %s, line: %d", s.path, lineNum+1))
		}
		el, err := ParseConversationElement(line)
		if err != nil {
			return nil, errors.Join(ErrMalformedConversationStorage, fmt.Errorf("file: %s, line: %d", s.path, lineNum+1), err)
		}
		elements = append(elements, el)
	}
	return elements, nil
}

// Append writes the given elements to the file, one JSON object per line,
// creating the file and parent directories if missing. ctx is ignored.
func (s *FilesystemStorage) Append(_ context.Context, delta []ConversationElement) (finalErr error) {
	if len(delta) == 0 {
		return nil
	}

	// 1. Serialize and validate the entire batch in memory first
	var buf bytes.Buffer
	for i, el := range delta {
		b, err := json.Marshal(el)
		if err != nil {
			return fmt.Errorf("failed to marshal conversation element [%d]: %w", i, err)
		}

		buf.Write(b)
		buf.WriteByte('\n')
	}

	// 2. Ensure the parent directory exists
	if dir := filepath.Dir(s.path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create storage dir %s: %w", dir, err)
		}
	}

	// 3. Write buffered batch to disk
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open storage file %s: %w", s.path, err)
	}
	defer closeWithError(&finalErr, f)

	if _, err = f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write to storage file %s: %w", s.path, err)
	}
	return nil
}
