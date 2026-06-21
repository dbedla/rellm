package agentsutills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LimitedFileSystem struct {
	readOnlyDirs []string
	outputDir    string
}

// NewLimitedFileSystem should return a new LimitedFileSystem instance
// error will be returned if any of readOnlyDirs dirs do not exist
// error will be returned if outputDir does not exist.
// error will be returned if the program does not have read access to readOnlyDirs
// error will be returned if the program does not have read and write access to outputDir
func NewLimitedFileSystem(readOnlyDirs []string, outputDir string) (*LimitedFileSystem, error) {
	if err := validateReadOnlyDirs(readOnlyDirs); err != nil {
		return nil, err
	}

	if err := validateOutputDir(outputDir); err != nil {
		return nil, err
	}

	return &LimitedFileSystem{
		readOnlyDirs: readOnlyDirs,
		outputDir:    outputDir,
	}, nil
}

// GetReadOnlyPaths returns a list of paths that are allowed to be read-only from
func (s *LimitedFileSystem) GetReadOnlyPaths() []string {
	return s.readOnlyDirs
}

// GetOutputDir returns the path of the output directory
// in this directory read, write and delete operation are allowed
func (s *LimitedFileSystem) GetOutputDir() string {
	return s.outputDir
}

// GetFileContentAsString returns the content of the file at the given path as a string
// error will be returned if the path is outside FSSandbox.readOnlyDirs or FSSandbox.outputDir
// error will be returned if the path does not exist or is not accessible.
func (s *LimitedFileSystem) GetFileContentAsString(path string) (string, error) {
	absPath, err := s.validatePath(path, false)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return string(content), nil
}

// GetFileContentAsByte returns the content of the file at the given path as a []byte
// error will be returned if the path is outside FSSandbox.readOnlyDirs or FSSandbox.outputDir
// error will be returned if the path does not exist or is not accessible.
func (s *LimitedFileSystem) GetFileContentAsByte(path string) ([]byte, error) {
	absPath, err := s.validatePath(path, false)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return content, nil
}

// WriteBytesToFile writes bytes into a specific file.
// if the file does not exist, it will be created
// error will be if the file already exists
// error will be returned if the path is outside FSSandbox.outputDir
func (s *LimitedFileSystem) WriteBytesToFile(content []byte, path string) error {
	absPath, err := s.validateWritePath(path)
	if err != nil {
		return err
	}

	err = os.WriteFile(absPath, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", path, err)
	}

	return nil
}

// DeleteFile delete file.
// error will be returned if the path is outside FSSandbox.outputDir
// error will be returned in any other standard case during deletion
func (s *LimitedFileSystem) DeleteFile(path string) error {
	absPath, err := s.validateInOutputDir(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %w", path, err)
	}

	if info.IsDir() {
		return fmt.Errorf("path %s is a directory, not a file", path)
	}

	err = os.Remove(absPath)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}

	return nil
}

// ListFilesIn returns a list of files in the given path
// error will be returned if the path does not exist or is not accessible.
// error will be returned if the path is not a directory
// error will be returned if the path is outside FSSandbox.readOnlyDirs or FSSandbox.outputDir
func (s *LimitedFileSystem) ListFilesIn(path string) ([]string, error) {
	absPath, err := s.validatePath(path, true)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	return filterFiles(absPath, entries), nil
}

// WriteStringToFile writes a given string into a specific file.
// if the file does not exist, it will be created
// error will be if the file already exists
// error will be returned if the path is outside FSSandbox.outputDir
func (s *LimitedFileSystem) WriteStringToFile(content, path string) error {
	absPath, err := s.validateWritePath(path)
	if err != nil {
		return err
	}

	err = os.WriteFile(absPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", path, err)
	}

	return nil
}

func validateReadOnlyDirs(dirs []string) error {
	for _, dir := range dirs {
		if err := checkReadAccess(dir); err != nil {
			return err
		}
	}
	return nil
}

func validateOutputDir(dir string) error {
	if err := checkReadAccess(dir); err != nil {
		return err
	}
	return checkWriteAccess(dir)
}

func checkReadAccess(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("directory %s does not exist or is not accessible: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("directory %s is not readable: %w", path, err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("file %s is not readable: %w", path, err)
	}
	return nil
}

func checkWriteAccess(path string) error {
	testFile := filepath.Join(path, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("directory %s is not writable: %w", path, err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("file %s is not writable: %w", path, err)
	}

	err = os.Remove(testFile)
	if err != nil {
		return fmt.Errorf("failed to remove test file: %w", err)
	}
	return nil
}

func (s *LimitedFileSystem) validatePath(path string, shouldBeDir bool) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	if !s.isAllowed(absPath) {
		return "", fmt.Errorf("path %s is outside of allowed directories", path)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("path %s does not exist or is not accessible: %w", path, err)
	}

	if shouldBeDir && !info.IsDir() {
		return "", fmt.Errorf("path %s is not a directory", path)
	}
	if !shouldBeDir && info.IsDir() {
		return "", fmt.Errorf("path %s is a directory, not a file", path)
	}

	return absPath, nil
}

func (s *LimitedFileSystem) validateWritePath(path string) (string, error) {
	absPath, err := s.validateInOutputDir(path)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(absPath); err == nil {
		return "", fmt.Errorf("file %s already exists", path)
	}

	return absPath, nil
}

func (s *LimitedFileSystem) validateInOutputDir(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	absOutputDir, err := filepath.Abs(s.outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute output directory path: %w", err)
	}

	if !strings.HasPrefix(absPath, absOutputDir) {
		return "", fmt.Errorf("path %s is outside of allowed output directory", path)
	}

	return absPath, nil
}

func filterFiles(basePath string, entries []os.DirEntry) []string {
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(basePath, entry.Name()))
		}
	}
	return files
}

func (s *LimitedFileSystem) isAllowed(path string) bool {
	allowedDirs := append(s.readOnlyDirs, s.outputDir)
	for _, dir := range allowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if strings.HasPrefix(path, absDir) {
			return true
		}
	}
	return false
}
