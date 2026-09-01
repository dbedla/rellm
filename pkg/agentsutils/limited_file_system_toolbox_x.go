package agentsutils

import (
	"context"
	"encoding/json"
	"fmt"
	"rellm/pkg/rellm"
)

var _ rellm.Toolset_X = (*FSToolset_X)(nil)

type FSToolset_X struct {
	*LimitedFileSystem
}

func (f *FSToolset_X) BuildTools_X() []rellm.Tool {
	return (&FSToolset{}).BuildTools()
}

type fsPathArgs struct {
	Path string `json:"path"`
}

type fsWriteStringArgs struct {
	Content string `json:"content"`
	Path    string `json:"path"`
}

type fsWriteBytesArgs struct {
	Content []byte `json:"content"`
	Path    string `json:"path"`
}

func decodeArgs[T any](arguments json.RawMessage) (T, error) {
	var args T
	return args, json.Unmarshal(arguments, &args)
}

func (f *FSToolset_X) DispatchTools_X(_ context.Context, name string, arguments json.RawMessage) (rellm.ToolCallResult, error) {
	arguments = normalizeToolArguments(arguments)
	switch name {
	case "FSToolset.GetReadOnlyPaths":
		return rellm.ToolCallResult{Value: f.GetReadOnlyPaths()}, nil
	case "FSToolset.GetOutputDir":
		return rellm.ToolCallResult{Value: f.GetOutputDir()}, nil
	case "FSToolset.GetFileContentAsString":
		args, err := decodeArgs[fsPathArgs](arguments)
		if err != nil {
			return rellm.ToolCallResult{Err: err}, nil
		}
		res, err := f.GetFileContentAsString(args.Path)
		return rellm.ToolCallResult{Value: res, Err: err}, nil
	case "FSToolset.GetFileContentAsBytes":
		args, err := decodeArgs[fsPathArgs](arguments)
		if err != nil {
			return rellm.ToolCallResult{Err: err}, nil
		}
		res, err := f.GetFileContentAsBytes(args.Path)
		return rellm.ToolCallResult{Value: res, Err: err}, nil
	case "FSToolset.WriteStringToFile":
		args, err := decodeArgs[fsWriteStringArgs](arguments)
		if err != nil {
			return rellm.ToolCallResult{Err: err}, nil
		}
		return rellm.ToolCallResult{Value: "ok", Err: f.WriteStringToFile(args.Content, args.Path)}, nil
	case "FSToolset.WriteBytesToFile":
		args, err := decodeArgs[fsWriteBytesArgs](arguments)
		if err != nil {
			return rellm.ToolCallResult{Err: err}, nil
		}
		return rellm.ToolCallResult{Value: "ok", Err: f.WriteBytesToFile(args.Content, args.Path)}, nil
	case "FSToolset.DeleteFile":
		args, err := decodeArgs[fsPathArgs](arguments)
		if err != nil {
			return rellm.ToolCallResult{Err: err}, nil
		}
		return rellm.ToolCallResult{Value: "ok", Err: f.DeleteFile(args.Path)}, nil
	case "FSToolset.ListFilesIn":
		args, err := decodeArgs[fsPathArgs](arguments)
		if err != nil {
			return rellm.ToolCallResult{Err: err}, nil
		}
		res, err := f.ListFilesIn(args.Path)
		return rellm.ToolCallResult{Value: res, Err: err}, nil
	}
	return rellm.ToolCallResult{}, fmt.Errorf("unknown tool name (%s)", name)
}
