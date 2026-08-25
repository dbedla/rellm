package rellm

import "fmt"

var ErrBuildNoProvider = fmt.Errorf("missing provider")
var ErrBuildNoAgentName = fmt.Errorf("missing agent name")
var ErrBuildNoConversationStorage = fmt.Errorf("missing conversation storage")
var ErrMalformedConversationStorage = fmt.Errorf("malformed conversation storage line")

var ErrNoToolsetButToolCallRequested = fmt.Errorf("no toolset provided but request to call tool received")
var ErrWhileDispatchToolCall = fmt.Errorf("unable to dispatch tool call")
var ErrInConversationResponse = fmt.Errorf("error in conversation response")
var ErrMaxAgentStepsReached = fmt.Errorf("max agent steps for single prompt reached")
var ErrConversationElementConversion = fmt.Errorf("error converting conversation element")
var ErrUnknownConversationElement = fmt.Errorf("unknown conversation element")

var ErrEndpointNilResponse = fmt.Errorf("endpoint returned nil response")
var ErrEndpointNilBodyInResponse = fmt.Errorf("endpoint returned nil body in response")
var ErrUnableToReadResponseBody = fmt.Errorf("unable to read response body")

var ErrEmptyPrompt = fmt.Errorf("prompt message is empty")
var ErrEmptyReasoningEffort = fmt.Errorf("reasoning effort cannot be empty")

var ErrNoImageHandler = fmt.Errorf("image handler not provided")
var ErrCustomImageHandlerFailed = fmt.Errorf("custom image handler failed")
var ErrImageParsingFailed = fmt.Errorf("error converting image generation element")
var ErrReasoningParsingFailed = fmt.Errorf("error parsing reasoning element")

var ErrUserMsgConversionFailed = fmt.Errorf("user message conversion failed")

var ErrEndpointMissingModelName = fmt.Errorf("missing model name for provider")
var ErrEndpointMissingHost = fmt.Errorf("missing host (e.g. localhost or IP) for provider")
var ErrEndpointMissingPort = fmt.Errorf("missing port for provider")
