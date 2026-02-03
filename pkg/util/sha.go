package util

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

func MakeSHA1(str ...string) string {
	h := sha1.New()
	val := ""
	for _, v := range str {
		val = fmt.Sprintf("%s%s", val, v)
	}
	h.Write([]byte(val))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func MakeSHA256(str ...string) string {
	h := sha256.New()
	val := ""
	for _, v := range str {
		val = fmt.Sprintf("%s%s", val, v)
	}
	h.Write([]byte(val))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func MakeSHA512(str ...string) string {
	h := sha512.New()
	val := ""
	for _, v := range str {
		val = fmt.Sprintf("%s%s", val, v)
	}
	h.Write([]byte(val))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func MakeMD5(str ...string) string {
	hash := md5.New()
	val := ""
	for _, v := range str {
		val = fmt.Sprintf("%s%s", val, v)
	}
	hash.Write([]byte(val))
	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

func BuildSecret(secretKey string, str ...string) string {
	var secret string
	for _, val := range str {
		secret = fmt.Sprintf("%s%s%s:", secret, MakeSHA1(secretKey), val)
	}
	return MakeSHA256(secret, secretKey)
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

type HashType int

const (
	SHA1 HashType = iota
	SHA256
	SHA512
)

func IsHash(str string, hashType HashType) bool {
	// Set the expected length based on hash type
	switch hashType {
	case SHA1:
		return IsSHA1(str)
	case SHA256:
		return IsSHA256(str)
	case SHA512:
		return IsSHA512(str)
	}
	return false
}

func IsSHA1(str string) bool {
	// Check if the length is exactly 128 characters
	if len(str) != 40 {
		return false
	}

	// Check if the string is a valid hexadecimal
	_, err := hex.DecodeString(str)
	return err == nil
}

func IsSHA256(str string) bool {
	// Check if the length is exactly 128 characters
	if len(str) != 64 {
		return false
	}

	// Check if the string is a valid hexadecimal
	_, err := hex.DecodeString(str)
	return err == nil
}

func IsSHA512(str string) bool {
	// Check if the length is exactly 128 characters
	if len(str) != 128 {
		return false
	}

	// Check if the string is a valid hexadecimal
	_, err := hex.DecodeString(str)
	return err == nil
}
