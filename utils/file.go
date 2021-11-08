package utils

import (
	"os"
)

// Check if file or directory exists
func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Delete file or directory
func Delete(filePath string) error {
	err := os.RemoveAll(filePath)
	if err != nil {
		return err
	}
	return nil
}
