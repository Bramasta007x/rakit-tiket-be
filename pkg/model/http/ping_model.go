package model

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type (
	pingModel struct {
		pubEntity.PingEntity
		Links LinkMap `json:"links"`
	}

	pingsModel []pingModel
)

type (
	PingRequestModel struct {
		CallbackURI string `query:"callback_uri"`
	}

	PingResponseModel struct {
		HTTPResponseModel
		Data pubEntity.PingEntity
	}
)

func MakePingResponseModel(httpCode int, pong pubEntity.PingEntity) (int, PingResponseModel) {
	return httpCode, PingResponseModel{
		HTTPResponseModel: MakeHTTPResponseModel(httpCode, 1, nil),
		Data:              pong,
	}
}
