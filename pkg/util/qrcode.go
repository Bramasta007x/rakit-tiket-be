package util

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/skip2/go-qrcode"
)

func GenerateQRCodeFile(content string, size int, outputPath string) (string, error) {
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	err = qrcode.WriteFile(content, qrcode.Medium, size, outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to write QR code file: %w", err)
	}

	return outputPath, nil
}

func GenerateQRCodeBase64(content string, size int) (string, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, size)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}
