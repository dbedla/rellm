package rellm

import "fmt"

// builders
var ErrBuildNoModelName = fmt.Errorf("missing model name for endpoint")
var ErrBuildNoResponsesApiEndpoint = fmt.Errorf("missing ResponsesApiEndpoint")
var ErrBuildNoWorkspaceDir = fmt.Errorf("missing workspaceDir path, inside this dir all agent session artefact will be stored")
var ErrBuildNoHttpClient = fmt.Errorf("missing http client")
var ErrBuildNoEndpoint = fmt.Errorf("missing endpoint")
var ErrBuildNoAgentName = fmt.Errorf("missing agent name")
var ErrBuildExactlyOneLogger = fmt.Errorf("exactly one logger must be configured")
var ErrBuildLogger = fmt.Errorf("unable to create logger")

var ErrNoToolsetBuToolCall = fmt.Errorf("request to call unknown tool")

var ErrUnknownToolCallsErrorsWillBePassedToModelInNextReq = fmt.Errorf("request to call unknown tool")
var ErrInConversationResponse = fmt.Errorf("error in conversation response")
var ErrTooManyMessagesInResponse = fmt.Errorf("too many messages in response")

var ErrImageGenerationResp = fmt.Errorf("error in image generation response")
