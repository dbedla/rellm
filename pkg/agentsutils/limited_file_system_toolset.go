package agentsutils

import (
	"context"
	"encoding/json"
	"fmt"
	"rellm/pkg/rellm"
)

var _ rellm.Toolset = (*FSToolset)(nil)

type FSToolset struct {
	*LimitedFileSystem
}

func NewFSToolset(lfs *LimitedFileSystem) *FSToolset {
	return &FSToolset{lfs}
}

func (f *FSToolset) Definitions() []rellm.ToolDefinition {
	return []rellm.ToolDefinition{
		{
			Type:        "function",
			Name:        "FSToolset.GetReadOnlyPaths",
			Description: "GetReadOnlyPaths returns a list of paths that are allowed to be read-only from",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Type:        "function",
			Name:        "FSToolset.GetOutputDir",
			Description: "GetOutputDir returns the path of the output directory\nin this directory read, write and delete operation are allowed",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Type:        "function",
			Name:        "FSToolset.GetFileContentAsString",
			Description: "GetFileContentAsString returns the content of the file at the given path as a string\nerror will be returned if the path is outside FSSandbox.readOnlyDirs or FSSandbox.outputDir\nerror will be returned if the path does not exist or is not accessible.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type": "string",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Type:        "function",
			Name:        "FSToolset.GetFileContentAsBytes",
			Description: "GetFileContentAsBytes returns the content of the file at the given path as a []byte\nerror will be returned if the path is outside FSSandbox.readOnlyDirs or FSSandbox.outputDir\nerror will be returned if the path does not exist or is not accessible.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type": "string",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Type:        "function",
			Name:        "FSToolset.WriteStringToFile",
			Description: "WriteStringToFile writes a given string into a specific file.\nif the file does not exist, it will be created\nerror will be if the file already exists\nerror will be returned if the path is outside FSSandbox.outputDir",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"content": map[string]any{
						"type": "string",
					},
					"path": map[string]any{
						"type": "string",
					},
				},
				"required": []string{"content", "path"},
			},
		},
		{
			Type:        "function",
			Name:        "FSToolset.WriteBytesToFile",
			Description: "WriteBytesToFile writes bytes into a specific file.\nif the file does not exist, it will be created\nerror will be if the file already exists\nerror will be returned if the path is outside FSSandbox.outputDir",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"content": map[string]any{
						"type":        "string",
						"description": "base64 encoded bytes",
					},
					"path": map[string]any{
						"type": "string",
					},
				},
				"required": []string{"content", "path"},
			},
		},
		{
			Type:        "function",
			Name:        "FSToolset.DeleteFile",
			Description: "DeleteFile delete file.\nerror will be returned if the path is outside FSSandbox.outputDir\nerror will be returned in any other standard case during deletion",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type": "string",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Type:        "function",
			Name:        "FSToolset.ListFilesIn",
			Description: "ListFilesIn returns a list of files in the given path\nerror will be returned if the path does not exist or is not accessible.\nerror will be returned if the path is not a directory\nerror will be returned if the path is outside FSSandbox.readOnlyDirs or FSSandbox.outputDir",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type": "string",
					},
				},
				"required": []string{"path"},
			},
		},
	}
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

func (f *FSToolset) Dispatch(_ context.Context, name string, arguments json.RawMessage) (rellm.ToolCallResult, error) {
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

func normalizeToolArguments(arguments json.RawMessage) json.RawMessage {
	var nested string
	if err := json.Unmarshal(arguments, &nested); err != nil {
		return arguments
	}
	if !json.Valid([]byte(nested)) {
		return arguments
	}
	return json.RawMessage(nested)
}
