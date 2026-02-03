package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"rakit-tiket-be/internal/app/app_file/dao"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_file"
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/envgo"
	"go.uber.org/zap"
)

type FileService interface {
	Search(ctx context.Context, query entity.FileQuery) (entity.FilesEntity, int, error)
	GetByID(ctx context.Context, fileID pubEntity.UUID) (entity.FileEntity, error)
	Insert(ctx context.Context, file *entity.FileEntity) error
	Inserts(ctx context.Context, files entity.FilesEntity) error
	Update(ctx context.Context, file *entity.FileEntity) error
	Updates(ctx context.Context, files entity.FilesEntity) error
	DeleteByID(ctx context.Context, fileID pubEntity.UUID) error
	Upsert(ctx context.Context, files entity.FilesEntity) error
}

type fileService struct {
	log   util.LogUtil
	sqlDB *sql.DB
}

func MakeFileService(log util.LogUtil, sqlDB *sql.DB) FileService {
	return fileService{
		log:   log,
		sqlDB: sqlDB,
	}
}

func (s fileService) Search(ctx context.Context, query entity.FileQuery) (entity.FilesEntity, int, error) {

	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	files, totalCount, err := dbTrx.GetFileDao().Search(ctx, query)
	if err != nil {
		s.log.Error(ctx, "fileService.Search", zap.Error(err))
		return nil, 0, err
	} else if len(files) < 1 {
		return nil, 0, fmt.Errorf("file(s) not found")
	}

	if query.FileData {
		for i, file := range files {
			savePath := fmt.Sprintf("%s/%s.ref", file.Path, file.ID)

			// Get saved encrypted file content
			file.Data, err = util.ReadFile(savePath)
			if err != nil {
				return nil, 0, err
			}

			// Get saved file content
			files[i] = file
		}
	}

	return files, totalCount, nil
}

func (s fileService) GetByID(ctx context.Context, fileID pubEntity.UUID) (entity.FileEntity, error) {

	var file entity.FileEntity
	files, _, err := s.Search(ctx, entity.FileQuery{
		IDs:      pubEntity.UUIDs{fileID},
		FileData: true,
	})
	if err != nil {
		return file, err
	} else if len(files) < 1 {
		return file, fmt.Errorf("record(s) not found")
	}
	file = files[0]

	return file, nil
}

func (s fileService) Insert(ctx context.Context, file *entity.FileEntity) error {

	files := entity.FilesEntity{*file}
	if err := s.Inserts(ctx, files); err != nil {
		return err
	}

	*file = files[0]
	return nil
}

func (s fileService) Inserts(ctx context.Context, files entity.FilesEntity) error {
	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()
	for idx, file := range files {
		if file.RelationSource == "" || file.RelationID == "" {
			s.log.Error(ctx, "fileService.Inserts.emptyRelation", zap.Error(fmt.Errorf("empty relation(s) data")))
			return fmt.Errorf("empty relation(s) data")
		}
		file.Mime = http.DetectContentType(file.Data)
		file.Path = fmt.Sprintf("%s/%s/%s", s.getFilePath(), file.RelationSource, file.RelationID)
		files[idx] = file
	}

	if err := dbTrx.GetFileDao().Insert(ctx, files); err != nil {
		s.log.Error(ctx, "fileService.Inserts.Insert", zap.Error(err))
		return err
	}

	for _, file := range files {
		// Save encrypted file content
		fileName := fmt.Sprintf("%s.ref", file.ID)
		if err := util.SaveFile(file.Path, fileName, file.Data); err != nil {
			s.log.Error(ctx, "fileService.Inserts.SaveFile", zap.Error(err))
			return err
		}
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		s.log.Error(ctx, "fileService.Inserts.Commit", zap.Error(err))
		return err
	}

	return nil
}

func (s fileService) Update(ctx context.Context, file *entity.FileEntity) error {
	files := entity.FilesEntity{*file}
	if err := s.Updates(ctx, files); err != nil {
		return err
	}

	*file = files[0]
	return nil
}

func (s fileService) Updates(ctx context.Context, files entity.FilesEntity) error {
	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()
	for idx, file := range files {
		if file.RelationSource == "" || file.RelationID == "" {
			return fmt.Errorf("empty relation(s) data")
		}
		file.Mime = http.DetectContentType(file.Data)
		file.Path = fmt.Sprintf("%s/%s/%s", s.getFilePath(), file.RelationSource, file.RelationID)
		files[idx] = file
	}

	if err := dbTrx.GetFileDao().Update(ctx, files); err != nil {
		return err
	}

	for _, file := range files {
		// Save encrypted file content
		fileName := fmt.Sprintf("%s.ref", file.ID)
		if err := util.SaveFile(file.Path, fileName, []byte(file.Data)); err != nil {
			return err
		}
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}

func (s fileService) DeleteByID(ctx context.Context, fileID pubEntity.UUID) error {
	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()

	if err := dbTrx.GetFileDao().Delete(ctx, pubEntity.UUIDs{pubEntity.UUID(fileID)}); err != nil {
		return err
	}
	return nil
}

func (s fileService) getFilePath() string {
	return envgo.GetString("APP_FILE_PATH", "../../../.tmp")
}

func (s fileService) Upsert(ctx context.Context, files entity.FilesEntity) error {
	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	defer dbTrx.GetSqlTx().Rollback()
	for idx, file := range files {
		if file.RelationSource == "" || file.RelationID == "" {
			return fmt.Errorf("empty relation(s) data")
		}
		file.Mime = http.DetectContentType(file.Data)
		file.Path = fmt.Sprintf("%s/%s/%s", s.getFilePath(), file.RelationSource, file.RelationID)
		files[idx] = file
	}

	if err := dbTrx.GetFileDao().Upsert(ctx, files); err != nil {
		return err
	}

	for _, file := range files {
		// Save encrypted file content
		fileName := fmt.Sprintf("%s.ref", file.ID)
		if err := util.SaveFile(file.Path, fileName, []byte(file.Data)); err != nil {
			return err
		}
	}

	if err := dbTrx.GetSqlTx().Commit(); err != nil {
		return err
	}

	return nil
}
