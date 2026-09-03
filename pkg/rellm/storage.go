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

type InMemoryStorage struct {
	Messages []ConversationElement
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{}
}

func (s *InMemoryStorage) Load(_ context.Context) ([]ConversationElement, error) {
	cp := make([]ConversationElement, len(s.Messages))
	copy(cp, s.Messages)
	return cp, nil
}

func (s *InMemoryStorage) Append(_ context.Context, delta []ConversationElement) error {
	s.Messages = append(s.Messages, delta...)
	return nil
}

type FilesystemStorage struct {
	path string
}

func NewFilesystemStorage(path string) *FilesystemStorage {
	return &FilesystemStorage{path: path}
}

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
