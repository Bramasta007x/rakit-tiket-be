package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"rakit-tiket-be/internal/app/app_artist/service"
	fileService "rakit-tiket-be/internal/app/app_file/service"
	"rakit-tiket-be/internal/pkg/middleware"
	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_artist"
	fileEntity "rakit-tiket-be/pkg/entity/app_file"

	"github.com/labstack/echo/v4"
)

type ArtistHandler interface {
	RegisterRouter(g *echo.Group)
}

type artistHandler struct {
	artistService service.ArtistService
	fileService   fileService.FileService
	middleware    middleware.AuthMiddleware
}

func MakeArtistHandler(
	artistService service.ArtistService,
	fileService fileService.FileService,
	middleware middleware.AuthMiddleware,
) ArtistHandler {
	return artistHandler{
		artistService: artistService,
		fileService:   fileService,
		middleware:    middleware,
	}
}

func (h artistHandler) RegisterRouter(g *echo.Group) {
	restricted := g.Group("/v1/admin")
	restrictedPublic := g.Group("/v1")

	restrictedPublic.GET("/artists", h.searchArtists)
	restrictedPublic.GET("/artist/:id", h.getArtist)

	restricted.Use(h.middleware.VerifyToken)
	restricted.Use(h.middleware.RequireAdmin)

	restricted.POST("/artists", h.upsertArtistsWrapper)
	restricted.PUT("/artists", h.updateArtists)
	restricted.PUT("/artist/:id", h.upsertArtistByIDWrapper)
	restricted.DELETE("/artist/:id", h.softDeleteArtist)
}

func (h artistHandler) searchArtists(c echo.Context) error {
	var query entity.ArtistQuery

	if err := c.Bind(&query); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	artists, err := h.artistService.Search(
		c.Request().Context(),
		query,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, artists)
}

func (h artistHandler) getArtist(c echo.Context) error {
	id := pubEntity.UUID(c.Param("id"))

	artist, err := h.artistService.SearchByID(
		c.Request().Context(),
		id,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, artist)
}

func (h artistHandler) upsertArtistsWrapper(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		return h.handleMultipartArtists(c)
	}

	return h.insertArtists(c)
}

func (h artistHandler) upsertArtistByIDWrapper(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		return h.handleMultipartArtist(c)
	}

	return h.updateArtist(c)
}

func (h artistHandler) handleMultipartArtists(c echo.Context) error {
	ctx := c.Request().Context()

	jsonData := c.FormValue("data")
	var artists entity.Artists

	if err := json.Unmarshal([]byte(jsonData), &artists); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format in 'data' field: "+err.Error())
	}

	for i := range artists {
		if artists[i].ID == "" {
			artists[i].ID = pubEntity.UUID(pubEntity.MakeUUID())
		}

		imageFileID, err := h.processArtistImageUpload(c, artists[i].ID, "artist_image_file")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed uploading artist image: "+err.Error())
		}

		if imageFileID != nil {
			artists[i].Image = imageFileID
			artists[i].ImageUrl = imageFileID
		}
	}

	if err := h.artistService.Insert(ctx, artists); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, artists)
}

func (h artistHandler) handleMultipartArtist(c echo.Context) error {
	ctx := c.Request().Context()

	jsonData := c.FormValue("data")
	var artist entity.Artist

	if err := json.Unmarshal([]byte(jsonData), &artist); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format in 'data' field: "+err.Error())
	}

	artist.ID = pubEntity.UUID(c.Param("id"))

	imageFileID, err := h.processArtistImageUpload(c, artist.ID, "artist_image_file")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed uploading artist image: "+err.Error())
	}

	if imageFileID != nil {
		artist.Image = imageFileID
		artist.ImageUrl = imageFileID
	}

	if err := h.artistService.Update(ctx, entity.Artists{artist}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, artist)
}

func (h artistHandler) processArtistImageUpload(c echo.Context, artistID pubEntity.UUID, formKey string) (*string, error) {
	ctx := c.Request().Context()

	fileHeader, err := c.FormFile(formKey)
	if err != nil {
		if err == http.ErrMissingFile {
			return nil, nil
		}
		return nil, err
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	fileEntity := &fileEntity.FileEntity{
		Name:        fileHeader.Filename,
		Description: fmt.Sprintf("Uploaded for Artist %s", artistID),
		Data:        fileBytes,
		RelationEntity: pubEntity.MakeRelationEntity(
			artistID,
			"artist_image",
		),
	}

	if err := h.fileService.Insert(ctx, fileEntity); err != nil {
		return nil, err
	}

	filePathFolder := h.fileService.GetFilePath()
	filePath := fmt.Sprintf("%s/%s/%s/%s.ref", filePathFolder, fileEntity.RelationSource, fileEntity.RelationID.String(), fileEntity.ID.String())

	return &filePath, nil
}

func (h artistHandler) insertArtists(c echo.Context) error {
	var artists entity.Artists

	if err := c.Bind(&artists); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.artistService.Insert(
		c.Request().Context(),
		artists,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, artists)
}

func (h artistHandler) updateArtists(c echo.Context) error {
	var artists entity.Artists

	if err := c.Bind(&artists); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.artistService.Update(
		c.Request().Context(),
		artists,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, artists)
}

func (h artistHandler) updateArtist(c echo.Context) error {
	var artist entity.Artist

	if err := c.Bind(&artist); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	artist.ID = pubEntity.UUID(c.Param("id"))

	if err := h.artistService.Update(
		c.Request().Context(),
		entity.Artists{artist},
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, artist)
}

func (h artistHandler) softDeleteArtist(c echo.Context) error {
	id := pubEntity.UUID(c.Param("id"))

	if err := h.artistService.SoftDelete(
		c.Request().Context(),
		id,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(
		http.StatusOK,
		map[string]string{
			"message": "Artist deleted successfully",
		},
	)
}
