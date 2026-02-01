package util

import (
	"gitlab.com/threetopia/envgo"
)

func BuildJwtSecret(args ...string) string {
	return BuildSecret(envgo.GetString("APP_AUTHENTICATION_JWT_SECRET", "MustBeASecretYouKnow?!"), args...)
}
