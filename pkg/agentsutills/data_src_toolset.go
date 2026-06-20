package agentsutills

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
	}
}

func (d *DataSrcToolset) DispatchTools(name string, callID string, arguments json.RawMessage) (rellm.FunctionCallResp, bool) {
	switch name {
	case "GetDataFor":
		var args struct {
			Input string `json:"input"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
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
