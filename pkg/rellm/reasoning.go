package rellm

import "encoding/json"

type ReasoningContentPart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type ReasoningConversationItem struct {
	Id               string                 `json:"id,omitempty"`
	Type             string                 `json:"type"`
	Status           string                 `json:"status,omitempty"`
	Summary          []json.RawMessage      `json:"summary,omitempty"`
	Content          []ReasoningContentPart `json:"content,omitempty"`
	Signature        string                 `json:"signature,omitempty"`
	EncryptedContent string                 `json:"encrypted_content,omitempty"`
	Format           string                 `json:"format,omitempty"`
}

func reasoningConversationElements(raw json.RawMessage, provider Provider) ([]json.RawMessage, error) {
	switch provider {
	case Provider_OpenRouter:
		var reasoning ReasoningConversationItem
		if err := json.Unmarshal(raw, &reasoning); err != nil {
			return nil, err
		}

		if len(reasoning.Content) == 0 {
			return nil, nil
		}

		reasoning.Signature = ""
		reasoning.EncryptedContent = ""
		reasoning.Format = ""

		sanitized, err := json.Marshal(reasoning)
		if err != nil {
			return nil, err
		}

		return []json.RawMessage{sanitized}, nil
	case Provider_LMStudio:
		fallthrough
	default:
		return []json.RawMessage{raw}, nil
	}
}

func normalizeOpenRouterReasoningItem(raw json.RawMessage) (json.RawMessage, error) {
	var reasoning ReasoningConversationItem
	if err := json.Unmarshal(raw, &reasoning); err != nil {
		return nil, err
	}

	if len(reasoning.Content) == 0 {
		return nil, nil
	}

	reasoning.Signature = ""
	reasoning.EncryptedContent = ""
	reasoning.Format = ""

	return json.Marshal(reasoning)
}
