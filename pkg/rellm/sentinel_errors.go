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

// ./dispatcher.go:34:			return nil, "", conversationResponse, fmt.Errorf("error in conversation response: %v", conversationResponse.Error.Message)
var ErrNoToolsetBuToolCall = fmt.Errorf("request to call unknown tool")

var ErrUnknownToolCallsErrorsWillBePassedToModelInNextReq = fmt.Errorf("request to call unknown tool")

//./dispatcher.go:130:		return nil, "", fmt.Errorf("too many messages in response")
//./endpoint.go:103:		return nil, errors.Join(err, fmt.Errorf("%s", string(rawBody)))
//./endpoint.go:193:		return *new(K), fmt.Errorf("error unmarshalling: %s; unmarshaling type %T; raw: %s", err, data, string(rawBody))
//./conversation_storage/conversation_storage.go:28:		return fmt.Errorf("failed to create directory: %w", err)
//./conversation_storage/conversation_storage.go:36:		return fmt.Errorf("failed to marshal conversation: %w", err)
//./conversation_storage/conversation_storage.go:40:		return fmt.Errorf("failed to write conversation file: %w", err)
//./conversation_storage/conversation_storage.go:52:		return nil, fmt.Errorf("failed to read directory: %w", err)
//./conversation_storage/conversation_storage.go:79:		return nil, fmt.Errorf("failed to read conversation file: %w", err)
//./conversation_storage/conversation_storage.go:84:		return nil, fmt.Errorf("failed to unmarshal conversation: %w", err)
//./prompt.go:16:		return nil, fmt.Errorf("promptMessageToConversation: %w", err)
