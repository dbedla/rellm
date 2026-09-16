package examplesutils_test

import (
	"os"
	"path/filepath"
	"rellm/internal/examplesutils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFSSandbox(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")

	err := os.Mkdir(readOnlyDir, 0555) // Read and execute only
	assert.NoError(t, err, "failed to create readonly dir")

	err = os.Mkdir(outputDir, 0777) // Read, write, and execute
	assert.NoError(t, err, "failed to create output dir")

	t.Run("Valid directories", func(t *testing.T) {
		sandbox, err := examplesutils.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
		assert.NoError(t, err, "expected no error")
		assert.NotNil(t, sandbox, "expected sandbox instance")
	})

	t.Run("ReadOnly dir does not exist", func(t *testing.T) {
		_, err := examplesutils.NewLimitedFileSystem([]string{filepath.Join(tmpDir, "nonexistent")}, outputDir)
		assert.Error(t, err, "expected error for nonexistent readonly dir")
	})

	t.Run("OutputDir does not exist", func(t *testing.T) {
		_, err := examplesutils.NewLimitedFileSystem([]string{readOnlyDir}, filepath.Join(tmpDir, "nonexistent_output"))
		assert.Error(t, err, "expected error for nonexistent output dir")
	})

	t.Run("ReadOnly dir no read access", func(t *testing.T) {
		noAccessDir := filepath.Join(tmpDir, "no_access")
		err := os.Mkdir(noAccessDir, 0000)
		assert.NoError(t, err, "failed to create no access dir")
		defer func() {
			err := os.Chmod(noAccessDir, 0755)
			assert.NoError(t, err, "failed to set permissions for no access dir")
		}() // cleanup to allow TempDir to be deleted

		_, err = examplesutils.NewLimitedFileSystem([]string{noAccessDir}, outputDir)
		assert.Error(t, err, "expected error for readonly dir with no read access")
	})

	t.Run("OutputDir no write access", func(t *testing.T) {
		noWriteOutputDir := filepath.Join(tmpDir, "no_write_output")
		err := os.Mkdir(noWriteOutputDir, 0555)
		assert.NoError(t, err, "failed to create no write output dir")
		defer func() {
			err := os.Chmod(noWriteOutputDir, 0755)
			assert.NoError(t, err, "failed to set permissions for no write output dir")
		}() // cleanup

		_, err = examplesutils.NewLimitedFileSystem([]string{readOnlyDir}, noWriteOutputDir)
		assert.Error(t, err, "expected error for output dir with no write access")
	})
}

func TestListFilesIn(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")
	outsideDir := filepath.Join(tmpDir, "outside")

	err := os.Mkdir(readOnlyDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(outputDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(outsideDir, 0755)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(readOnlyDir, "file1.txt"), []byte("content1"), 0644)
	assert.NoError(t, err)
	err = os.Mkdir(filepath.Join(readOnlyDir, "subdir"), 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(readOnlyDir, "subdir1", "subdir2", "subdir3", "subdir4"), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(readOnlyDir, "subdir1", "subdir2", "subdir3", "subdir4", "somefile.txt"), []byte("deep"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(readOnlyDir, "subdir", "nested.txt"), []byte("nested"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(outputDir, "file2.txt"), []byte("content2"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(outsideDir, "file3.txt"), []byte("content3"), 0644)
	assert.NoError(t, err)

	sandbox, err := examplesutils.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	assert.NoError(t, err, "failed to create sandbox")

	t.Run("List files in read-only dir", func(t *testing.T) {
		files, err := sandbox.ListFilesIn(readOnlyDir)
		assert.NoError(t, err, "expected no error")
		expectedPath, err := filepath.EvalSymlinks(filepath.Join(readOnlyDir, "file1.txt"))
		assert.NoError(t, err)
		nestedPath, err := filepath.EvalSymlinks(filepath.Join(readOnlyDir, "subdir", "nested.txt"))
		assert.NoError(t, err)
		deepPath, err := filepath.EvalSymlinks(filepath.Join(readOnlyDir, "subdir1", "subdir2", "subdir3", "subdir4", "somefile.txt"))
		assert.NoError(t, err)
		assert.Len(t, files, 3)
		assert.Contains(t, files, expectedPath)
		assert.Contains(t, files, nestedPath)
		assert.Contains(t, files, deepPath)
	})

	t.Run("List files in output dir", func(t *testing.T) {
		files, err := sandbox.ListFilesIn(outputDir)
		assert.NoError(t, err, "expected no error")
		expectedPath, err := filepath.EvalSymlinks(filepath.Join(outputDir, "file2.txt"))
		assert.NoError(t, err)
		assert.Len(t, files, 1)
		assert.Equal(t, expectedPath, files[0])
	})

	t.Run("Lists deeply nested files", func(t *testing.T) {
		deepDir := filepath.Join(outputDir, "subdir1", "subdir2", "subdir3", "subdir4")
		err := os.MkdirAll(deepDir, 0755)
		assert.NoError(t, err)
		deepFile := filepath.Join(deepDir, "somefile.txt")
		err = os.WriteFile(deepFile, []byte("deep"), 0644)
		assert.NoError(t, err)

		files, err := sandbox.ListFilesIn(outputDir)
		assert.NoError(t, err, "expected no error")
		expectedDeep, err := filepath.EvalSymlinks(deepFile)
		assert.NoError(t, err)
		assert.Contains(t, files, expectedDeep)
	})

	t.Run("Error for outside path", func(t *testing.T) {
		_, err := sandbox.ListFilesIn(outsideDir)
		assert.Error(t, err, "expected error for path outside sandbox")
	})

	t.Run("Error for non-existent path", func(t *testing.T) {
		_, err := sandbox.ListFilesIn(filepath.Join(readOnlyDir, "nonexistent"))
		assert.Error(t, err, "expected error for non-existent path")
	})

	t.Run("Error for file path (not a directory)", func(t *testing.T) {
		_, err := sandbox.ListFilesIn(filepath.Join(readOnlyDir, "file1.txt"))
		assert.Error(t, err, "expected error for file path")
	})
}

func TestGetFileContentAsString(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")
	outsideDir := filepath.Join(tmpDir, "outside")
	readOnlySiblingDir := filepath.Join(tmpDir, "readonly-evil")

	err := os.Mkdir(readOnlyDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(outputDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(outsideDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(readOnlySiblingDir, 0755)
	assert.NoError(t, err)

	content1 := "content1"
	content2 := "content2"
	err = os.WriteFile(filepath.Join(readOnlyDir, "file1.txt"), []byte(content1), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(outputDir, "file2.txt"), []byte(content2), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(outsideDir, "file3.txt"), []byte("content3"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(readOnlySiblingDir, "secret.txt"), []byte("secret"), 0644)
	assert.NoError(t, err)

	sandbox, err := examplesutils.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	assert.NoError(t, err, "failed to create sandbox")

	t.Run("Read file in read-only dir", func(t *testing.T) {
		content, err := sandbox.GetFileContentAsString(filepath.Join(readOnlyDir, "file1.txt"))
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, content1, content)
	})

	t.Run("Read file in output dir", func(t *testing.T) {
		content, err := sandbox.GetFileContentAsString(filepath.Join(outputDir, "file2.txt"))
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, content2, content)
	})

	t.Run("Error for outside path", func(t *testing.T) {
		_, err := sandbox.GetFileContentAsString(filepath.Join(outsideDir, "file3.txt"))
		assert.Error(t, err, "expected error for path outside sandbox")
	})

	t.Run("Error for sibling prefix attack", func(t *testing.T) {
		_, err := sandbox.GetFileContentAsString(filepath.Join(readOnlySiblingDir, "secret.txt"))
		assert.Error(t, err, "expected error for sibling prefix path")
	})

	t.Run("Error for symlink escape", func(t *testing.T) {
		linkPath := filepath.Join(readOnlyDir, "outside_link")
		err := os.Symlink(outsideDir, linkPath)
		assert.NoError(t, err)

		_, err = sandbox.GetFileContentAsString(filepath.Join(linkPath, "file3.txt"))
		assert.Error(t, err, "expected error for symlink escape")
	})

	t.Run("Error for non-existent file", func(t *testing.T) {
		_, err := sandbox.GetFileContentAsString(filepath.Join(readOnlyDir, "nonexistent.txt"))
		assert.Error(t, err, "expected error for non-existent file")
	})

	t.Run("Error for directory path", func(t *testing.T) {
		_, err := sandbox.GetFileContentAsString(readOnlyDir)
		assert.Error(t, err, "expected error for directory path")
	})
}

func TestGetFileContentAsBytes(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")

	err := os.Mkdir(readOnlyDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(outputDir, 0755)
	assert.NoError(t, err)

	content1 := []byte("content1")
	err = os.WriteFile(filepath.Join(readOnlyDir, "file1.txt"), content1, 0644)
	assert.NoError(t, err)

	sandbox, err := examplesutils.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	assert.NoError(t, err, "failed to create sandbox")

	t.Run("Read file in read-only dir", func(t *testing.T) {
		content, err := sandbox.GetFileContentAsBytes(filepath.Join(readOnlyDir, "file1.txt"))
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, content1, content)
	})

	t.Run("Error for non-existent file", func(t *testing.T) {
		_, err := sandbox.GetFileContentAsBytes(filepath.Join(readOnlyDir, "nonexistent.txt"))
		assert.Error(t, err, "expected error for non-existent file")
	})
}

func TestWriteToFile(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")
	outputSiblingDir := filepath.Join(tmpDir, "output-evil")

	err := os.Mkdir(readOnlyDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(outputDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(outputSiblingDir, 0755)
	assert.NoError(t, err)

	sandbox, err := examplesutils.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	assert.NoError(t, err, "failed to create sandbox")

	t.Run("Write string to new file in output dir", func(t *testing.T) {
		path := filepath.Join(outputDir, "new_file.txt")
		content := "hello world"
		err := sandbox.WriteStringToFile(content, path)
		assert.NoError(t, err, "expected no error")

		// Verify content
		data, err := os.ReadFile(path)
		assert.NoError(t, err, "failed to read written file")
		assert.Equal(t, content, string(data))
	})

	t.Run("Write bytes to new file in output dir", func(t *testing.T) {
		path := filepath.Join(outputDir, "new_bytes.bin")
		content := []byte{0x01, 0x02, 0x03}
		err := sandbox.WriteBytesToFile(content, path)
		assert.NoError(t, err, "expected no error")

		// Verify content
		data, err := os.ReadFile(path)
		assert.NoError(t, err, "failed to read written file")
		assert.Equal(t, content, data)
	})

	t.Run("Error when file already exists (string)", func(t *testing.T) {
		path := filepath.Join(outputDir, "existing.txt")
		err := os.WriteFile(path, []byte("existing"), 0644)
		assert.NoError(t, err)

		err = sandbox.WriteStringToFile("new content", path)
		assert.Error(t, err, "expected error for existing file")
	})

	t.Run("Error when file already exists (bytes)", func(t *testing.T) {
		path := filepath.Join(outputDir, "existing_bytes.bin")
		err := os.WriteFile(path, []byte("existing"), 0644)
		assert.NoError(t, err)

		err = sandbox.WriteBytesToFile([]byte{0x01}, path)
		assert.Error(t, err, "expected error for existing file")
	})

	t.Run("Error for path outside output dir (read-only dir)", func(t *testing.T) {
		path := filepath.Join(readOnlyDir, "trying_to_write.txt")
		err := sandbox.WriteStringToFile("content", path)
		assert.Error(t, err, "expected error for path in read-only dir")
	})

	t.Run("Error for path completely outside sandbox", func(t *testing.T) {
		path := filepath.Join(tmpDir, "completely_outside.txt")
		err := sandbox.WriteStringToFile("content", path)
		assert.Error(t, err, "expected error for path outside sandbox")
	})

	t.Run("Error for sibling prefix attack", func(t *testing.T) {
		path := filepath.Join(outputSiblingDir, "secret.txt")
		err := sandbox.WriteStringToFile("content", path)
		assert.Error(t, err, "expected error for sibling prefix path")
	})

	t.Run("Error for parent traversal", func(t *testing.T) {
		path := filepath.Join(outputDir, "..", "traversal.txt")
		err := sandbox.WriteStringToFile("content", path)
		assert.Error(t, err, "expected error for parent traversal")
	})

	t.Run("Error for symlink escape", func(t *testing.T) {
		linkPath := filepath.Join(outputDir, "outside_link")
		err := os.Symlink(outputSiblingDir, linkPath)
		assert.NoError(t, err)

		err = sandbox.WriteStringToFile("content", filepath.Join(linkPath, "secret.txt"))
		assert.Error(t, err, "expected error for symlink escape")
	})
}

func TestDeleteFile(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")

	err := os.Mkdir(readOnlyDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(outputDir, 0755)
	assert.NoError(t, err)

	sandbox, err := examplesutils.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	assert.NoError(t, err, "failed to create sandbox")

	t.Run("Delete existing file in output dir", func(t *testing.T) {
		path := filepath.Join(outputDir, "to_delete.txt")
		err = os.WriteFile(path, []byte("goodbye"), 0644)
		assert.NoError(t, err)

		err = sandbox.DeleteFile(path)
		assert.NoError(t, err, "expected no error")

		// Verify file is gone
		_, err := os.Stat(path)
		assert.Error(t, err, "file still exists after deletion")
	})

	t.Run("Error for path outside output dir (read-only dir)", func(t *testing.T) {
		path := filepath.Join(readOnlyDir, "cannot_delete.txt")
		err = os.WriteFile(path, []byte("protected"), 0644)
		assert.NoError(t, err)

		err := sandbox.DeleteFile(path)
		assert.Error(t, err, "expected error for deleting from read-only dir")

		// Verify file is still there
		_, err = os.Stat(path)
		assert.NoError(t, err, "file should still exist")
	})

	t.Run("Error for path outside output dir (completely outside)", func(t *testing.T) {
		path := filepath.Join(tmpDir, "completely_outside.txt")
		err = os.WriteFile(path, []byte("outside"), 0644)
		assert.NoError(t, err)

		err = sandbox.DeleteFile(path)
		assert.Error(t, err, "expected error for path outside sandbox")
	})

	t.Run("Error for non-existent file", func(t *testing.T) {
		path := filepath.Join(outputDir, "non_existent.txt")
		err := sandbox.DeleteFile(path)
		assert.Error(t, err, "expected error for non-existent file")
	})

	t.Run("Error for directory path", func(t *testing.T) {
		path := filepath.Join(outputDir, "some_dir")
		err = os.Mkdir(path, 0755)
		assert.NoError(t, err)

		err = sandbox.DeleteFile(path)
		assert.Error(t, err, "expected error for directory path")
	})
}

func TestCheckWriteAccessNoFixedNameOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")
	err := os.Mkdir(readOnlyDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(outputDir, 0755)
	assert.NoError(t, err)

	baitPath := filepath.Join(outputDir, ".write_test")
	err = os.WriteFile(baitPath, []byte("do not delete"), 0644)
	assert.NoError(t, err)

	sandbox, err := examplesutils.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	assert.NoError(t, err)
	content, err := sandbox.GetFileContentAsString(baitPath)
	assert.NoError(t, err)
	assert.Equal(t, content, "do not delete")

	data, err := os.ReadFile(baitPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("do not delete"), data)
}

func TestReadOnlyDirsSliceIsolation(t *testing.T) {
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")
	err := os.Mkdir(readOnlyDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(outputDir, 0755)
	assert.NoError(t, err)

	inputDirs := []string{readOnlyDir}
	sandbox, err := examplesutils.NewLimitedFileSystem(inputDirs, outputDir)
	assert.NoError(t, err)

	// Mutate input slice after construction
	inputDirs[0] = filepath.Join(tmpDir, "mutated")
	assert.Equal(t, readOnlyDir, sandbox.GetReadOnlyPaths()[0])

	// Mutate result of getter
	got := sandbox.GetReadOnlyPaths()
	got[0] = filepath.Join(tmpDir, "mutated2")
	assert.Equal(t, readOnlyDir, sandbox.GetReadOnlyPaths()[0])
}

func TestOperationsDoNotMutateReadOnlyDirs(t *testing.T) {
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")
	err := os.Mkdir(readOnlyDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(outputDir, 0755)
	assert.NoError(t, err)

	sandbox, err := examplesutils.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	assert.NoError(t, err)

	existing := filepath.Join(readOnlyDir, "exists.txt")
	err = os.WriteFile(existing, []byte("hi"), 0644)
	assert.NoError(t, err)

	before := sandbox.GetReadOnlyPaths()
	_, _ = sandbox.GetFileContentAsString(existing)
	_, _ = sandbox.ListFilesIn(readOnlyDir)
	after := sandbox.GetReadOnlyPaths()

	assert.Equal(t, before, after, "readOnlyDirs mutated after isAllowed")
}
