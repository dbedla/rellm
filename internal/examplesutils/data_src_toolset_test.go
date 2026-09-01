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
	resp, err := toolset.Dispatch(ctx, "GetDataFor", arguments)

	assert.NoError(t, err)
	assert.NoError(t, resp.Err)

	data, ok := resp.Value.([]string)
	assert.True(t, ok)
	assert.Len(t, data, 2)

	assert.Equal(t, []string{"abc", "def"}, data)
}
