package entity

type (
	PingEntity struct {
		AppName string `json:"appName"`
		Pong    bool   `json:"pong"`

		Apps PingMapEntity `json:"apps,omitempty"`
	}

	PingMapEntity map[string]PingEntity
)

func (e PingMapEntity) GetByKey(key string) PingEntity {
	if ping, ok := e[key]; ok {
		return ping
	}
	return PingEntity{}
}
