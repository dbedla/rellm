package toolsets

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/dbedla/rellm/pkg/rellm"
)

func dispatch(t *testing.T, tool, args string) rellm.ToolCallResult {
	t.Helper()
	ts := NewCalculatorToolset(&Calculator{})
	res, err := ts.Dispatch(context.Background(), tool, json.RawMessage(args))
	assert.NoError(t, err)
	return res
}

func TestCalculatorToolsetDispatchAdd(t *testing.T) {
	res := dispatch(t, "Calculator_Add", `{"a":2,"b":3}`)
	assert.Equal(t, float64(5), res.Value)
}

func TestCalculatorToolsetDispatchSub(t *testing.T) {
	res := dispatch(t, "Calculator_Sub", `{"a":2,"b":3}`)
	assert.Equal(t, float64(-1), res.Value)
}

func TestCalculatorToolsetDispatchMul(t *testing.T) {
	res := dispatch(t, "Calculator_Mul", `{"a":2,"b":3}`)
	assert.Equal(t, float64(6), res.Value)
}

func TestCalculatorToolsetDispatchUnknownTool(t *testing.T) {
	ts := NewCalculatorToolset(&Calculator{})
	_, err := ts.Dispatch(context.Background(), "Calculator_Div", json.RawMessage(`{"a":2,"b":3}`))
	assert.Error(t, err)
}

func TestCalculatorToolsetDispatchBadJSON(t *testing.T) {
	res := dispatch(t, "Calculator_Add", `{oops`)
	assert.Error(t, res.Err)
}

func TestCalculatorToolsetDefinitions(t *testing.T) {
	ts := NewCalculatorToolset(&Calculator{})
	defs := ts.Definitions()
	assert.Len(t, defs, 3)

	want := []string{"Calculator_Add", "Calculator_Sub", "Calculator_Mul"}
	for i, name := range want {
		assert.Equal(t, name, defs[i].Name)
		assert.Equal(t, "function", defs[i].Type)
	}
}
