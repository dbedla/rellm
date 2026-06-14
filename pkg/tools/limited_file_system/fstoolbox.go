package lfs

import (
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

func (f *FSToolset) BuildTools() []rellm.Tool {
	return []rellm.Tool{
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
			Name:        "FSToolset.GetFileContentAsByte",
			Description: "GetFileContentAsByte returns the content of the file at the given path as a []byte\nerror will be returned if the path is outside FSSandbox.readOnlyDirs or FSSandbox.outputDir\nerror will be returned if the path does not exist or is not accessible.",
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

func (f *FSToolset) DispatchTools(name string, callID string, arguments string) (rellm.FunctionCallResp, bool) {
	switch name {
	case "FSToolset.GetReadOnlyPaths":
		return rellm.FuncResultToFunctionCallResp(callID, f.GetReadOnlyPaths()), true
	case "FSToolset.GetOutputDir":
		return rellm.FuncResultToFunctionCallResp(callID, f.GetOutputDir()), true
	case "FSToolset.GetFileContentAsString":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		res, err := f.GetFileContentAsString(args.Path)
		if err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		return rellm.FuncResultToFunctionCallResp(callID, res), true
	case "FSToolset.GetFileContentAsByte":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		res, err := f.GetFileContentAsByte(args.Path)
		if err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		return rellm.FuncResultToFunctionCallResp(callID, res), true
	case "FSToolset.WriteStringToFile":
		var args struct {
			Content string `json:"content"`
			Path    string `json:"path"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		err := f.WriteStringToFile(args.Content, args.Path)
		if err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		return rellm.FuncResultToFunctionCallResp(callID, "ok"), true
	case "FSToolset.WriteBytesToFile":
		var args struct {
			Content []byte `json:"content"`
			Path    string `json:"path"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		err := f.WriteBytesToFile(args.Content, args.Path)
		if err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		return rellm.FuncResultToFunctionCallResp(callID, "ok"), true
	case "FSToolset.DeleteFile":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		err := f.DeleteFile(args.Path)
		if err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		return rellm.FuncResultToFunctionCallResp(callID, "ok"), true
	case "FSToolset.ListFilesIn":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		res, err := f.ListFilesIn(args.Path)
		if err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, fmt.Sprintf("error: %s", err)), true
		}
		return rellm.FuncResultToFunctionCallResp(callID, res), true
	}
	return rellm.FunctionCallResp{}, false
}
