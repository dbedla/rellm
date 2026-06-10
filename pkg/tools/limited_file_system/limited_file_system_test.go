package lfs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFSSandbox(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")

	err := os.Mkdir(readOnlyDir, 0555) // Read and execute only
	if err != nil {
		t.Fatalf("failed to create readonly dir: %v", err)
	}

	err = os.Mkdir(outputDir, 0777) // Read, write, and execute
	if err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	t.Run("Valid directories", func(t *testing.T) {
		sandbox, err := NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if sandbox == nil {
			t.Fatal("expected sandbox instance, got nil")
		}
	})

	t.Run("ReadOnly dir does not exist", func(t *testing.T) {
		_, err := NewLimitedFileSystem([]string{filepath.Join(tmpDir, "nonexistent")}, outputDir)
		if err == nil {
			t.Error("expected error for nonexistent readonly dir, got nil")
		}
	})

	t.Run("OutputDir does not exist", func(t *testing.T) {
		_, err := NewLimitedFileSystem([]string{readOnlyDir}, filepath.Join(tmpDir, "nonexistent_output"))
		if err == nil {
			t.Error("expected error for nonexistent output dir, got nil")
		}
	})

	t.Run("ReadOnly dir no read access", func(t *testing.T) {
		noAccessDir := filepath.Join(tmpDir, "noaccess")
		err := os.Mkdir(noAccessDir, 0000)
		if err != nil {
			t.Fatalf("failed to create no access dir: %v", err)
		}
		defer os.Chmod(noAccessDir, 0755) // cleanup to allow TempDir to be deleted

		_, err = NewLimitedFileSystem([]string{noAccessDir}, outputDir)
		if err == nil {
			t.Error("expected error for readonly dir with no read access, got nil")
		}
	})

	t.Run("OutputDir no write access", func(t *testing.T) {
		noWriteOutputDir := filepath.Join(tmpDir, "nowrite_output")
		err := os.Mkdir(noWriteOutputDir, 0555)
		if err != nil {
			t.Fatalf("failed to create no write output dir: %v", err)
		}
		defer os.Chmod(noWriteOutputDir, 0755) // cleanup

		_, err = NewLimitedFileSystem([]string{readOnlyDir}, noWriteOutputDir)
		if err == nil {
			t.Error("expected error for output dir with no write access, got nil")
		}
	})
}

func TestListFilesIn(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")
	outsideDir := filepath.Join(tmpDir, "outside")

	os.Mkdir(readOnlyDir, 0755)
	os.Mkdir(outputDir, 0755)
	os.Mkdir(outsideDir, 0755)

	os.WriteFile(filepath.Join(readOnlyDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(outputDir, "file2.txt"), []byte("content2"), 0644)
	os.WriteFile(filepath.Join(outsideDir, "file3.txt"), []byte("content3"), 0644)

	sandbox, err := NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	t.Run("List files in read-only dir", func(t *testing.T) {
		files, err := sandbox.ListFilesIn(readOnlyDir)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		expectedPath := filepath.Join(readOnlyDir, "file1.txt")
		if len(files) != 1 || files[0] != expectedPath {
			t.Errorf("expected [%s], got %v", expectedPath, files)
		}
	})

	t.Run("List files in output dir", func(t *testing.T) {
		files, err := sandbox.ListFilesIn(outputDir)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		expectedPath := filepath.Join(outputDir, "file2.txt")
		if len(files) != 1 || files[0] != expectedPath {
			t.Errorf("expected [%s], got %v", expectedPath, files)
		}
	})

	t.Run("Error for outside path", func(t *testing.T) {
		_, err := sandbox.ListFilesIn(outsideDir)
		if err == nil {
			t.Error("expected error for path outside sandbox, got nil")
		}
	})

	t.Run("Error for non-existent path", func(t *testing.T) {
		_, err := sandbox.ListFilesIn(filepath.Join(readOnlyDir, "nonexistent"))
		if err == nil {
			t.Error("expected error for non-existent path, got nil")
		}
	})

	t.Run("Error for file path (not a directory)", func(t *testing.T) {
		_, err := sandbox.ListFilesIn(filepath.Join(readOnlyDir, "file1.txt"))
		if err == nil {
			t.Error("expected error for file path, got nil")
		}
	})
}

func TestGetFileContentAsString(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")
	outsideDir := filepath.Join(tmpDir, "outside")

	os.Mkdir(readOnlyDir, 0755)
	os.Mkdir(outputDir, 0755)
	os.Mkdir(outsideDir, 0755)

	content1 := "content1"
	content2 := "content2"
	os.WriteFile(filepath.Join(readOnlyDir, "file1.txt"), []byte(content1), 0644)
	os.WriteFile(filepath.Join(outputDir, "file2.txt"), []byte(content2), 0644)
	os.WriteFile(filepath.Join(outsideDir, "file3.txt"), []byte("content3"), 0644)

	sandbox, err := NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	t.Run("Read file in read-only dir", func(t *testing.T) {
		content, err := sandbox.GetFileContentAsString(filepath.Join(readOnlyDir, "file1.txt"))
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if content != content1 {
			t.Errorf("expected %s, got %s", content1, content)
		}
	})

	t.Run("Read file in output dir", func(t *testing.T) {
		content, err := sandbox.GetFileContentAsString(filepath.Join(outputDir, "file2.txt"))
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if content != content2 {
			t.Errorf("expected %s, got %s", content2, content)
		}
	})

	t.Run("Error for outside path", func(t *testing.T) {
		_, err := sandbox.GetFileContentAsString(filepath.Join(outsideDir, "file3.txt"))
		if err == nil {
			t.Error("expected error for path outside sandbox, got nil")
		}
	})

	t.Run("Error for non-existent file", func(t *testing.T) {
		_, err := sandbox.GetFileContentAsString(filepath.Join(readOnlyDir, "nonexistent.txt"))
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})

	t.Run("Error for directory path", func(t *testing.T) {
		_, err := sandbox.GetFileContentAsString(readOnlyDir)
		if err == nil {
			t.Error("expected error for directory path, got nil")
		}
	})
}

func TestGetFileContentAsByte(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")

	os.Mkdir(readOnlyDir, 0755)
	os.Mkdir(outputDir, 0755)

	content1 := []byte("content1")
	os.WriteFile(filepath.Join(readOnlyDir, "file1.txt"), content1, 0644)

	sandbox, err := NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	t.Run("Read file in read-only dir", func(t *testing.T) {
		content, err := sandbox.GetFileContentAsByte(filepath.Join(readOnlyDir, "file1.txt"))
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if string(content) != string(content1) {
			t.Errorf("expected %s, got %s", string(content1), string(content))
		}
	})

	t.Run("Error for non-existent file", func(t *testing.T) {
		_, err := sandbox.GetFileContentAsByte(filepath.Join(readOnlyDir, "nonexistent.txt"))
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})
}

func TestWriteToFile(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")

	os.Mkdir(readOnlyDir, 0755)
	os.Mkdir(outputDir, 0755)

	sandbox, err := NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	t.Run("Write string to new file in output dir", func(t *testing.T) {
		path := filepath.Join(outputDir, "new_file.txt")
		content := "hello world"
		err := sandbox.WriteStringToFile(content, path)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify content
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read written file: %v", err)
		}
		if string(data) != content {
			t.Errorf("expected %s, got %s", content, string(data))
		}
	})

	t.Run("Write bytes to new file in output dir", func(t *testing.T) {
		path := filepath.Join(outputDir, "new_bytes.bin")
		content := []byte{0x01, 0x02, 0x03}
		err := sandbox.WriteBytesToFile(content, path)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify content
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read written file: %v", err)
		}
		if string(data) != string(content) {
			t.Errorf("expected %v, got %v", content, data)
		}
	})

	t.Run("Error when file already exists (string)", func(t *testing.T) {
		path := filepath.Join(outputDir, "existing.txt")
		os.WriteFile(path, []byte("existing"), 0644)

		err := sandbox.WriteStringToFile("new content", path)
		if err == nil {
			t.Error("expected error for existing file, got nil")
		}
	})

	t.Run("Error when file already exists (bytes)", func(t *testing.T) {
		path := filepath.Join(outputDir, "existing_bytes.bin")
		os.WriteFile(path, []byte("existing"), 0644)

		err := sandbox.WriteBytesToFile([]byte{0x01}, path)
		if err == nil {
			t.Error("expected error for existing file, got nil")
		}
	})

	t.Run("Error for path outside output dir (read-only dir)", func(t *testing.T) {
		path := filepath.Join(readOnlyDir, "trying_to_write.txt")
		err := sandbox.WriteStringToFile("content", path)
		if err == nil {
			t.Error("expected error for path in read-only dir, got nil")
		}
	})

	t.Run("Error for path completely outside sandbox", func(t *testing.T) {
		path := filepath.Join(tmpDir, "completely_outside.txt")
		err := sandbox.WriteStringToFile("content", path)
		if err == nil {
			t.Error("expected error for path outside sandbox, got nil")
		}
	})
}

func TestDeleteFile(t *testing.T) {
	tmpDir := t.TempDir()

	readOnlyDir := filepath.Join(tmpDir, "readonly")
	outputDir := filepath.Join(tmpDir, "output")

	os.Mkdir(readOnlyDir, 0755)
	os.Mkdir(outputDir, 0755)

	sandbox, err := NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	t.Run("Delete existing file in output dir", func(t *testing.T) {
		path := filepath.Join(outputDir, "to_delete.txt")
		os.WriteFile(path, []byte("goodbye"), 0644)

		err := sandbox.DeleteFile(path)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify file is gone
		if _, err := os.Stat(path); err == nil {
			t.Error("file still exists after deletion")
		}
	})

	t.Run("Error for path outside output dir (read-only dir)", func(t *testing.T) {
		path := filepath.Join(readOnlyDir, "cannot_delete.txt")
		os.WriteFile(path, []byte("protected"), 0644)

		err := sandbox.DeleteFile(path)
		if err == nil {
			t.Error("expected error for deleting from read-only dir, got nil")
		}

		// Verify file is still there
		if _, err := os.Stat(path); err != nil {
			t.Errorf("file should still exist: %v", err)
		}
	})

	t.Run("Error for path outside output dir (completely outside)", func(t *testing.T) {
		path := filepath.Join(tmpDir, "completely_outside.txt")
		os.WriteFile(path, []byte("outside"), 0644)

		err := sandbox.DeleteFile(path)
		if err == nil {
			t.Error("expected error for path outside sandbox, got nil")
		}
	})

	t.Run("Error for non-existent file", func(t *testing.T) {
		path := filepath.Join(outputDir, "non_existent.txt")
		err := sandbox.DeleteFile(path)
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})

	t.Run("Error for directory path", func(t *testing.T) {
		path := filepath.Join(outputDir, "some_dir")
		os.Mkdir(path, 0755)

		err := sandbox.DeleteFile(path)
		if err == nil {
			t.Error("expected error for directory path, got nil")
		}
	})
}
