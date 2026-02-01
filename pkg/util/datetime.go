package util

import (
	"time"

	"gitlab.com/threetopia/envgo"
)

const JakartaTimeZone string = "Asia/Jakarta"

func GetTimeZone() *time.Location {
	t, err := time.LoadLocation(envgo.GetString("TIME_ZONE", JakartaTimeZone))
	if err != nil {
		return time.UTC
	}
	return t
}
