package conversation_storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type ConversationStorage struct {
	dir       string
	agentName string
}

func New(dir string) *ConversationStorage {
	return &ConversationStorage{dir: dir, agentName: ""}
}

func NewForAgent(dir string, agentName string) *ConversationStorage {
	return &ConversationStorage{dir: dir, agentName: agentName}
}

func (s *ConversationStorage) Store(messages []json.RawMessage) error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	filename := time.Now().Format("2006-01-02-15-04-05") + "-" + s.agentName + "-conversation.json"
	path := filepath.Join(s.dir, filename)

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write conversation file: %w", err)
	}

	return nil
}

func (s *ConversationStorage) RestoreLatest() ([]json.RawMessage, error) {
	files, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []json.RawMessage{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var conversationFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), "-conversation.json") {
			conversationFiles = append(conversationFiles, f.Name())
		}
	}

	if len(conversationFiles) == 0 {
		return []json.RawMessage{}, nil
	}

	slices.Sort(conversationFiles)
	latest := conversationFiles[len(conversationFiles)-1]

	return s.RestoreSpecific(latest)
}

func (s *ConversationStorage) RestoreSpecific(filename string) ([]json.RawMessage, error) {
	path := filepath.Join(s.dir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []json.RawMessage{}, nil
		}
		return nil, fmt.Errorf("failed to read conversation file: %w", err)
	}

	var messages []json.RawMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation: %w", err)
	}

	return messages, nil
}
