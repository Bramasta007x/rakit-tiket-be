package util

import (
	"fmt"
	"os"
	"path/filepath"
)

func SaveFile(filePath, fileName string, data []byte) error {
	savePath := fmt.Sprintf("%s/%s", filePath, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filePath, 0755); err != nil {
			return err
		}
	}
	err := os.WriteFile(savePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func ReadFile(savePath string) ([]byte, error) {
	data, err := os.ReadFile(savePath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func RemoveFile(savePath string) error {
	if err := os.Remove(savePath); err != nil {
		return err
	}
	return nil
}

type (
	File struct {
		FilePath string
		IsDir    bool
		Deep     int
	}

	Files []File
)

func ScanDir(filePath string, recursive bool, deep int) Files {
	var result Files
	entries, err := os.ReadDir(filePath)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		filePath := filepath.Join(filePath, e.Name())
		result = append(result, File{
			FilePath: filePath,
			IsDir:    e.IsDir(),
			Deep:     deep,
		})
		if recursive && e.IsDir() {
			rScanDir := ScanDir(filePath, recursive, deep+1)
			result = append(result, rScanDir...)
		}
	}

	return result
}
