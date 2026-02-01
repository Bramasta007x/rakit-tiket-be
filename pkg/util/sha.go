package util

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

func SHA1(str ...string) string {
	h := sha1.New()
	val := ""
	for _, v := range str {
		val = fmt.Sprintf("%s%s", val, v)
	}
	h.Write([]byte(val))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func SHA256(str ...string) string {
	h := sha256.New()
	val := ""
	for _, v := range str {
		val = fmt.Sprintf("%s%s", val, v)
	}
	h.Write([]byte(val))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func SHA512(str ...string) string {
	h := sha512.New()
	val := ""
	for _, v := range str {
		val = fmt.Sprintf("%s%s", val, v)
	}
	h.Write([]byte(val))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func BuildSecret(secretKey string, str ...string) string {
	var secret string
	for _, val := range str {
		secret = fmt.Sprintf("%s%s;", secret, val)
	}
	return SHA256(secret, secretKey)
}

func SHANum(byteLen int, str ...string) int64 {
	h := sha512.New()
	val := ""
	for _, v := range str {
		val = fmt.Sprintf("%s%s", val, v)
	}
	h.Write([]byte(val))
	bs := h.Sum(nil)
	return BytesToInt64(bs, byteLen)
}

func PBKDF2(keyLen int, secretKey string, str ...string) string {
	var salt string
	for _, v := range str {
		salt = fmt.Sprintf("%s%s", salt, v)
	}
	return string(pbkdf2.Key([]byte(secretKey), []byte(salt), 4096, keyLen, sha256.New))
}
