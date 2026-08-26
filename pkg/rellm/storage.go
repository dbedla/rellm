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
	Messages []ConversationElement
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{}
}

func (s *InMemoryStorage) Load() ([]ConversationElement, error) {
	cp := make([]ConversationElement, len(s.Messages))
	copy(cp, s.Messages)
	return cp, nil
}

func (s *InMemoryStorage) Append(delta []ConversationElement) error {
	s.Messages = append(s.Messages, delta...)
	return nil
}

type FilesystemStorage struct {
	path string
}

func NewFilesystemStorage(path string) *FilesystemStorage {
	return &FilesystemStorage{path: path}
}

func (s *FilesystemStorage) Load() ([]ConversationElement, error) {
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

func (s *FilesystemStorage) Append(delta []ConversationElement) (err error) {
	if len(delta) == 0 {
		return nil
	}
	if dir := filepath.Dir(s.path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	var buf bytes.Buffer
	for _, el := range delta {
		b, err := json.Marshal(el)
		if err != nil {
			return err
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, b); err != nil {
			return err
		}
		buf.Write(compact.Bytes())
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
