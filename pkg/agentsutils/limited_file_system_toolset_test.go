package agentsutils

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFSToolsetDispatchToolsAcceptsStringWrappedArguments(t *testing.T) {
	readOnlyDir, outputDir := buildToolsetDirs(t)
	toolset := buildToolset(t, readOnlyDir, outputDir)
	arguments := buildStringWrappedArguments(t, readOnlyDir)

	ctx := context.Background()
	resp, err := toolset.Dispatch(ctx, "FSToolset_ListFilesIn", arguments)

	assert.NoError(t, err)
	assert.NoError(t, resp.Err)

	rv, ok := resp.Value.([]string)
	assert.True(t, ok)
	assert.Len(t, rv, 1)
	assert.Contains(t, rv[0], filepath.Join(readOnlyDir, "file.txt"))
}

func buildToolsetDirs(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")
	assert.NoError(t, os.Mkdir(readOnlyDir, 0755))
	assert.NoError(t, os.Mkdir(outputDir, 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(readOnlyDir, "file.txt"), []byte("content"), 0644))
	return readOnlyDir, outputDir
}

func buildToolset(t *testing.T, readOnlyDir string, outputDir string) *FSToolset {
	t.Helper()
	limitedFS, err := NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	assert.NoError(t, err)
	return &FSToolset{limitedFS}
}

func buildStringWrappedArguments(t *testing.T, path string) json.RawMessage {
	t.Helper()
	wrapped, err := json.Marshal(map[string]string{"path": path})
	assert.NoError(t, err)
	doubleWrapped, err := json.Marshal(string(wrapped))
	assert.NoError(t, err)
	return doubleWrapped
}
