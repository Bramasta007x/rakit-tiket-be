package entity

import (
	"fmt"
	"strings"

	"rakit-tiket-be/pkg/util"

	pubEntity "rakit-tiket-be/pkg/entity"

	"gitlab.com/threetopia/envgo"
)

type FileData []byte

func (e FileData) EncodeToBase64() string {
	return util.EncodeToBase64(e)
}

func (e *FileData) DecodeFromBase64(encStr string) {
	*e, _ = util.DecodeFromBase64(encStr)
}

func (e FileData) buildSecret(secrets ...string) string {
	cryptSecret := envgo.GetString("ENCRYPTION_SECRET", "1234567890abcdef")
	if len(secrets) > 0 {
		cryptSecret = fmt.Sprintf("%s:%s", strings.Join(secrets, ":"), cryptSecret)
	}
	return util.MakeMD5(cryptSecret)
}

type (
	FileQuery struct {
		IDs      pubEntity.UUIDs `query:"id"`
		Names    []string        `query:"name"`
		Paths    []string        `query:"path"`
		Mimes    []string        `query:"mime"`
		FileData bool            `query:"file_data"`
		pubEntity.RelationQuery
		pubEntity.DaoQuery
		pubEntity.PagingQuery
	}

	FileEntity struct {
		ID          pubEntity.UUID `json:"id"`
		Name        string         `json:"name"`
		Path        string         `json:"path"`
		Mime        string         `json:"mime"`
		Description string         `json:"description"`
		Data        FileData       `json:"data"`
		pubEntity.RelationEntity
		pubEntity.DaoEntity
	}
)

func (file FileEntity) MakeDataHash() pubEntity.Hash {
	return file.DaoEntity.MakeDataHash(
		file.RelationID.String(),
		file.RelationSource,
		file.Name,
		file.Mime,
		file.Path,
	)
}

type (
	FilesEntity    []FileEntity
	FileMapEntity  map[string]FileEntity
	FilesMapEntity map[string]FilesEntity
)

func (files FilesEntity) GetIDs() pubEntity.UUIDs {
	var ids pubEntity.UUIDs
	for _, file := range files {
		ids = append(ids, file.ID)
	}
	return ids
}

func (files FilesEntity) ExportByRelation() FilesMapEntity {
	var filesMap FilesMapEntity
	for _, file := range files {
		if filesMap == nil {
			filesMap = make(FilesMapEntity)
		}

		mapKey := file.GetRelationMapKey()
		if _, ok := filesMap[mapKey]; !ok {
			filesMap[mapKey] = make(FilesEntity, 0)
		}
		filesMap[mapKey] = append(filesMap[mapKey], file)
	}
	return filesMap
}

func (files FilesEntity) GetRelationIDs() pubEntity.UUIDs {
	var relationIDs pubEntity.UUIDs
	for _, file := range files {
		relationIDs = append(relationIDs, file.RelationID)
	}
	return relationIDs
}

func (files FilesEntity) GetRelationSource() []string {
	var relationSource []string
	for _, file := range files {
		relationSource = append(relationSource, file.RelationSource)
	}
	return relationSource
}

func (files FilesEntity) ExportByIDAndDataHash() FileMapEntity {
	fileMap := make(FileMapEntity)
	for _, file := range files {
		if file.ID != "" {
			fileMap[file.ID.String()] = file
		}
		fileMap[file.MakeDataHash().String()] = file
	}
	return fileMap
}

func (files FilesEntity) ExportByIDRelationAndDataHash() FileMapEntity {
	fileMap := make(FileMapEntity)
	for _, file := range files {
		if file.ID != "" {
			fileMap[file.ID.String()] = file
		}
		fileMap[file.GetRelationMapKey()] = file
		fileMap[file.MakeDataHash().String()] = file
	}
	return fileMap
}

func (files FilesEntity) ExportByDataHash() FileMapEntity {
	fileMap := make(FileMapEntity)
	for _, file := range files {
		fileMap[file.MakeDataHash().String()] = file
	}
	return fileMap
}
