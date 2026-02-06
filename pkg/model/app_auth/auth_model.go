package model

import (
	model "rakit-tiket-be/pkg/model/http"
)

type (
	LoginRequestModel struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	LoginResponseModel struct {
		model.HTTPResponseModel
		Token string `json:"token"`
	}
)

func MakeLoginResponseModel(httpCode int, token string) (int, LoginResponseModel) {
	return httpCode, LoginResponseModel{
		HTTPResponseModel: model.MakeHTTPResponseModel(httpCode, 1, nil),
		Token:             token,
	}
}
