package utils

import (
	"os"
	"path"
)

func CreateDirInSysTmp(prefix string) (string, error) {
	pattern := prefix + "-*"

	tempDir, err := os.MkdirTemp("", pattern)
	if err != nil {
		return "", err
	}

	return tempDir, nil
}

func CreateSubDir(basePath, subDirName string) (string, error) {
	newDirPath := path.Join(basePath, subDirName)
	err := os.Mkdir(newDirPath, 0755)
	if err != nil {
		return "", err
	}

	return newDirPath, nil
}
