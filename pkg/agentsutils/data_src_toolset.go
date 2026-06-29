package agentsutils

import (
	"encoding/json"
	"rellm/pkg/rellm"
)

type DataSrcToolset struct {
	TestDataSource
}

func (d *DataSrcToolset) BuildTools() []rellm.Tool {
	return []rellm.Tool{
		{
			Type:        "function",
			Name:        "GetDataFor",
			Description: "Get data for a specific input",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"input": map[string]interface{}{
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
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Type:        "function",
			Name:        "GetSpecialData",
			Description: "Get static data",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (d *DataSrcToolset) DispatchTools(name string, callID string, arguments json.RawMessage) (rellm.FunctionCallResp, bool) {
	switch name {
	case "GetDataFor":
		args, err := parseGetDataForArgs(arguments)
		if err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, "invalid arguments"), true
		}
		res := d.GetDataFor(args.Input)
		return rellm.FuncResultToFunctionCallResp(callID, res), true
	case "GetStaticData":
		res := d.GetStaticData()
		return rellm.FuncResultToFunctionCallResp(callID, res), true
	}
	return rellm.FunctionCallResp{}, false
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
