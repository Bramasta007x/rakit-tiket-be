package entity

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type ArtistSocialMedia struct {
	Link string `json:"link"`
	Name string `json:"name"`
}

type ArtistQuery struct {
	IDs    []string `query:"id"`
	Names  []string `query:"name"`
	Genres []string `query:"genre"`
}

type Artist struct {
	ID                pubEntity.UUID      `json:"id"`
	Image             *string             `json:"image"`
	ImageUrl          *string             `json:"imageUrl"`
	Name              string              `json:"name"`
	Genre             string              `json:"genre"`
	ArtistSocialMedia []ArtistSocialMedia `json:"artist_social_media"`

	pubEntity.DaoEntity
}

type Artists []Artist
