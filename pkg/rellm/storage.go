package rellm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type InMemoryStorage struct {
	Messages []json.RawMessage
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{}
}

func (s *InMemoryStorage) Load() ([]json.RawMessage, error) {
	cp := make([]json.RawMessage, len(s.Messages))
	copy(cp, s.Messages)
	return cp, nil
}

func (s *InMemoryStorage) Append(delta []json.RawMessage) error {
	s.Messages = append(s.Messages, delta...)
	return nil
}

type FilesystemStorage struct {
	path string
}

func NewFilesystemStorage(path string) *FilesystemStorage {
	return &FilesystemStorage{path: path}
}

func (s *FilesystemStorage) Load() ([]json.RawMessage, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []json.RawMessage{}, nil
		}
		return nil, err
	}
	messages := []json.RawMessage{}
	for lineNum, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if !json.Valid(line) {
			return nil, errors.Join(ErrMalformedConversationStorage, fmt.Errorf("file: %s, line: %d", s.path, lineNum+1))
		}
		messages = append(messages, json.RawMessage(line))
	}
	return messages, nil
}

func (s *FilesystemStorage) Append(delta []json.RawMessage) (err error) {
	if len(delta) == 0 {
		return nil
	}
	if dir := filepath.Dir(s.path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	var buf bytes.Buffer
	for _, m := range delta {
		if err := json.Compact(&buf, m); err != nil {
			return err
		}
		buf.WriteByte('\n')
	}
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer closeWithError(&err, f)
	_, err = f.Write(buf.Bytes())
	return err
}
