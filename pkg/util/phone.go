package util

import (
	"fmt"
	"strings"
)

func FormatPhone(phoneNumber string, prefix int64) string {
	if strings.HasPrefix(phoneNumber, "0") {
		phoneNumber = strings.TrimPrefix(phoneNumber, "0")
		return fmt.Sprintf("%d%s", prefix, phoneNumber)
	}
	return phoneNumber
}
