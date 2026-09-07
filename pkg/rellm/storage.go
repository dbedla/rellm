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

// InMemoryStorage is a storage implementation that stores conversation elements in memory.
// minimal and basic, no concurrency support
type InMemoryStorage struct {
	messages []ConversationElement
}

// NewInMemoryStorage creates a new instance of InMemoryStorage.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{}
}

// Load loads conversation elements from memory. Context ignored
// elements are copy of the internal storage
func (s *InMemoryStorage) Load(_ context.Context) ([]ConversationElement, error) {
	cp := make([]ConversationElement, len(s.messages))
	copy(cp, s.messages)
	return cp, nil
}

// Append appends conversation elements to memory. Context ignored
func (s *InMemoryStorage) Append(_ context.Context, delta []ConversationElement) error {
	s.messages = append(s.messages, delta...)
	return nil
}

// FilesystemStorage provides a storage implementation that reads and writes conversation elements to a file.
// no concurrency support
// basic implementation with data persistance, file corruption during Load or Append makes further work impossible
type FilesystemStorage struct {
	path string
}

// NewFilesystemStorage creates a new FilesystemStorage instance.
// path to file where data are or will be stored
func NewFilesystemStorage(path string) *FilesystemStorage {
	return &FilesystemStorage{path: path}
}

// Load returns conversation elements from file
// Context ignored
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

// Append append new conversation elements to file, context ignored
// if no file exists, it will be created
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
