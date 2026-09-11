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
// extracting the status field when present (OpenRouter wire shape).
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
func parseReasoningSigned(raw json.RawMessage) (ConversationElement, error) {
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
	fc := map[string]interface{}{
		"id":        el.ID,
		"name":      el.Name,
		"arguments": json.RawMessage(el.Args),
		"call_id":   el.CallID,
		"status":    "completed",
		"type":      "function_call",
	}
	return json.Marshal(fc)
}

// marshalFunctionCallResp serializes a FunctionCallResp for the Responses wire
// format.
func marshalFunctionCallResp(el *FunctionCallResp) (json.RawMessage, error) {
	resp := map[string]interface{}{
		"call_id": el.CallID,
		"type":    "function_call_output",
	}
	if el.ID != "" {
		resp["id"] = el.ID
	}
	// Use Output which may be JSON-encoded or plain text.
	resp["output"] = el.Output
	return json.Marshal(resp)
}

// marshalImageGeneration serializes an ImageGeneration for the Responses wire
// format.
func marshalImageGeneration(el *ImageGeneration) (json.RawMessage, error) {
	sig := map[string]interface{}{
		"id":     el.ID,
		"type":   "image_generation_call",
		"status": el.Status,
		"result": el.Result,
	}
	return json.Marshal(sig)
}

// marshalMessageAsUntypedParts serializes a role-typed message whose content
// is always a structured array and whose wire item carries no "type" field
// (LM Studio shape).
func marshalMessageAsUntypedParts(mc MessageContent) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"role": mc.Role,
	}
	if mc.ID != "" {
		payload["id"] = mc.ID
	}
	if mc.Status != "" {
		payload["status"] = mc.Status
	}
	if len(mc.Content) == 0 {
		payload["content"] = []MessagePart{}
	} else {
		payload["content"] = mc.Content
	}
	return json.Marshal(payload)
}

// marshalMessageAsTypedParts serializes a role-typed message with a
// "type":"message" field whose content is always a structured array
// (OpenRouter user-message shape).
func marshalMessageAsTypedParts(mc MessageContent) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"role": mc.Role,
		"type": "message",
	}
	if mc.ID != "" {
		payload["id"] = mc.ID
	}
	if mc.Status != "" {
		payload["status"] = mc.Status
	}
	if len(mc.Content) == 0 {
		payload["content"] = []MessagePart{}
	} else {
		payload["content"] = mc.Content
	}
	return json.Marshal(payload)
}

// marshalMessageAsTypedText serializes a role-typed message with a
// "type":"message" field whose text-only content becomes a plain string
// (matching the observed wire format, which stringifies output_text parts
// too); multimodal content stays an array (OpenRouter assistant/system
// shape).
func marshalMessageAsTypedText(mc MessageContent) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"role": mc.Role,
		"type": "message",
	}
	if mc.ID != "" {
		payload["id"] = mc.ID
	}
	if mc.Status != "" {
		payload["status"] = mc.Status
	}
	if len(mc.Content) == 0 {
		payload["content"] = ""
	} else if hasImageParts(mc.Content) {
		payload["content"] = mc.Content
	} else {
		payload["content"] = TextFromContent(mc.Content)
	}
	return json.Marshal(payload)
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
	r := map[string]interface{}{
		"id":     el.ID,
		"status": el.Status,
		"type":   "reasoning",
	}
	// Signed and encrypted reasoning blocks are provider continuation state; preserve
	// their summary field even when it is empty.
	if len(el.Summary) > 0 {
		r["summary"] = el.Summary
	} else if el.Signature != "" || el.EncryptedContent != "" {
		r["summary"] = []ReasoningSummaryPart{}
	}
	if el.Text != "" {
		r["content"] = []MessagePart{{Type: "reasoning_text", Text: el.Text}}
	}
	if el.Signature != "" {
		r["signature"] = el.Signature
	}
	if el.EncryptedContent != "" {
		r["encrypted_content"] = el.EncryptedContent
	}
	if el.Format != "" {
		r["format"] = el.Format
	}
	return json.Marshal(r)
}

// marshalReasoningWithSummary serializes a Reasoning item, always including
// the summary field (even when empty) for faithful round-trip; no signature
// support (LM Studio shape).
func marshalReasoningWithSummary(el *Reasoning) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"id":     el.ID,
		"status": el.Status,
		"type":   "reasoning",
	}
	payload["summary"] = []ReasoningSummaryPart{}
	if len(el.Summary) > 0 {
		payload["summary"] = el.Summary
	}
	if el.Text != "" {
		payload["content"] = []MessagePart{{Type: "reasoning_text", Text: el.Text}}
	}
	if el.EncryptedContent != "" {
		payload["encrypted_content"] = el.EncryptedContent
	}
	if el.Format != "" {
		payload["format"] = el.Format
	}
	return json.Marshal(payload)
}
