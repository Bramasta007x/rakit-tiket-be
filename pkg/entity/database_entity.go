package entity

import "fmt"

type DSNEntity struct {
	Host     string
	User     string
	Password string
	Port     int
	SSLMode  bool
	Database string
	Charset  string
	TimeZone string
	Schema   string
}

func (m DSNEntity) GetPostgresParam() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s%s",
		m.Host, m.User, m.Password, m.Database, m.Port, func(sslMode bool) string {
			if sslMode {
				return "require"
			}
			return "disable"
		}(m.SSLMode), m.TimeZone, m.getSchema(" "))
}

func (m DSNEntity) GetPostgresURI() string {
	return fmt.Sprintf("postgresql://%s:%d/%s?user=%s&password=%s%s",
		m.Host, m.Port, m.Database, m.User, m.Password, m.getSchema("&"))
}

func (m DSNEntity) getSchema(prefix string) string {
	if m.Schema != "" {
		return fmt.Sprintf("%ssearch_path=%s", prefix, m.Schema)
	}
	return ""
}
