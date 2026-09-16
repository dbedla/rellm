package agentsutils

import (
	"context"
	"encoding/json"
	"fmt"

	"rellm/pkg/rellm"
)

var _ rellm.Toolset = (*CalculatorToolset)(nil)

// CalculatorToolset is a thin adapter that exposes Calculator to the agent.
type CalculatorToolset struct {
	*Calculator
}

func NewCalculatorToolset(c *Calculator) *CalculatorToolset {
	return &CalculatorToolset{Calculator: c}
}

func (t *CalculatorToolset) Definitions() []rellm.ToolDefinition {
	return []rellm.ToolDefinition{
		{
			Type:        "function",
			Name:        "Calculator_Add",
			Description: "Add two numbers.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"a": map[string]any{"type": "number"},
					"b": map[string]any{"type": "number"},
				},
				"required": []string{"a", "b"},
			},
		},
		{
			Type:        "function",
			Name:        "Calculator_Sub",
			Description: "Subtract b from a.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"a": map[string]any{"type": "number"},
					"b": map[string]any{"type": "number"},
				},
				"required": []string{"a", "b"},
			},
		},
		{
			Type:        "function",
			Name:        "Calculator_Mul",
			Description: "Multiply two numbers.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"a": map[string]any{"type": "number"},
					"b": map[string]any{"type": "number"},
				},
				"required": []string{"a", "b"},
			},
		},
	}
}

func (t *CalculatorToolset) Dispatch(_ context.Context, name string, args json.RawMessage) (rellm.ToolCallResult, error) {
	var a struct {
		A, B float64
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return rellm.ToolCallResult{Err: err}, nil
	}

	switch name {
	case "Calculator_Add":
		return rellm.ToolCallResult{Value: t.Add(a.A, a.B)}, nil
	case "Calculator_Sub":
		return rellm.ToolCallResult{Value: t.Sub(a.A, a.B)}, nil
	case "Calculator_Mul":
		return rellm.ToolCallResult{Value: t.Mul(a.A, a.B)}, nil
	default:
		return rellm.ToolCallResult{}, fmt.Errorf("unknown tool: %s", name)
	}
}
