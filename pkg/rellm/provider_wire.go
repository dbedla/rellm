package rellm

// Wire-format translation helpers for provider implementations.
//
// Every parse*/marshal* helper that converts between the Responses API wire
// format and canonical conversation elements lives here. Functions shared by
// all providers are named by the value they handle (marshalFunctionCall).
// Functions whose wire shape differs between backends are named by that
// shape (marshalMessageAsTypedText vs marshalMessageAsUntypedParts), so a new
// provider picks by behavior rather than by provider brand. Per-type dispatch
// stays in each provider's marshalConversationElement.

import (
	"encoding/json"
	"errors"
)

// --- Parse: wire item -> canonical element ------------------------------------

// wireItem mirrors the output-item fields common to Responses API wire items.
type wireItem struct {
	Type    string          `json:"type"`
	Role    string          `json:"role"`
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	CallID  string          `json:"call_id"`
	Args    json.RawMessage `json:"arguments"`
	Content json.RawMessage `json:"content"`
}

// parseWireItem decodes an output item's common fields.
func parseWireItem(raw json.RawMessage) (wireItem, error) {
	var msg wireItem
	err := json.Unmarshal(raw, &msg)
	return msg, err
}

// parseFunctionCallItem builds a FunctionCall from a function_call item.
func parseFunctionCallItem(msg wireItem) *FunctionCall {
	return &FunctionCall{
		ID:     msg.ID,
		Name:   msg.Name,
		Args:   msg.Args,
		CallID: msg.CallID,
	}
}

// parseFunctionCallRespItem builds a FunctionCallResp from a
// function_call_output item. The output field is extracted directly from the
// raw JSON; on failure the content field is used as-is.
func parseFunctionCallRespItem(raw json.RawMessage, msg wireItem) *FunctionCallResp {
	var out struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(raw, &out); err == nil {
		return &FunctionCallResp{
			ID:     msg.ID,
			Type:   "function_call_output",
			CallID: msg.CallID,
			Output: out.Output,
		}
	}
	return &FunctionCallResp{
		ID:     msg.ID,
		Type:   "function_call_output",
		CallID: msg.CallID,
		Output: string(msg.Content),
	}
}

// parseImageGeneration parses an image_generation_call wire item into an
// ImageGeneration element.
func parseImageGeneration(raw json.RawMessage) (ConversationElement, error) {
	var ig struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Result string `json:"result"`
	}
	if err := json.Unmarshal(raw, &ig); err != nil {
		return nil, errors.Join(ErrImageParsingFailed, err)
	}
	return &ImageGeneration{ID: ig.ID, Status: ig.Status, Result: ig.Result}, nil
}

// parseMessageByRole parses a message item into a role-typed message without
// extracting a status field (LM Studio wire shape: messages carry no status).
func parseMessageByRole(raw json.RawMessage, provider, id, itemType, role string, content json.RawMessage) ConversationElement {
	parts := normalizeMessageParts(ParseMessageContent(content), role)
	switch role {
	case "assistant":
		return &AssistantMessage{MessageContent{ID: id, Role: role, Content: parts}}
	case "system":
		return &SystemMessage{MessageContent{ID: id, Role: role, Content: parts}}
	case "user":
		return &UserMessage{MessageContent{ID: id, Role: role, Content: parts}}
	default:
		return newUnknownElement(provider, itemType, role, raw)
	}
}

// parseMessageWithStatus parses a message item into a role-typed message,
// extracting the status field when present (OpenRouter and OpenAI wire shape).
func parseMessageWithStatus(raw json.RawMessage, provider, id, itemType, role string, content json.RawMessage) ConversationElement {
	var statusInfo struct {
		Status string `json:"status"`
	}
	status := ""
	if err := json.Unmarshal(raw, &statusInfo); err == nil {
		status = statusInfo.Status
	}
	parts := normalizeMessageParts(ParseMessageContent(content), role)
	switch role {
	case "assistant":
		return &AssistantMessage{MessageContent{ID: id, Role: role, Status: status, Content: parts}}
	case "system":
		return &SystemMessage{MessageContent{ID: id, Role: role, Content: parts}}
	case "user":
		return &UserMessage{MessageContent{ID: id, Role: role, Status: status, Content: parts}}
	default:
		return newUnknownElement(provider, itemType, role, raw)
	}
}

// parseReasoningBasic extracts reasoning text and summary from a raw item,
// ignoring unmarshal errors field by field (LM Studio wire shape: no
// signature).
func parseReasoningBasic(raw json.RawMessage) (ConversationElement, error) {
	r := &Reasoning{}

	// Extract metadata and continuation state from the top level.
	var meta struct {
		ID               string                 `json:"id"`
		Status           string                 `json:"status"`
		Summary          []ReasoningSummaryPart `json:"summary"`
		EncryptedContent string                 `json:"encrypted_content"`
		Format           string                 `json:"format"`
	}
	err := json.Unmarshal(raw, &meta)
	if err != nil {
		return nil, errors.Join(ErrReasoningParsingFailed, err)
	}

	r.ID = meta.ID
	r.Status = meta.Status
	r.EncryptedContent = meta.EncryptedContent
	r.Format = meta.Format
	if len(meta.Summary) > 0 {
		r.Summary = meta.Summary
	}

	// Extract text from content array (reasoning_text items)
	var msg struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &msg); err == nil {
		var textParts []string
		for _, c := range msg.Content {
			if c.Type == "reasoning_text" || c.Type == "text" {
				textParts = append(textParts, c.Text)
			}
		}
		r.Text = JoinTextParts(textParts)
	}

	return r, nil
}

// ReasoningContentPart is a content part of a reasoning wire item.
type ReasoningContentPart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// parseReasoningSigned extracts reasoning text, summary, and provider
// continuation state; fails on malformed items (OpenRouter wire shape).
func parseReasoningSigned(raw json.RawMessage) (*Reasoning, error) {
	var item struct {
		ID               string                 `json:"id"`
		Status           string                 `json:"status"`
		Summary          []ReasoningSummaryPart `json:"summary"`
		Content          []ReasoningContentPart `json:"content"`
		Signature        string                 `json:"signature"`
		EncryptedContent string                 `json:"encrypted_content"`
		Format           string                 `json:"format"`
	}
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, errors.Join(ErrReasoningParsingFailed, err) // skip malformed items
	}

	textParts := make([]string, 0, len(item.Content))
	for _, part := range item.Content {
		if part.Type == "reasoning_text" || part.Type == "text" {
			textParts = append(textParts, part.Text)
		}
	}
	if len(item.Summary) == 0 {
		item.Summary = nil
	}

	return &Reasoning{
		ID:               item.ID,
		Status:           item.Status,
		Summary:          item.Summary,
		Text:             JoinTextParts(textParts),
		Signature:        item.Signature,
		EncryptedContent: item.EncryptedContent,
		Format:           item.Format,
	}, nil
}

// ParseMessageContent normalizes a wire content field (string, []string, or
// []MessagePart) into []MessagePart. Exported for external providers that need
// the same content normalization when parsing wire messages.
func ParseMessageContent(content json.RawMessage) []MessagePart {
	var textStr string
	if json.Unmarshal(content, &textStr) == nil {
		return messagePartsWithStrings([]string{textStr})
	}
	var textParts []string
	if json.Unmarshal(content, &textParts) == nil {
		return messagePartsWithStrings(textParts)
	}
	var parts []MessagePart
	if json.Unmarshal(content, &parts) == nil {
		return parts
	}
	return nil
}

// normalizeMessageParts canonicalizes provider text parts independently of
// the shape used on the wire. Providers may replay text as a plain string or
// return it as an input_text/output_text object; the canonical representation
// is determined by the message role.
func normalizeMessageParts(parts []MessagePart, role string) []MessagePart {
	textType := "input_text"
	if role == "assistant" {
		textType = "output_text"
	}

	for i := range parts {
		if parts[i].Type == "input_text" || parts[i].Type == "output_text" {
			parts[i].Type = textType
		}
	}

	return parts
}

// messagePartsWithStrings creates MessagePart slice from a string list.
func messagePartsWithStrings(texts []string) []MessagePart {
	parts := make([]MessagePart, 0, len(texts))
	for _, t := range texts {
		parts = append(parts, MessagePart{Type: "input_text", Text: t})
	}
	return parts
}

// --- Marshal: canonical element -> wire item ----------------------------------

// marshalFunctionCall serializes a FunctionCall for the Responses wire format.
func marshalFunctionCall(el *FunctionCall) (json.RawMessage, error) {
	type payload struct {
		ID        string          `json:"id"`
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
		CallID    string          `json:"call_id"`
		Status    string          `json:"status"`
		Type      string          `json:"type"`
	}
	p := payload{
		ID:        el.ID,
		Name:      el.Name,
		Arguments: el.Args,
		CallID:    el.CallID,
		Status:    "completed",
		Type:      "function_call",
	}
	return json.Marshal(p)
}

// marshalFunctionCallResp serializes a FunctionCallResp for the Responses wire
// format.
func marshalFunctionCallResp(el *FunctionCallResp) (json.RawMessage, error) {
	type payload struct {
		ID     string `json:"id,omitempty"`
		CallID string `json:"call_id"`
		Type   string `json:"type"`
		Output string `json:"output"`
	}
	p := payload{
		ID:     el.ID,
		CallID: el.CallID,
		Type:   "function_call_output",
		Output: el.Output,
	}
	return json.Marshal(p)
}

// marshalImageGeneration serializes an ImageGeneration for the Responses wire
// format.
func marshalImageGeneration(el *ImageGeneration) (json.RawMessage, error) {
	type payload struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		Status string `json:"status"`
		Result string `json:"result"`
	}
	p := payload{
		ID:     el.ID,
		Type:   "image_generation_call",
		Status: el.Status,
		Result: el.Result,
	}
	return json.Marshal(p)
}

// marshalMessageAsUntypedParts serializes a role-typed message whose content
// is always a structured array and whose wire item carries no "type" field
// (LM Studio shape).
func marshalMessageAsUntypedParts(mc MessageContent) (json.RawMessage, error) {
	type payload struct {
		ID      string        `json:"id,omitempty"`
		Role    string        `json:"role"`
		Status  string        `json:"status,omitempty"`
		Content []MessagePart `json:"content"`
	}
	content := mc.Content
	if len(content) == 0 {
		content = []MessagePart{}
	}
	p := payload{
		ID:      mc.ID,
		Role:    mc.Role,
		Status:  mc.Status,
		Content: content,
	}
	return json.Marshal(p)
}

// marshalMessageAsTypedParts serializes a role-typed message with a
// "type":"message" field whose content is always a structured array
// (OpenRouter user-message shape).
func marshalMessageAsTypedParts(mc MessageContent) (json.RawMessage, error) {
	type payload struct {
		ID      string        `json:"id,omitempty"`
		Role    string        `json:"role"`
		Type    string        `json:"type"`
		Status  string        `json:"status,omitempty"`
		Content []MessagePart `json:"content"`
	}
	content := mc.Content
	if len(content) == 0 {
		content = []MessagePart{}
	}
	p := payload{
		ID:      mc.ID,
		Role:    mc.Role,
		Type:    "message",
		Status:  mc.Status,
		Content: content,
	}
	return json.Marshal(p)
}

// marshalMessageAsTypedText serializes a role-typed message with a
// "type":"message" field whose text-only content becomes a plain string
// (matching the observed wire format, which stringifies output_text parts
// too); multimodal content stays an array (OpenRouter assistant/system
// shape).
func marshalMessageAsTypedText(mc MessageContent) (json.RawMessage, error) {
	type payload struct {
		ID      string `json:"id,omitempty"`
		Role    string `json:"role"`
		Type    string `json:"type"`
		Status  string `json:"status,omitempty"`
		Content any    `json:"content"`
	}
	var content any
	if len(mc.Content) == 0 {
		content = ""
	} else if hasImageParts(mc.Content) {
		content = mc.Content
	} else {
		content = TextFromContent(mc.Content)
	}
	p := payload{
		ID:      mc.ID,
		Role:    mc.Role,
		Type:    "message",
		Status:  mc.Status,
		Content: content,
	}
	return json.Marshal(p)
}

// hasImageParts reports whether any part carries an image (multimodal content
// that must stay a structured array).
func hasImageParts(parts []MessagePart) bool {
	for _, p := range parts {
		if p.ImageURL != nil {
			return true
		}
	}
	return false
}

// marshalReasoningWithSignature serializes a Reasoning item carrying provider
// continuation state; returns (nil, nil) for an item with no
// user-visible content and no continuation state, which the caller skips
// (OpenRouter shape).
func marshalReasoningWithSignature(el *Reasoning) (json.RawMessage, error) {
	// A signature-only or encrypted-only item carries provider continuation state
	// and must be replayed even though it has no user-visible text.
	if el.Text == "" && el.Signature == "" && el.EncryptedContent == "" && len(el.Summary) == 0 {
		return nil, nil
	}
	type payload struct {
		ID               string                  `json:"id"`
		Type             string                  `json:"type"`
		Status           string                  `json:"status,omitempty"`
		Summary          *[]ReasoningSummaryPart `json:"summary,omitempty"`
		Content          []MessagePart           `json:"content,omitempty"`
		Signature        string                  `json:"signature,omitempty"`
		EncryptedContent string                  `json:"encrypted_content,omitempty"`
		Format           string                  `json:"format,omitempty"`
	}
	p := payload{
		ID:               el.ID,
		Type:             "reasoning",
		Status:           el.Status,
		Signature:        el.Signature,
		EncryptedContent: el.EncryptedContent,
		Format:           el.Format,
	}
	// Signed and encrypted reasoning blocks are provider continuation state; preserve
	// their summary field even when it is empty.
	if len(el.Summary) > 0 {
		p.Summary = &el.Summary
	} else if el.Signature != "" || el.EncryptedContent != "" {
		emptySummary := []ReasoningSummaryPart{}
		p.Summary = &emptySummary
	}
	if el.Text != "" {
		p.Content = []MessagePart{{Type: "reasoning_text", Text: el.Text}}
	}
	return json.Marshal(p)
}

// marshalReasoningWithSummary serializes a Reasoning item, always including
// the summary field (even when empty) for faithful round-trip; no signature
// support (LM Studio shape).
func marshalReasoningWithSummary(el *Reasoning) (json.RawMessage, error) {
	type payload struct {
		ID               string                 `json:"id"`
		Type             string                 `json:"type"`
		Status           string                 `json:"status,omitempty"`
		Summary          []ReasoningSummaryPart `json:"summary"`
		Content          []MessagePart          `json:"content,omitempty"`
		EncryptedContent string                 `json:"encrypted_content,omitempty"`
		Format           string                 `json:"format,omitempty"`
	}
	summary := el.Summary
	if len(summary) == 0 {
		summary = []ReasoningSummaryPart{}
	}
	p := payload{
		ID:               el.ID,
		Type:             "reasoning",
		Status:           el.Status,
		Summary:          summary,
		EncryptedContent: el.EncryptedContent,
		Format:           el.Format,
	}
	if el.Text != "" {
		p.Content = []MessagePart{{Type: "reasoning_text", Text: el.Text}}
	}
	return json.Marshal(p)
}
