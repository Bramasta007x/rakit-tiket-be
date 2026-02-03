package http

import (
	"fmt"
	"net/url"
	"strconv"

	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_file"
	model "rakit-tiket-be/pkg/model/http"
)

type (
	fileModel struct {
		entity.FileEntity
		Links model.LinkMap `json:"links"`
	}

	filesModel []fileModel
)

func MakeFileModel(id, name, path, mime string, data entity.FileData, relation model.RelationModel) fileModel {
	return fileModel{
		FileEntity: entity.FileEntity{
			ID:             pubEntity.UUID(id),
			Name:           name,
			Path:           path,
			Mime:           mime,
			Data:           data,
			RelationEntity: relation.RelationEntity,
		},
		Links: model.LinkMap{
			"GET": model.Link{
				Href: fmt.Sprintf("file/%s", id),
				Rel:  "file",
				Type: mime,
			},
			"PUT": model.Link{
				Href: fmt.Sprintf("file/%s", id),
				Rel:  "file",
				Type: mime,
			},
		},
	}
}

func MakeFileModelFromEntity(file entity.FileEntity) fileModel {
	return fileModel{
		FileEntity: file,
		Links: model.LinkMap{
			"GET": model.Link{
				Href: fmt.Sprintf("file/%s", file.ID.String()),
				Rel:  "file",
				Type: file.Mime,
			},
			"PUT": model.Link{
				Href: fmt.Sprintf("file/%s", file.ID.String()),
				Rel:  "file",
				Type: file.Mime,
			},
		},
	}
}

func MakeFilesModel(files ...fileModel) filesModel {
	var newFiles filesModel
	for _, file := range files {
		newFiles = append(newFiles,
			MakeFileModel(file.ID.String(), file.Name, file.Path, file.Mime, file.Data, model.MakeRelationModel(file.RelationID.String(), file.RelationSource)),
		)
	}
	return newFiles
}

func MakeFilesModelFromEntity(files entity.FilesEntity) filesModel {
	var newFiles filesModel
	for _, file := range files {
		newFiles = append(newFiles,
			MakeFileModel(file.ID.String(), file.Name, file.Path, file.Mime, file.Data, model.MakeRelationModel(file.RelationID.String(), file.RelationSource)),
		)
	}
	return newFiles
}

func (m fileModel) GetFileEntity() entity.FileEntity {
	return m.FileEntity
}

func (m filesModel) GetFilesEntity() entity.FilesEntity {
	var files entity.FilesEntity
	for _, fileModel := range m {
		files = append(files, fileModel.GetFileEntity())
	}
	return files
}

type (
	SearchFilesRequestModel struct {
		model.HTTPRequestModel
		entity.FileQuery
	}

	SearchFilesResponseModel struct {
		model.HTTPResponseModel
		Data filesModel
	}
)

func MakeSearchFilesRequestModel(ids, names, paths, mimes []string, fileData bool, paging model.PaginationModel) SearchFilesRequestModel {
	var uuids pubEntity.UUIDs
	for _, id := range ids {
		uuids = append(uuids, pubEntity.UUID(id))
	}
	return SearchFilesRequestModel{
		FileQuery: entity.FileQuery{
			IDs:         uuids,
			Names:       names,
			Paths:       paths,
			Mimes:       mimes,
			FileData:    fileData,
			PagingQuery: pubEntity.PagingQuery(paging),
		},
	}
}

func (r SearchFilesRequestModel) BuildUrlValues() url.Values {
	urlValues := url.Values{}
	if len(r.IDs) > 0 {
		for _, v := range r.IDs {
			urlValues.Add("id", v.String())
		}
	}
	if len(r.Deleted) > 0 {
		for _, v := range r.Deleted {
			urlValues.Add("deleted", strconv.FormatBool(v))
		}
	}
	return model.BuildDaoFieldUrlValues(r.HTTPRequestModel.BuildUrlValues(urlValues), r.DaoQuery)
}

func MakeSearchFilesResponseModel(httpCode, totalCount int, files entity.FilesEntity) (int, SearchFilesResponseModel) {
	return httpCode, SearchFilesResponseModel{
		HTTPResponseModel: model.MakeHTTPResponseModel(httpCode, totalCount, nil),
		Data:              MakeFilesModelFromEntity(files),
	}
}

type (
	GetFileByIDRequestModel struct {
		model.HTTPRequestModel
		entity.FileQuery
	}

	GetFileByIDResponseModel struct {
		model.HTTPResponseModel
		Data fileModel
	}
)

func MakeGetFileByIDRequestModel(ids, names, paths, mimes []string, fileData bool, paging model.PaginationModel) GetFileByIDRequestModel {
	var uuids pubEntity.UUIDs
	for _, id := range ids {
		uuids = append(uuids, pubEntity.UUID(id))
	}
	return GetFileByIDRequestModel{
		FileQuery: entity.FileQuery{
			IDs:         uuids,
			Names:       names,
			Paths:       paths,
			Mimes:       mimes,
			FileData:    fileData,
			PagingQuery: pubEntity.PagingQuery(paging),
		},
	}
}

func MakeGetFileByIDResponseModel(httpCode int, file entity.FileEntity) (int, GetFileByIDResponseModel) {
	return httpCode, GetFileByIDResponseModel{
		HTTPResponseModel: model.MakeHTTPResponseModel(httpCode, 1, nil),
		Data:              MakeFileModelFromEntity(file),
	}
}

type (
	InsertFileRequestModel struct {
		model.HTTPRequestModel
		entity.FileEntity
	}

	InsertFileResponseModel struct {
		model.HTTPResponseModel
		Data fileModel
	}
)

func MakeInsertFileRequestModel(name, path, mime string, data []byte, relation model.RelationModel) InsertFileRequestModel {
	return InsertFileRequestModel{
		FileEntity: entity.FileEntity{
			Name:           name,
			Path:           path,
			Mime:           mime,
			Data:           entity.FileData(data),
			RelationEntity: pubEntity.MakeRelationEntity(relation.RelationID, relation.RelationSource),
		},
	}
}

func MakeInsertFileResponseModel(httpCode int, file entity.FileEntity) (int, InsertFileResponseModel) {
	return httpCode, InsertFileResponseModel{
		HTTPResponseModel: model.MakeHTTPResponseModel(httpCode, 1, nil),
		Data:              MakeFileModelFromEntity(file),
	}
}

type (
	InsertFilesRequestModel struct {
		model.HTTPRequestModel
		entity.FilesEntity
	}

	InsertFilesResponseModel struct {
		model.HTTPResponseModel
		Data filesModel
	}
)

func MakeInsertFilesRequestModel(filesModel ...fileModel) InsertFilesRequestModel {
	var files entity.FilesEntity
	for _, fileModel := range filesModel {
		files = append(files, fileModel.FileEntity)
	}
	return InsertFilesRequestModel{
		FilesEntity: files,
	}
}

func MakeInsertFilesResponseModel(httpCode int, files entity.FilesEntity) (int, InsertFilesResponseModel) {
	return httpCode, InsertFilesResponseModel{
		HTTPResponseModel: model.MakeHTTPResponseModel(httpCode, len(files), nil),
		Data:              MakeFilesModelFromEntity(files),
	}
}

type (
	UpdateFileRequestModel struct {
		model.HTTPRequestModel
		entity.FileEntity
	}

	UpdateFileResponseModel struct {
		model.HTTPResponseModel
		Data fileModel
	}
)

func MakeUpdateFileRequestModel(name, path, mime string, data []byte, relation model.RelationModel) UpdateFileRequestModel {
	return UpdateFileRequestModel{
		FileEntity: entity.FileEntity{
			Name:           name,
			Path:           path,
			Mime:           mime,
			Data:           entity.FileData(data),
			RelationEntity: pubEntity.MakeRelationEntity(relation.RelationID, relation.RelationSource),
		},
	}
}

func MakeUpdateFileResponseModel(httpCode int, file entity.FileEntity) (int, UpdateFileResponseModel) {
	return httpCode, UpdateFileResponseModel{
		HTTPResponseModel: model.MakeHTTPResponseModel(httpCode, 1, nil),
		Data:              MakeFileModelFromEntity(file),
	}
}

type (
	UpdateFilesRequestModel struct {
		model.HTTPRequestModel
		entity.FilesEntity
	}

	UpdateFilesResponseModel struct {
		model.HTTPResponseModel
		Data filesModel
	}
)

func MakeUpdateFilesRequestModel(filesModel ...fileModel) UpdateFilesRequestModel {
	var files entity.FilesEntity
	for _, fileModel := range filesModel {
		files = append(files, fileModel.FileEntity)
	}
	return UpdateFilesRequestModel{
		FilesEntity: files,
	}
}

func MakeUpdateFilesResponseModel(httpCode int, files entity.FilesEntity) (int, UpdateFilesResponseModel) {
	return httpCode, UpdateFilesResponseModel{
		HTTPResponseModel: model.MakeHTTPResponseModel(httpCode, 1, nil),
		Data:              MakeFilesModelFromEntity(files),
	}
}

type (
	DeleteFileByIDRequestModel struct {
		model.HTTPRequestModel
		entity.FilesEntity
	}

	DeleteFileByIDResponseModel struct {
		model.HTTPResponseModel
		Data filesModel
	}
)

func MakeDeleteFileByIDRequestModel(fileID pubEntity.UUID) UpdateFileRequestModel {
	return UpdateFileRequestModel{
		FileEntity: entity.FileEntity{
			ID: fileID,
		},
	}
}

func MakeDeleteFileByIDResponseModel(httpCode int, file entity.FileEntity) (int, UpdateFileResponseModel) {
	return httpCode, UpdateFileResponseModel{
		HTTPResponseModel: model.MakeHTTPResponseModel(httpCode, 1, nil),
		Data:              MakeFileModelFromEntity(file),
	}
}

type (
	UpsertFilesRequestModel struct {
		model.HTTPRequestModel
		entity.FilesEntity
	}

	UpsertFilesResponseModel struct {
		model.HTTPResponseModel
		Data filesModel
	}
)

func MakeUpsertFilesRequestModel(filesModel ...fileModel) UpsertFilesRequestModel {
	var files entity.FilesEntity
	for _, FileModel := range filesModel {
		files = append(files, FileModel.FileEntity)
	}
	return UpsertFilesRequestModel{
		FilesEntity: files,
	}
}

func MakeUpsertFilesResponseModel(httpCode int, files entity.FilesEntity) (int, UpsertFilesResponseModel) {
	return httpCode, UpsertFilesResponseModel{
		HTTPResponseModel: model.MakeHTTPResponseModel(httpCode, 1, nil),
		Data:              MakeFilesModelFromEntity(files),
	}
}
