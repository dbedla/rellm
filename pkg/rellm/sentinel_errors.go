package rellm

import "fmt"

var ErrBuildNoModelName = fmt.Errorf("missing model name for endpoint")
var ErrBuildNoResponsesApiEndpoint = fmt.Errorf("missing ResponsesApiEndpoint")
var ErrBuildNoWorkspaceDir = fmt.Errorf("missing workspaceDir path, inside this dir all agent session artefact will be stored")
var ErrBuildNoHttpClient = fmt.Errorf("missing http client")
var ErrBuildNoEndpoint = fmt.Errorf("missing endpoint")
var ErrBuildNoAgentName = fmt.Errorf("missing agent name")
var ErrBuildExactlyOneLogger = fmt.Errorf("exactly one logger must be configured")
var ErrBuildLogger = fmt.Errorf("unable to create logger")

var ErrNoToolsetButToolCall = fmt.Errorf("request to call unknown tool")
var ErrUnknownToolCall = fmt.Errorf("request to call unknown tool")
var ErrInConversationResponse = fmt.Errorf("error in conversation response")
var ErrMaxToolIterationsReached = fmt.Errorf("max tool iterations reached without a text response")
var ErrModelReturnedUnproductiveOutput = fmt.Errorf("model returned output but no function call and no text message")

var ErrPromptIsExpired = fmt.Errorf("prompt is expired")

var ErrImageGenerationResp = fmt.Errorf("error in image generation response")
var ErrNoImageHandler = fmt.Errorf("image handler not provided")
var ErrCustomImageHandlerFailed = fmt.Errorf("custom image handler failed")

var ErrUserMsgConversionFailed = fmt.Errorf("user message conversion failed")
