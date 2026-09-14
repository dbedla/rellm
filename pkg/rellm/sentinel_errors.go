package rellm

import (
	"errors"
)

var ErrBuildNoProvider = errors.New("missing provider")
var ErrBuildNoConversationStorage = errors.New("missing conversation")
var ErrMalformedConversationStorage = errors.New("malformed conversation storage line")

var ErrNoNewConversationElementAfterDispatch = errors.New("no new conversation element")
var ErrNoToolsetButToolCallRequested = errors.New("no toolset provided but request to call tool received")
var ErrWhileDispatchToolCall = errors.New("unable to dispatch tool call")
var ErrInConversationResponse = errors.New("error in conversation response")
var ErrMaxAgentStepsReached = errors.New("max agent steps for single prompt reached")
var ErrConversationElementConversion = errors.New("error converting conversation element")
var ErrUnknownTypeForConversationElement = errors.New("unknown type for conversation element")
var ErrUnknownElementProviderMismatch = errors.New("unknown element provider mismatch")
var ErrMalformedUnknownElement = errors.New("malformed unknown element")
var ErrLMSMarshalingConversationElement = errors.New("error marshaling LM Studio conversation element")
var ErrOpenRouterMarshalingConversationElement = errors.New("error marshaling OpenRouter conversation element")

var ErrEndpointNilResponse = errors.New("endpoint returned nil response")
var ErrEndpointNilBodyInResponse = errors.New("endpoint returned nil body in response")
var ErrUnableToReadResponseBody = errors.New("unable to read response body")

var ErrEmptyPrompt = errors.New("prompt message is empty")
var ErrEmptyReasoningEffort = errors.New("reasoning effort cannot be empty")
var ErrEmptyTextFormat = errors.New("text format type cannot be empty")
var ErrEmptyTextFormatName = errors.New("json_schema text format requires a name")
var ErrEmptyTextFormatSchema = errors.New("json_schema text format requires a schema")
var ErrUnknownTextFormatType = errors.New("unknown text format type")

var ErrNoUnknownConversationElementHandler = errors.New("unknown conversation element handler not provided")
var ErrCustomConversationElementHandlerFailed = errors.New("custom unknown conversation element handler failed")
var ErrNoImageHandler = errors.New("image handler not provided")
var ErrImageHandlerFailed = errors.New("image handler failed")
var ErrImageParsingFailed = errors.New("error converting image generation element")
var ErrReasoningParsingFailed = errors.New("error parsing reasoning element")

var ErrUserMsgConversionFailed = errors.New("user message conversion failed")

var ErrEndpointMissingModelName = errors.New("missing model name for provider")
var ErrEndpointMissingHost = errors.New("missing host (e.g. localhost or IP) for provider")
var ErrEndpointMissingPort = errors.New("missing port for provider")

var ErrMissingHTTPClientForProvider = errors.New("missing HTTP client for provider")
var ErrMissingApiKeyForProvider = errors.New("missing API key for provider")
