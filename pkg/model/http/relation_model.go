package model

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type RelationModel struct {
	pubEntity.RelationEntity
}

func MakeRelationModel(relationID, relationSource string) RelationModel {
	return RelationModel{
		RelationEntity: pubEntity.MakeRelationEntity(pubEntity.UUID(relationID), relationSource),
	}
}
