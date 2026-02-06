package entity

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type UserRole string

const (
	RoleAdmin       UserRole = "ADMIN"
	RoleGroundStaff UserRole = "GROUND STAFF"
)

type (
	UserQuery struct {
		IDs    pubEntity.UUIDs `query:"id"`
		Emails []string        `query:"email"`
	}

	UserEntity struct {
		ID           pubEntity.UUID `json:"id"`
		Name         string         `json:"name"`
		Email        string         `json:"email"`
		PasswordHash string         `json:"-"`
		Role         UserRole       `json:"role"`
		pubEntity.DaoEntity
	}

	UsersEntity []UserEntity
)
