package rellm

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type openRouterTextMessage struct {
	Type    string `json:"type"`
	Role    string `json:"role"`
	Id      string `json:"id,omitempty"`
	Status  string `json:"status,omitempty"`
	Content string `json:"content"`
}

type openRouterStructuredMessage struct {
	Type    string        `json:"type"`
	Role    string        `json:"role"`
	Id      string        `json:"id,omitempty"`
	Status  string        `json:"status,omitempty"`
	Content []MessagePart `json:"content"`
}

type genericConversationMessage struct {
	Type    string          `json:"type,omitempty"`
	Role    string          `json:"role,omitempty"`
	Id      string          `json:"id,omitempty"`
	Status  string          `json:"status,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
}

func normalizeConversationForProvider(provider Provider, conversation []json.RawMessage) ([]json.RawMessage, error) {
	switch provider {
	case Provider_OpenRouter:
		normalized := make([]json.RawMessage, 0, len(conversation))
		for _, raw := range conversation {
			item, err := normalizeOpenRouterConversationItem(raw)
			if err != nil {
				return nil, err
			}
			if len(item) == 0 {
				continue
			}
			normalized = append(normalized, item)
		}

		return normalized, nil
	case Provider_LMStudio:
		fallthrough
	default:
		return conversation, nil
	}
}

func normalizeOpenRouterConversationItem(raw json.RawMessage) (json.RawMessage, error) {
	var msg genericConversationMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}

	switch msg.Role {
	case "system":
		text, err := rawContentToText(msg.Content)
		if err != nil {
			return nil, err
		}
		return json.Marshal(openRouterTextMessage{
			Type:    "message",
			Role:    msg.Role,
			Id:      msg.Id,
			Status:  msg.Status,
			Content: text,
		})
	case "assistant":
		text, err := rawContentToText(msg.Content)
		if err != nil {
			return nil, err
		}
		return json.Marshal(openRouterTextMessage{
			Type:    "message",
			Role:    msg.Role,
			Id:      msg.Id,
			Status:  msg.Status,
			Content: text,
		})
	case "user":
		parts, err := rawContentToParts(msg.Content)
		if err != nil {
			return nil, err
		}
		return json.Marshal(openRouterStructuredMessage{
			Type:    "message",
			Role:    msg.Role,
			Id:      msg.Id,
			Status:  msg.Status,
			Content: parts,
		})
	default:
		if msg.Type == "reasoning" {
			return normalizeOpenRouterReasoningItem(raw)
		}
		return raw, nil
	}
}

func rawContentToText(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return "", nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}

	parts, err := rawContentToParts(raw)
	if err != nil {
		return "", err
	}

	return messagePartsToText(parts), nil
}

func rawContentToParts(raw json.RawMessage) ([]MessagePart, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}

	var parts []MessagePart
	if err := json.Unmarshal(raw, &parts); err != nil {
		return nil, fmt.Errorf("normalize content parts: %w", err)
	}

	return parts, nil
}

func messagePartsToText(parts []MessagePart) string {
	var text string
	for _, part := range parts {
		if part.Text == "" {
			continue
		}
		if text != "" {
			text += " "
		}
		text += part.Text
	}
	return text
}
