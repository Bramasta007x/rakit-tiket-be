package dao

import (
	"context"
	"fmt"
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_file"

	"gitlab.com/threetopia/sqlgo/v2"
)

type FileDAO interface {
	Search(ctx context.Context, query entity.FileQuery) (entity.FilesEntity, int, error)
	Insert(ctx context.Context, files entity.FilesEntity) error
	Update(ctx context.Context, files entity.FilesEntity) error
	Delete(ctx context.Context, ids pubEntity.UUIDs) error
	Upsert(ctx context.Context, files entity.FilesEntity) error
}

type fileDAO struct {
	dbTrx DBTransaction
}

func MakeFileDAO(dbTrx DBTransaction) FileDAO {
	return fileDAO{
		dbTrx: dbTrx,
	}
}

func (d fileDAO) Search(ctx context.Context, query entity.FileQuery) (entity.FilesEntity, int, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("f.id", "id").
		SetSQLSelect("f.relation_id", "relation_id").
		SetSQLSelect("f.relation_source", "relation_source").
		SetSQLSelect("f.file_name", "file_name").
		SetSQLSelect("f.file_path", "file_path").
		SetSQLSelect("f.file_mime", "file_mime").
		SetSQLSelect("f.description", "description").
		SetSQLSelect("f.data_hash", "data_hash").
		SetSQLSelect("f.created_at", "created_at").
		SetSQLSelect("f.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom(`"file"`, "f")

	sqlWhere := sqlgo.NewSQLGoWhere()

	if len(query.IDs) > 0 {
		// Pastikan perubahan ini tersimpan dan server di-restart
		sqlWhere.SetSQLWhere("AND", "f.id", "IN", query.IDs)
	}
	if len(query.RelationQuery.RelationIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "f.relation_id", "IN", query.RelationQuery.RelationIDs.Strings())
	}
	if len(query.RelationQuery.RelationSources) > 0 {
		sqlWhere.SetSQLWhere("AND", "f.relation_source", "IN", query.RelationQuery.RelationSources)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)
	if err != nil {
		return nil, 0, err
	}

	var files entity.FilesEntity
	for rows.Next() {
		var file entity.FileEntity

		if err := rows.Scan(
			&file.ID,
			&file.RelationEntity.RelationID,
			&file.RelationEntity.RelationSource,
			&file.Name,
			&file.Path,
			&file.Mime,
			&file.Description,
			&file.DaoEntity.DataHash,
			&file.DaoEntity.CreatedAt,
			&file.DaoEntity.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		files = append(files, file)
	}

	totalCount, err := d.Count(ctx, query)

	if err != nil {
		return nil, 0, err
	}

	return files, totalCount, nil
}

func (d fileDAO) Count(ctx context.Context, query entity.FileQuery) (int, error) {

	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("COUNT(f.id)", "count")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom(`"file"`, "f")

	sqlWhere := sqlgo.NewSQLGoWhere()

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "f.id", "IN", query.IDs)
	}

	if len(query.RelationQuery.RelationIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "f.relation_id", "IN", query.RelationQuery.RelationIDs)
	}

	if len(query.Names) > 0 {
		sqlWhere.SetSQLWhere("AND", "f.file_name", "IN", query.Names)
	}

	if len(query.Mimes) > 0 {
		sqlWhere.SetSQLWhere("AND", "f.file_mime", "IN", query.Mimes)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	var total int
	err := d.dbTrx.GetSqlDB().QueryRowContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	).Scan(&total)

	if err != nil {
		return 0, err
	}

	return total, nil
}

func (d fileDAO) Insert(ctx context.Context, files entity.FilesEntity) error {

	if len(files) < 1 {
		return fmt.Errorf("empty file data")
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("file").
		SetSQLInsertColumn(
			"id",
			"relation_id",
			"relation_source",
			"file_name",
			"file_path",
			"file_mime",
			"description",
			"data_hash",
			"created_at",
		)

	for i, file := range files {
		file.DataHash = file.MakeDataHash()
		file.CreatedAt = time.Now()
		file.ID = pubEntity.MakeUUID(file.DataHash.String(), file.CreatedAt.String())

		sqlInsert.SetSQLInsertValue(
			file.ID,
			file.RelationEntity.RelationID,
			file.RelationEntity.RelationSource,
			file.Name,
			file.Path,
			file.Mime,
			file.Description,
			file.DataHash,
			file.CreatedAt,
		)

		files[i] = file
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoInsert(sqlInsert)

		// Debugging: Print SQL query jika error masih muncul
	fmt.Println("DEBUG SQL:", sql.BuildSQL())

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	if err != nil {
		return err
	}

	return nil
}

func (d fileDAO) Update(ctx context.Context, files entity.FilesEntity) error {

	if len(files) < 1 {
		return fmt.Errorf("empty file data")
	}

	for i, file := range files {
		now := time.Now()
		file.DataHash = file.MakeDataHash()
		file.UpdatedAt = &now

		sql := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("file").
			SetSQLUpdateValue("file_name", file.Name).
			SetSQLUpdateValue("file_path", file.Path).
			SetSQLUpdateValue("file_mime", file.Mime).
			SetSQLUpdateValue("description", file.Description).
			SetSQLUpdateValue("data_hash", file.DataHash).
			SetSQLUpdateValue("updated_at", file.UpdatedAt).
			SetSQLWhere("AND", "id", "=", file.ID)

		_, err := d.dbTrx.GetSqlTx().ExecContext(
			ctx,
			sql.BuildSQL(),
			sql.GetSQLGoParameter().GetSQLParameter()...,
		)
		if err != nil {
			return err
		}

		files[i] = file
	}

	return nil
}

func (d fileDAO) Delete(ctx context.Context, ids pubEntity.UUIDs) error {

	if len(ids) < 1 {
		return fmt.Errorf("empty file id")
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLDelete("file").
		SetSQLWhere("AND", "id", "IN", ids)

	_, err := d.dbTrx.GetSqlTx().ExecContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)

	return err
}

func (d fileDAO) Upsert(ctx context.Context, files entity.FilesEntity) error {
	if len(files) < 1 {
		return nil
	}

	searchFiles, _, err := d.Search(ctx, entity.FileQuery{
		RelationQuery: pubEntity.RelationQuery{
			RelationIDs:     files.GetRelationIDs(),
			RelationSources: files.GetRelationSource(),
		},
	})
	if err != nil {
		return err
	}

	inputMap := files.ExportByIDAndDataHash()
	databaseMap := searchFiles.ExportByIDAndDataHash()

	var insertFiles, updateFiles, deleteFiles entity.FilesEntity
	processedData := make(map[string]bool)

	for _, file := range append(files, searchFiles...) {
		fileKey := getFileKey(file)
		if _, ok := processedData[fileKey]; ok {
			continue
		}
		processedData[fileKey] = true

		// Flag for data existence in input and database
		var inInput, inDatabase bool
		var inputData, databaseData entity.FileEntity
		if inputData, inInput = isFileInMap(file, inputMap); inInput {
			file.RelationID = inputData.RelationID
			file.RelationSource = inputData.RelationSource
			file.Name = inputData.Name
			file.Mime = inputData.Mime
			file.Path = inputData.Path
			file.Description = inputData.Description
		}
		if databaseData, inDatabase = isFileInMap(file, databaseMap); inDatabase {
			file.ID = databaseData.ID
		}

		switch {
		case inInput && inDatabase:
			// update if data is found both in database and user input or have id from input
			updateFiles = append(updateFiles, file)
		case inInput && !inDatabase:
			// insert if data is found in user input but not in database
			insertFiles = append(insertFiles, file)
		case !inInput && inDatabase:
			// delete if data is found in database but not in user input
			deleteFiles = append(deleteFiles, file)
		}
	}

	if len(insertFiles) > 0 {
		if err := d.Insert(ctx, insertFiles); err != nil {
			return nil
		}
	}
	if len(updateFiles) > 0 {
		if err := d.Update(ctx, updateFiles); err != nil {
			return nil
		}
	}
	if len(deleteFiles) > 0 {
		if err := d.Delete(ctx, deleteFiles.GetIDs()); err != nil {
			return nil
		}
	}

	// Update the input data with the latest database records.
	updatesMap := append(append(insertFiles, updateFiles...), deleteFiles...).ExportByDataHash()
	for i, file := range files {
		if updatedFile, exists := updatesMap[file.MakeDataHash().String()]; exists {
			files[i] = updatedFile
		}
	}

	return nil
}

func getFileKey(file entity.FileEntity) string {
	if file.ID != "" {
		return file.ID.String()
	}
	return file.MakeDataHash().String()
}

func isFileInMap(file entity.FileEntity, fileMap entity.FileMapEntity) (entity.FileEntity, bool) {
	if file, foundByID := fileMap[file.ID.String()]; foundByID {
		return file, foundByID
	} else if file, foundByRelationKey := fileMap[file.GetRelationMapKey()]; foundByRelationKey {
		return file, foundByRelationKey
	}
	file, foundByHash := fileMap[file.MakeDataHash().String()]
	return file, foundByHash
}
