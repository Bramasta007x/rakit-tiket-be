package util

import (
	"crypto/sha512"
	"strings"

	"github.com/google/uuid"
	"gitlab.com/threetopia/envgo"
)

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func MakeUUID(strs ...string) string {
	return MakeUUIDv5(strs...)
}

func MakeUUIDv5(strs ...string) string {
	nameSpace := uuid.NewSHA1(uuid.NameSpaceOID, []byte(envgo.GetString("UUID_SECRET", "MustBeSecret")))
	return uuid.NewHash(sha512.New(), nameSpace, []byte(strings.Join(strs, ".")), int(uuid.NameSpaceOID.Version())).String()
}

func MakeUUIDv4() string {
	return uuid.NewString()
}
