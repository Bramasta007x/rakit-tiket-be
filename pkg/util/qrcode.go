package util

import (
	"encoding/base64"

	"github.com/skip2/go-qrcode"
)

// GenerateQRCodeBase64 menghasilkan string base64 dari QR Code
func GenerateQRCodeBase64(content string, size int) (string, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, size)
	if err != nil {
		return "", err
	}
	// Format untuk img src HTML
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}
