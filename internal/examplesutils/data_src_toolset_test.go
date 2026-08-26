package examplesutils

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataSrcToolsetDispatchesGetDataFor(t *testing.T) {
	toolset := &DataSrcToolset{}
	arguments := json.RawMessage(`{"input":"test"}`)

	ctx := context.Background()
	resp, ok := toolset.DispatchTools(ctx, "GetDataFor", "call_123", arguments)

	assert.True(t, ok)
	assert.Equal(t, "function_call_output", resp.Type)
	assert.Equal(t, "call_123", resp.CallID)
	assert.JSONEq(t, `["abc","def"]`, resp.Output)
}

func TestDataSrcToolsetDispatchesGetDataForWithStringArguments(t *testing.T) {
	toolset := &DataSrcToolset{}
	arguments := json.RawMessage(`"{\"input\":\"test\"}"`)

	ctx := context.Background()
	resp, ok := toolset.DispatchTools(ctx, "GetDataFor", "call_123", arguments)

	assert.True(t, ok)
	assert.Equal(t, "function_call_output", resp.Type)
	assert.Equal(t, "call_123", resp.CallID)
	assert.JSONEq(t, `["abc","def"]`, resp.Output)
}
