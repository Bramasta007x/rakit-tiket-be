package entity

import (
	"fmt"

	pubConstant "rakit-tiket-be/pkg/constant"
)

// Relation
type (
	RelationQuery struct {
		RelationIDs     UUIDs    `query:"relation_id" validate:"required"`
		RelationSources []string `query:"relation_source" validate:"required"`
	}

	RelationEntity struct {
		RelationID     UUID   `gorm:"type:uuid" json:"relationId,omitempty"`
		RelationSource string `gorm:"type:varchar;size:25" json:"relationSource,omitempty"`
	}

	RelationEntitySlice []RelationEntity
)

func MakeRelationMapKey(relationID, relationSource string) string {
	return fmt.Sprintf("%s;%s", relationID, relationSource)
}

func MakeRelationEntity(relationID UUID, relationSource string) RelationEntity {
	return RelationEntity{
		RelationID:     relationID,
		RelationSource: relationSource,
	}
}

func (ownershipFields RelationEntity) SetRelationID(defaultUUID UUID) UUID {
	if ownershipFields.RelationID != "" && ownershipFields.RelationID != UUID(pubConstant.DefaultRelationID) {
		return ownershipFields.RelationID
	}
	return defaultUUID
}

func (ownershipFields RelationEntity) SetOwnerSource(defaultOwnerSource string) string {
	if ownershipFields.RelationSource != "" && ownershipFields.RelationSource != pubConstant.DefaultRelationSource {
		return ownershipFields.RelationSource
	}
	return defaultOwnerSource
}

func (relation RelationEntity) GetRelationMapKey() string {
	return MakeRelationMapKey(relation.RelationID.String(), relation.RelationSource)
}
