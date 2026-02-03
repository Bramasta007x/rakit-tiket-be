package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"rakit-tiket-be/internal/app/app_file/service"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_file"

	"github.com/labstack/echo/v4"

	model "rakit-tiket-be/pkg/model/app_file"
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/envgo"
	"go.uber.org/zap"
)

type FileAdapter interface {
	RegisterRouter(g *echo.Group)
}

type fileAdapter struct {
	log         util.LogUtil
	fileService service.FileService
}

func MakeFileAdapter(log util.LogUtil, fileService service.FileService) FileAdapter {
	return fileAdapter{
		log:         log,
		fileService: fileService,
	}
}

func (a fileAdapter) RegisterRouter(g *echo.Group) {
	restricted := g.Group("/v1/admin")

	restricted.GET("/files", a.searchFile)
	restricted.GET("/file/:fileID", a.getFileByID)

	restricted.POST("/file", a.insertFile)
	restricted.POST("/files", a.insertFiles)

	restricted.PUT("/file/:fileID", a.updateFileByID)
	restricted.PUT("/files", a.updateFiles)

	restricted.DELETE("/file/:fileID", a.deleteFileByID)

	restricted.PATCH("/files", a.upsertFiles)
}

func (a fileAdapter) searchFile(c echo.Context) error {
	var request model.SearchFilesRequestModel

	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	files, count, err := a.fileService.Search(c.Request().Context(), request.FileQuery)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	enctype := c.Request().Header.Get("Accept-Encoding")
	if enctype == "application/json" {
		return c.JSON(http.StatusOK, files)
	}

	return c.JSON(model.MakeSearchFilesResponseModel(http.StatusOK, count, files))
}

func (a fileAdapter) getFileByID(c echo.Context) error {
	file, err := a.fileService.GetByID(c.Request().Context(), pubEntity.UUID(c.Param("fileID")))

	if err != nil {
		return err
	}

	enctype := c.Request().Header.Get("Accept-Encoding")
	if enctype == "application/json" {
		return c.JSON(http.StatusOK, file)
	}

	c.Response().Header().Add("Content-Disposition", fmt.Sprintf("filename=%s", file.Name))

	return c.Blob(http.StatusOK, file.Mime, file.Data)
}

func (a fileAdapter) insertFile(c echo.Context) error {
	enctype := c.Request().Header.Get("Content-Type")

	if strings.Contains(enctype, "multipart/form-data") {
		return a.InsertFileMultipartForm(c)
	} else if strings.Contains(enctype, "application/json") {
		return a.insertFileJSON(c)
	}

	a.log.Error(c.Request().Context(), "fileAdapter.insertFile", zap.Error(fmt.Errorf("wrong content type given")))
	return fmt.Errorf("wrong content type given")
}

func (a fileAdapter) InsertFileMultipartForm(c echo.Context) error {
	filePath := envgo.GetString("FILE_PATH", "../../../assets/app_file")
	//source
	file, err := c.FormFile("file")
	if err != nil {
		a.log.Error(c.Request().Context(), "fileAdapter.InsertFileMultipartForm.FormFile", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	src, err := file.Open()
	if err != nil {
		a.log.Error(c.Request().Context(), "fileAdapter.InsertFileMultipartForm.file.Open", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	defer src.Close()

	//read file content
	fileData, err := io.ReadAll(src)
	if err != nil {
		a.log.Error(c.Request().Context(), "fileAdapter.InsertFileMultipartForm.io.ReadAll", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	fileName := c.FormValue("name")
	fileDescription := c.FormValue("description")

	fileEntity := entity.FileEntity{
		Name:        fileName,
		Description: fileDescription,
		Path:        filePath,
		Data:        fileData,
		Mime:        http.DetectContentType(fileData),
		RelationEntity: pubEntity.RelationEntity{
			RelationID:     pubEntity.MakeUUID(c.FormValue("relationId")),
			RelationSource: c.FormValue("relationSource"),
		},
	}

	// if err := c.Validate(fileEntity); err != nil {
	// 	a.log.Error(c.Request().Context(), "fileAdapter.InsertFileMultipartForm.Validate", zap.Error(err))
	// 	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	// }

	if err := a.fileService.Insert(c.Request().Context(), &fileEntity); err != nil {
		a.log.Error(c.Request().Context(), "fileAdapter.InsertFileMultipartForm.Insert", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, fileEntity)
}

func (a fileAdapter) insertFileJSON(c echo.Context) error {
	var request entity.FileEntity

	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	} else if err := c.Validate(request); err != nil {
		return err
	}

	if err := a.fileService.Insert(c.Request().Context(), &request); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, request)
}

func (a fileAdapter) insertFiles(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		return a.insertFilesMultipartForm(c)
	} else if strings.Contains(contentType, "application/json") {
		return a.insertFilesJSON(c)
	}

	return fmt.Errorf("wrong content type given")
}

func (a fileAdapter) insertFilesMultipartForm(c echo.Context) error {
	ctx := c.Request().Context()

	formParams, _ := c.FormParams()
	formInputs := util.GetFormRepeaterInput(formParams, "file", "name", "file", "description", "relationId", "relationSource")

	name := formInputs.GetByName("name")
	description := formInputs.GetByName("description")
	relationID := formInputs.GetByName("relationId")
	relationSource := formInputs.GetByName("relationSource")

	var files entity.FilesEntity
	for i := 0; i < formInputs.Len(); i++ {
		// Source
		fileData, err := c.FormFile(fmt.Sprintf("file[%d][file]", i))
		if err != nil {
			a.log.Error(ctx, "fileAdapter.insertFilesMultipartForm", zap.Error(err))
			return err
		}
		src, err := fileData.Open()
		if err != nil {
			a.log.Error(ctx, "fileAdapter.insertFilesMultipartForm", zap.Error(err))
			return err
		}
		defer src.Close()

		// Read file content
		data, err := io.ReadAll(src)
		if err != nil {
			a.log.Error(ctx, "fileAdapter.insertFilesMultipartForm", zap.Error(err))
			return err
		}
		file := entity.FileEntity{
			Name:        name.GetByIdx(i),
			Description: description.GetByIdx(i),
			Data:        data,
			RelationEntity: pubEntity.MakeRelationEntity(
				pubEntity.UUID(relationID.GetByIdx(i)),
				relationSource.GetByIdx(i),
			),
		}
		if err := c.Validate(file); err != nil {
			return err
		}

		files = append(files, file)

	}

	if err := a.fileService.Inserts(c.Request().Context(), files); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, files)
}

func (a fileAdapter) insertFilesJSON(c echo.Context) error {
	var request entity.FilesEntity

	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	} else if len(request) > 0 {
		for _, file := range request {
			if err := c.Validate(file); err != nil {
				return err
			}
		}
	}

	if err := a.fileService.Inserts(c.Request().Context(), request); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, request)
}

func (a fileAdapter) updateFileByID(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		return a.updateFileMultipartForm(c)
	} else if strings.Contains(contentType, "application/json") {
		return a.updateFileJSON(c)
	}

	return fmt.Errorf("wrong content type given")
}

func (a fileAdapter) updateFileMultipartForm(c echo.Context) error {
	// Source
	file, err := c.FormFile("file")
	if err != nil {
		a.log.Error(c.Request().Context(), "error when get form file", zap.Error(err))
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Read file content
	data, err := io.ReadAll(src)
	if err != nil {
		log.Println(err)
		return err
	}
	fileName := c.FormValue("name")
	fileEntity := entity.FileEntity{
		ID:             pubEntity.UUID(c.Param("fileID")),
		Name:           fileName,
		Data:           entity.FileData(data),
		RelationEntity: pubEntity.MakeRelationEntity(pubEntity.MakeUUID(c.FormValue("relationId")), c.FormValue("relationSource")),
	}

	if err := a.fileService.Update(c.Request().Context(), &fileEntity); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, fileEntity)
}

func (a fileAdapter) updateFileJSON(c echo.Context) error {
	var request entity.FileEntity

	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	request.ID = pubEntity.UUID(c.Param("documentID"))

	if err := a.fileService.Update(c.Request().Context(), &request); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, request)
}

func (a fileAdapter) updateFiles(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		return a.updateFilesMultipartForm(c)
	} else if strings.Contains(contentType, "application/json") {
		return a.updateFilesJSON(c)
	}

	return fmt.Errorf("wrong content type given")
}

func (a fileAdapter) updateFilesMultipartForm(c echo.Context) error {
	ctx := c.Request().Context()

	formParams, _ := c.FormParams()
	formInputs := util.GetFormRepeaterInput(formParams, "file", "name", "file", "description", "relationId", "relationSource")

	name := formInputs.GetByName("name")
	description := formInputs.GetByName("description")
	relationID := formInputs.GetByName("relationId")
	relationSource := formInputs.GetByName("relationSource")

	var files entity.FilesEntity
	for i := 0; i < formInputs.Len(); i++ {
		// Source
		fileData, err := c.FormFile(fmt.Sprintf("file[%d][file]", i))
		if err != nil {
			a.log.Error(ctx, "fileAdapter.updateFilesMultipartForm", zap.Error(err))
			return err
		}

		src, err := fileData.Open()
		if err != nil {
			a.log.Error(ctx, "fileAdapter.updateFilesMultipartForm", zap.Error(err))
			return err
		}
		defer src.Close()

		// Read file content
		data, err := io.ReadAll(src)
		if err != nil {
			a.log.Error(ctx, "fileAdapter.updateFilesMultipartForm", zap.Error(err))
			return err
		}

		file := entity.FileEntity{
			Name:        name.GetByIdx(i),
			Description: description.GetByIdx(i),
			Data:        data,
			RelationEntity: pubEntity.MakeRelationEntity(
				pubEntity.UUID(relationID.GetByIdx(i)),
				relationSource.GetByIdx(i),
			),
		}

		if err := c.Validate(file); err != nil {
			return err
		}

		files = append(files, file)

	}

	if err := a.fileService.Updates(c.Request().Context(), files); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, files)
}

func (a fileAdapter) updateFilesJSON(c echo.Context) error {
	var request entity.FilesEntity

	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	} else if len(request) > 0 {
		for _, document := range request {
			if err := c.Validate(document); err != nil {
				return err
			}
		}
	}

	if err := a.fileService.Updates(c.Request().Context(), request); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, request)
}

func (a fileAdapter) deleteFileByID(c echo.Context) error {
	err := a.fileService.DeleteByID(c.Request().Context(), pubEntity.UUID(c.Param("fileID")))

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Landing page deleted successfully"})
}

func (a fileAdapter) upsertFiles(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		return a.upsertFilesMultipartForm(c)
	} else if strings.Contains(contentType, "application/json") {
		return a.upsertFilesJSON(c)
	}

	return fmt.Errorf("wrong content type given")
}

func (a fileAdapter) upsertFilesMultipartForm(c echo.Context) error {
	ctx := c.Request().Context()

	formParams, _ := c.FormParams()
	formInputs := util.GetFormRepeaterInput(formParams, "file", "name", "file", "description", "relationId", "relationSource")

	name := formInputs.GetByName("name")
	description := formInputs.GetByName("description")
	relationID := formInputs.GetByName("relationId")
	relationSource := formInputs.GetByName("relationSource")

	var files entity.FilesEntity
	for i := 0; i < formInputs.Len(); i++ {
		// Source
		fileData, err := c.FormFile(fmt.Sprintf("file[%d][file]", i))
		if err != nil {
			a.log.Error(ctx, "fileAdapter.upsertFilesMultipartForm", zap.Error(err))
			return err
		}
		src, err := fileData.Open()
		if err != nil {
			a.log.Error(ctx, "fileAdapter.upsertFilesMultipartForm", zap.Error(err))
			return err
		}
		defer src.Close()

		// Read file content
		data, err := io.ReadAll(src)
		if err != nil {
			a.log.Error(ctx, "fileAdapter.upsertFilesMultipartForm", zap.Error(err))
			return err
		}

		file := entity.FileEntity{
			Name:        name.GetByIdx(i),
			Description: description.GetByIdx(i),
			Data:        data,
			RelationEntity: pubEntity.MakeRelationEntity(
				pubEntity.UUID(relationID.GetByIdx(i)),
				relationSource.GetByIdx(i),
			),
		}

		if err := c.Validate(file); err != nil {
			return err
		}

		files = append(files, file)

	}

	if err := a.fileService.Upsert(c.Request().Context(), files); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, files)
}

func (a fileAdapter) upsertFilesJSON(c echo.Context) error {
	var request entity.FilesEntity

	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	} else if len(request) > 0 {
		for _, file := range request {
			if err := c.Validate(file); err != nil {
				return err
			}
		}
	}

	if err := a.fileService.Upsert(c.Request().Context(), request); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, request)

}
