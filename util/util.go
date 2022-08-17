package util

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func CreateFile(fileKey string) (*os.File, error) {
	if err := createDir(fileKey); err != nil {
		return nil, err
	}

	f, err := os.Create(fileKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create a file for key: %s. %s", fileKey, err)
	}

	return f, nil
}

func createDir(key string) error {
	if err := os.MkdirAll(path.Dir(key), os.FileMode(0744)); err != nil {
		return fmt.Errorf("failed to create dir for key: %s. %s", key, err)
	}
	return nil
}

func RemoveSpaces(s string) string {
	return strings.ReplaceAll(s, " ", "")
}
