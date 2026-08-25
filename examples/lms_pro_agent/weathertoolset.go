package main

import (
	"context"
	"encoding/json"
	"rellm/pkg/rellm"
)

var _ rellm.Toolset = (*WeatherToolset)(nil)

type WeatherToolset struct {
}

func (w *WeatherToolset) BuildTools() []rellm.Tool {
	return []rellm.Tool{
		{
			Type:        "function",
			Name:        "WeatherToolset.GetWeather",
			Description: "GetWeather returns current weather information",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}
}

func (w *WeatherToolset) DispatchTools(_ context.Context, name string, callID string, arguments json.RawMessage) (rellm.FunctionCallResp, bool) {
	switch name {
	case "WeatherToolset.GetWeather":
		res, err := GetWeather()
		if err != nil {
			return rellm.FuncResultToFunctionCallResp(callID, err.Error()), true
		}
		return rellm.FuncResultToFunctionCallResp(callID, res), true
	}
	return rellm.FunctionCallResp{}, false
}
