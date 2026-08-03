package rellm

import "fmt"

var ErrBuildNoModelName = fmt.Errorf("missing model name for endpoint")
var ErrBuildNoResponsesApiEndpoint = fmt.Errorf("missing ResponsesApiEndpoint")
var ErrBuildNoLLMProvider = fmt.Errorf("missing llm provider")
var ErrBuildNoWorkspaceDir = fmt.Errorf("missing workspaceDir path, inside this dir all agent session artefact will be stored")
var ErrBuildNoHttpClient = fmt.Errorf("missing http client")
var ErrBuildNoEndpoint = fmt.Errorf("missing endpoint")
var ErrBuildNoAgentName = fmt.Errorf("missing agent name")
var ErrBuildExactlyOneLogger = fmt.Errorf("exactly one logger must be configured")
var ErrBuildLogger = fmt.Errorf("unable to create logger")

var ErrNoToolsetButToolCallRequested = fmt.Errorf("no toolset provided but request to call tool received")
var ErrWhileDispatchToolCall = fmt.Errorf("unable to dispatch tool call")
var ErrInConversationResponse = fmt.Errorf("error in conversation response")
var ErrMaxToolIterationsReached = fmt.Errorf("max tool iterations reached without a text response")
var ErrModelReturnedUnproductiveOutput = fmt.Errorf("model returned output but no function call and no text message")
var ErrUnknownResponseMessageFormat = fmt.Errorf("unknown message format")

var ErrEmptyPrompt = fmt.Errorf("prompt message is empty")
var ErrEmptyReasoningEffort = fmt.Errorf("reasoning effort cannot be empty")

var ErrImageGenerationResp = fmt.Errorf("error in image generation response")
var ErrNoImageHandler = fmt.Errorf("image handler not provided")
var ErrCustomImageHandlerFailed = fmt.Errorf("custom image handler failed")

var ErrUserMsgConversionFailed = fmt.Errorf("user message conversion failed")

var ErrEndpointMissingModelName = fmt.Errorf("missing model name for endpoint")
var ErrEndpointMissingHost = fmt.Errorf("missing host (e.g. localhost or IP) for endpoint")
var ErrEndpointMissingPort = fmt.Errorf("missing port for endpoint")
