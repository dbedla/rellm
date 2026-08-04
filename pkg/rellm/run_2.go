package rellm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (a *Agent) run_newFlow(msg string, params promptParams) (string, error) {

	if a.endpoint.provider == Provider_LMStudio {
		a.llmProvider = LLMProvider{
			cc:       &LMSConversationConverter{},
			endpoint: a.endpoint,
		}
	}
	if a.endpoint.provider == Provider_OpenRouter {
		a.llmProvider = LLMProvider{
			cc:       &OpenRouterConversationConverter{},
			endpoint: a.endpoint,
		}
	}

	a.logger.Info().Msgf("question to agent: %s", string(msg))
	defer a.logger.Info().Msg("question answered")

	conversation := a.CurrentConversation()

	userMsg, err := PromptMessageToConversation(msg, "user")
	if err != nil {
		return "", errors.Join(ErrUserMsgConversionFailed, err)
	}

	conversation = append(conversation, userMsg)

	req := toBaseResponsesApiReq(params, a.endpoint.model, a.endpoint.provider, conversation)

	if a.toolset != nil {
		req.Tools = a.toolset.BuildTools()
	}

	newConversation, msgRespFromLLM, _, err := a.process_newFlow(req)

	if len(newConversation) > 0 {
		a.inMemoryConversation = newConversation
	}

	if err != nil {
		a.logger.Error().Err(err).Msgf("unable to process conversation %s", err.Error())
		return "", err
	}

	a.logger.Info().Msgf("message: %s", msgRespFromLLM)
	return msgRespFromLLM, nil
}

func (a *Agent) process_newFlow(req *ResponsesApiReq) ([]json.RawMessage, string, *ResponsesApiResp, error) {
	for i := uint64(0); i < a.maxToolsIterationWithoutReturnMessage; i++ {
		conversationResponse, err := a.endpoint.Post(req, a.inspectReq, a.inspectResp)
		if err != nil {
			return nil, "", conversationResponse, err
		}

		if conversationResponse.Error != nil {
			a.logger.Error().Msgf("error in conversation response: %v", conversationResponse.Error.Message)
			return nil, "", conversationResponse, errors.Join(ErrInConversationResponse, fmt.Errorf("err msg: %v", conversationResponse.Error.Message))
		}

		conversation, err := a.llmProvider.cc.ToConversationElements(conversationResponse.Output)
		if err != nil {
			return req.Input, "", conversationResponse, errors.Join(ErrUnknownResponseMessageFormat, err)
		}
		msg, fnCallsResp, imageHandled, err := a.dispatchConversation(conversation)
		for _, fResp := range fnCallsResp {
			conversation = append(conversation, fResp)
		}
		raw, errProviderRep := a.llmProvider.cc.ToProviderRepresentation(conversation)
		req.Input = append(req.Input, raw...)
		if errProviderRep != nil {
			return req.Input, "", conversationResponse, errors.Join(ErrUnknownResponseMessageFormat, errProviderRep)
		}
		if err != nil {
			return req.Input, "", conversationResponse, errors.Join(ErrUnknownResponseMessageFormat, err)
		}
		if imageHandled || msg != "" {
			return req.Input, msg, conversationResponse, nil
		}

	}

	a.logger.Error().Msgf("max tool iterations (%d) reached without a return message", a.maxToolsIterationWithoutReturnMessage)
	return req.Input, "", nil,
		errors.Join(ErrMaxToolIterationsReached,
			fmt.Errorf("exceeded %d iterations", a.maxToolsIterationWithoutReturnMessage))
}

func (a *Agent) dispatchConversation(conversation []ConversationElement) (string, []*FunctionCallResp, bool, error) {

	message := ""
	fnCallsResults := []*FunctionCallResp{}
	imageHandled := false

	var outputErr error

	for _, conversationElement := range conversation {
		switch el := conversationElement.(type) {
		//case *SystemMessage:

		case *FunctionCall:
			fResp, err := a.handleFunctionCall_newFlow(el)
			if err != nil {
				outputErr = errors.Join(outputErr, err)
			}
			if fResp != nil {
				fnCallsResults = append(fnCallsResults, fResp)
			}
		//case *UserMessage:

		case *AssistantMessage:
			message += messagesFromParts(el.Content)
			continue

		//case *FunctionCallResponse:

		case *Reasoning:
			continue

		case *ImageGeneration:
			err := a.handleImageGenerationCall_newFlow(el)
			if err != nil {
				outputErr = errors.Join(outputErr, err)
			}
			imageHandled = true
			continue

		default:
			continue // skip unknown types
		}
	}

	return message, fnCallsResults, imageHandled, outputErr
}

func messagesFromParts(parts []MessagePart) string {
	var msg strings.Builder
	for _, part := range parts {
		msg.WriteString(part.Text)
	}
	return msg.String()
}

func (a *Agent) handleImageGenerationCall_newFlow(image *ImageGeneration) error {

	// todo: return image as message or something?
	if a.handleImage_newFlow == nil {
		return ErrNoImageHandler
	}

	resultNote, err := a.handleImage_newFlow(image)
	if err != nil {
		return errors.Join(ErrCustomImageHandlerFailed, err)
	}

	image.Result = resultNote

	return nil
}

func (a *Agent) handleFunctionCall_newFlow(fn *FunctionCall) (*FunctionCallResp, error) {
	if a.toolset == nil {
		a.logger.Warn().Msgf("tool call (%s) but no tools provided)", fn.Name)
		return nil, ErrNoToolsetButToolCallRequested
	}

	funcCallResp, ok := a.toolset.DispatchTools(fn.Name, fn.CallId, fn.Args)
	if !ok {
		funcCallResp = invalidFunctionCallResp_newFlow(fn)
		return &funcCallResp, errors.Join(ErrWhileDispatchToolCall, fmt.Errorf("unknown tool name (%s)", fn.Name))
	}

	return &funcCallResp, nil
}

func invalidFunctionCallResp_newFlow(fn *FunctionCall) FunctionCallResp {
	return FunctionCallResp{
		Type:   "function_call_output",
		CallId: fn.CallId,
		Output: "invalid function call (function not found)" + fn.Name,
	}

}
