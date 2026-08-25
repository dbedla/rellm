package rellm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (a *Agent) run(msg string, params promptParams) (string, error) {
	conversation, err := a.CurrentConversation()
	if err != nil {
		return "", err
	}

	userMsg, err := PromptMessageToConversation(msg, "user")
	if err != nil {
		return "", errors.Join(ErrUserMsgConversionFailed, err)
	}

	conversation = append(conversation, userMsg)
	if err := a.conversationStorage.Append([]json.RawMessage{userMsg}); err != nil {
		return "", err
	}

	elements, err := a.provider.ToConversationElements(conversation)
	if err != nil {
		return "", errors.Join(ErrConversationElementConversion, err)
	}
	wire, err := a.provider.ToProviderRepresentation(elements)
	if err != nil {
		return "", errors.Join(ErrConversationElementConversion, err)
	}

	req := toBaseResponsesApiReq(params, a.provider.Model(), wire)

	if a.toolset != nil {
		req.Tools = a.toolset.BuildTools()
	}

	msgRespFromLLM, err := a.process(req)

	if err != nil {
		return "", err
	}

	return msgRespFromLLM, nil
}

func (a *Agent) post(req *ResponsesApiReq) (_ *ResponsesApiResp, err error) {
	if a.inspectReq != nil {
		a.inspectReq(req)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	apiUrl, err := url.Parse(a.provider.URL())
	if err != nil {
		return nil, err
	}

	httpReq := &http.Request{
		Method: "POST",
		Header: a.provider.Header(),
		URL:    apiUrl,
		Body:   io.NopCloser(bytes.NewReader(body)),
	}

	resp, err := a.provider.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, ErrEndpointNilResponse
	}

	if resp.Body == nil {
		return nil, errors.Join(ErrEndpointNilBodyInResponse, fmt.Errorf("response status: %s", resp.Status))
	}

	defer closeWithError(&err, resp.Body)
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(err, ErrUnableToReadResponseBody, fmt.Errorf("response status: %s", resp.Status))
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, newHTTPStatusError(resp, rawBody, apiUrl.String())
	}

	return parseResponsesApiResponse(rawBody, a.inspectResp)
}

func (a *Agent) process(req *ResponsesApiReq) (string, error) {
	for i := uint64(0); i < a.maxToolsIterationWithoutReturnMessage; i++ {
		conversationResponse, err := a.post(req)
		if err != nil {
			return "", err
		}

		if conversationResponse.Error != nil {
			return "", errors.Join(ErrInConversationResponse, fmt.Errorf("err msg: %v", conversationResponse.Error.Message))
		}

		conversation, err := a.provider.ToConversationElements(conversationResponse.Output)
		if err != nil {
			return "", errors.Join(ErrConversationElementConversion, err)
		}
		msg, fnCallsResp, imageHandled, err := a.dispatchConversation(conversation)
		for _, fResp := range fnCallsResp {
			conversation = append(conversation, fResp)
		}
		raw, errProviderRep := a.provider.ToProviderRepresentation(conversation)
		if errProviderRep != nil {
			return "", errors.Join(ErrConversationElementConversion, errProviderRep)
		}
		req.Input = append(req.Input, raw...)

		conversationErr := a.conversationStorage.Append(raw)
		if conversationErr != nil {
			return "", conversationErr
		}

		if err != nil {
			return "", err
		}

		if imageHandled || msg != "" {
			return msg, nil
		}

	}

	return "",
		errors.Join(ErrMaxToolIterationsReached,
			fmt.Errorf("exceeded %d iterations", a.maxToolsIterationWithoutReturnMessage))
}

func (a *Agent) dispatchConversation(conversation []ConversationElement) (string, []*FunctionCallResp, bool, error) {

	var message strings.Builder
	fnCallsResults := []*FunctionCallResp{}
	imageHandled := false

	var outputErr error

	for _, conversationElement := range conversation {
		switch el := conversationElement.(type) {
		case *SystemMessage:
			continue
		case *UserMessage:
			continue
		case *FunctionCallResp:
			continue
		case *Reasoning:
			continue

		case *FunctionCall:
			fResp, err := a.handleFunctionCall(el)
			if err != nil {
				outputErr = errors.Join(outputErr, err)
			}
			if fResp != nil {
				fnCallsResults = append(fnCallsResults, fResp)
			}

		case *AssistantMessage:
			message.WriteString(messagesFromParts(el.Content))
			continue

		case *ImageGeneration:
			err := a.handleImageGenerationCall(el)
			if err != nil {
				outputErr = errors.Join(outputErr, err)
			}
			imageHandled = true
			continue

		default:
			return "", nil, false, ErrUnknownConversationElement
		}
	}

	return message.String(), fnCallsResults, imageHandled, outputErr
}

func messagesFromParts(parts []MessagePart) string {
	var msg strings.Builder
	for _, part := range parts {
		msg.WriteString(part.Text)
	}
	return msg.String()
}

func (a *Agent) handleImageGenerationCall(image *ImageGeneration) error {
	if a.handleImageGeneration == nil {
		return ErrNoImageHandler
	}

	resultNote, err := a.handleImageGeneration(image)
	if err != nil {
		return errors.Join(ErrCustomImageHandlerFailed, err)
	}

	image.Result = resultNote

	return nil
}

func (a *Agent) handleFunctionCall(fn *FunctionCall) (*FunctionCallResp, error) {
	if a.toolset == nil {
		return nil, ErrNoToolsetButToolCallRequested
	}

	funcCallResp, ok := a.toolset.DispatchTools(fn.Name, fn.CallId, fn.Args)
	if !ok {
		funcCallResp = invalidFunctionCallResp(fn)
		return &funcCallResp, errors.Join(ErrWhileDispatchToolCall, fmt.Errorf("unknown tool name (%s)", fn.Name))
	}

	return &funcCallResp, nil
}

func invalidFunctionCallResp(fn *FunctionCall) FunctionCallResp {
	return FunctionCallResp{
		Type:   "function_call_output",
		CallId: fn.CallId,
		Output: "invalid function call (function not found) " + fn.Name,
	}

}

func toBaseResponsesApiReq(params promptParams, model Model, conversation []json.RawMessage) *ResponsesApiReq {
	return &ResponsesApiReq{
		Model:            string(model),
		Input:            conversation,
		Temperature:      params.Temperature,
		Reasoning:        params.Reasoning,
		MaxOutputTokens:  params.MaxOutputTokens,
		TopP:             params.TopP,
		PresencePenalty:  params.PresencePenalty,
		FrequencyPenalty: params.FrequencyPenalty,
		Seed:             params.Seed,
		Logprobs:         params.Logprobs,
		TopLogprobs:      params.TopLogprobs,
	}
}
