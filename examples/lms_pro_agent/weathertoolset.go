package main

import (
	"context"
	"encoding/json"
	"fmt"
	"rellm/pkg/rellm"
)

var _ rellm.Toolset = (*WeatherToolset)(nil)

type WeatherToolset struct{}

func (w *WeatherToolset) Definitions() []rellm.ToolDefinition {
	return []rellm.ToolDefinition{
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

func (w *WeatherToolset) Dispatch(_ context.Context, name string, _ json.RawMessage) (rellm.ToolCallResult, error) {
	switch name {
	case "WeatherToolset.GetWeather":
		res, err := GetWeather()
		return rellm.ToolCallResult{Value: res, Err: err}, nil
	}
	return rellm.ToolCallResult{}, fmt.Errorf("unknown tool name (%s)", name)
}
