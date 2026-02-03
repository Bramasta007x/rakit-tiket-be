package model

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type PaginationModel pubEntity.PagingQuery

func MakePaginationModel(page int, limit int, noLimit bool) PaginationModel {
	return PaginationModel{
		Page:    pubEntity.Page(page),
		Limit:   pubEntity.Limit(limit),
		NoLimit: noLimit,
	}
}
