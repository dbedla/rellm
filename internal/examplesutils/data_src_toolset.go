package examplesutils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dbedla/rellm/pkg/rellm"
)

var _ rellm.Toolset = (*DataSrcToolset)(nil)

type DataSrcToolset struct {
	TestDataSource
}

func (d *DataSrcToolset) Definitions() []rellm.ToolDefinition {
	return []rellm.ToolDefinition{
		{
			Type:        "function",
			Name:        "GetDataFor",
			Description: "Get data for a specific input",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"input": map[string]any{
						"type": "string",
					},
				},
				"required": []string{"input"},
			},
		},
		{
			Type:        "function",
			Name:        "GetStaticData",
			Description: "Get static data",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}
}

func (d *DataSrcToolset) Dispatch(_ context.Context, name string, arguments json.RawMessage) (rellm.ToolCallResult, error) {
	switch name {
	case "GetDataFor":
		args, err := parseGetDataForArgs(arguments)
		if err != nil {
			return rellm.ToolCallResult{Err: err}, nil
		}
		return rellm.ToolCallResult{Value: d.GetDataFor(args.Input)}, nil
	case "GetStaticData":
		return rellm.ToolCallResult{Value: d.GetStaticData()}, nil
	}
	return rellm.ToolCallResult{}, fmt.Errorf("unknown tool name (%s)", name)
}

type getDataForArgs struct {
	Input string `json:"input"`
}

func parseGetDataForArgs(arguments json.RawMessage) (getDataForArgs, error) {
	var args getDataForArgs
	if err := json.Unmarshal(arguments, &args); err == nil {
		return args, nil
	}

	var rawArgs string
	if err := json.Unmarshal(arguments, &rawArgs); err != nil {
		return args, err
	}
	err := json.Unmarshal([]byte(rawArgs), &args)
	return args, err
}
