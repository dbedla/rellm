package examplesutils

import (
	"context"
	"encoding/json"
	"fmt"
	"rellm/pkg/rellm"
)

var _ rellm.Toolset_X = (*DataSrcToolset_X)(nil)

type DataSrcToolset_X struct {
	TestDataSource
}

func (d *DataSrcToolset_X) BuildTools_X() []rellm.Tool {
	return []rellm.Tool{
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

func (d *DataSrcToolset_X) DispatchTools_X(_ context.Context, name string, arguments json.RawMessage) (rellm.ToolCallResult, error) {
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
