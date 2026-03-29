package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	fileService "rakit-tiket-be/internal/app/app_file/service"
	"rakit-tiket-be/internal/app/app_landing_page/service"
	"rakit-tiket-be/internal/pkg/middleware"
	pubEntity "rakit-tiket-be/pkg/entity"
	fileEntity "rakit-tiket-be/pkg/entity/app_file"
	entity "rakit-tiket-be/pkg/entity/app_landing_page"

	"github.com/labstack/echo/v4"
)

type LandingPageHandler interface {
	RegisterRouter(g *echo.Group)
}

type landingPageHandler struct {
	landingPageService service.LandingPageService
	fileService        fileService.FileService
	middleware         middleware.AuthMiddleware
}

func MakeLandingPageHandler(landingPageService service.LandingPageService, fileService fileService.FileService, middleware middleware.AuthMiddleware) landingPageHandler {
	return landingPageHandler{
		landingPageService: landingPageService,
		fileService:        fileService,
		middleware:         middleware,
	}
}

func (h landingPageHandler) RegisterRouter(g *echo.Group) {
	restricted := g.Group("/v1/admin")
	restrictedPublic := g.Group("/v1")

	restrictedPublic.GET("/landing-pages", h.searchLandingPages)

	restricted.Use(h.middleware.VerifyToken)
	restricted.Use(h.middleware.RequireAdmin)

	restricted.POST("/landing-pages", h.insertLandingPages)
	restricted.POST("/landing-page", h.upsertLandingPageWrapper)

	restricted.PUT("/landing-pages", h.updateLandingPages)
	restricted.PUT("/landing-page/:id", h.upsertLandingPageWrapper)

	restricted.DELETE("/landing-page/:id", h.softDeleteLandingPage)
}

func (h landingPageHandler) searchLandingPages(c echo.Context) error {
	var query entity.LandingPageQuery

	if err := c.Bind(&query); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	pages, err := h.landingPageService.Search(c.Request().Context(), query)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, pages)
}

func (h landingPageHandler) insertLandingPages(c echo.Context) error {
	var pages entity.LandingPages

	if err := c.Bind(&pages); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.landingPageService.Insert(c.Request().Context(), pages); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, pages)
}

func (h landingPageHandler) upsertLandingPageWrapper(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		return h.handleMultipartLandingPage(c)
	}

	if c.Request().Method == http.MethodPost {
		return h.insertLandingPage(c)
	}

	return h.updateLandingPage(c)
}

func (h landingPageHandler) handleMultipartLandingPage(c echo.Context) error {
	ctx := c.Request().Context()

	jsonData := c.FormValue("data")

	var landingPage entity.LandingPage

	if err := json.Unmarshal([]byte(jsonData), &landingPage); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format in 'data' field: "+err.Error())
	}

	if landingPage.ID == "" {
		if ParamID := c.Param("id"); ParamID != "" {
			landingPage.ID = pubEntity.UUID(ParamID)
		} else {
			landingPage.ID = pubEntity.UUID(pubEntity.MakeUUID())
		}
	}

	processUpload := func(formKey string, relationSource string) (*string, error) {
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
			Description: fmt.Sprintf("Uploaded for Landing Page %s", relationSource),
			Data:        fileBytes,
			RelationEntity: pubEntity.MakeRelationEntity(
				landingPage.ID,
				relationSource,
			),
		}

		if err := h.fileService.Insert(ctx, fileEntity); err != nil {
			return nil, err
		}

		filePathFolder := h.fileService.GetFilePath()
		filePath := fmt.Sprintf("%s/%s/%s/%s.ref", filePathFolder, fileEntity.RelationSource, fileEntity.RelationID.String(), fileEntity.ID.String())

		return &filePath, nil
	}

	// Upload banner_image
	bannerFileID, err := processUpload("banner_image", "landing_page_banner")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed uploading banner_image: "+err.Error())
	}
	if bannerFileID != nil {
		landingPage.BannerImage = bannerFileID
	}

	// Upload logo_image
	logoFileID, err := processUpload("logo_image", "landing_page_logo")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed uploading logo_image: "+err.Error())
	}
	if logoFileID != nil {
		landingPage.LogoImage = logoFileID
	}

	// Upload venue_image
	venueImageFileID, err := processUpload("venue_image", "landing_page_venue")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed uploading venue_image: "+err.Error())
	}
	if venueImageFileID != nil {
		landingPage.VenueImage = venueImageFileID
	}

	// Upload venue_layout
	venueLayoutFileID, err := processUpload("venue_layout", "landing_page_venue_layout")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed uploading venue_layout: "+err.Error())
	}
	if venueLayoutFileID != nil {
		landingPage.VenueLayout = venueLayoutFileID
	}

	// Upload artist images (multiple with index suffix: artist_image_0, artist_image_1, etc.)
	for i := range landingPage.Artists {
		formKey := fmt.Sprintf("artist_image_%d", i)
		artistImageFileID, err := processUpload(formKey, "artist_image")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed uploading artist_image_%d: %v", i, err.Error()))
		}
		if artistImageFileID != nil {
			landingPage.Artists[i].Image = artistImageFileID
			landingPage.Artists[i].ImageUrl = artistImageFileID
		}
	}

	// Also check for generic artist_image (single upload without index)
	artistImageFileID, err := processUpload("artist_image", "artist_image")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed uploading artist_image: "+err.Error())
	}
	if artistImageFileID != nil && len(landingPage.Artists) > 0 {
		// Apply to first artist if no indexed image was uploaded
		landingPage.Artists[0].Image = artistImageFileID
		landingPage.Artists[0].ImageUrl = artistImageFileID
	}

	var serviceError error
	if c.Request().Method == http.MethodPost {
		serviceError = h.landingPageService.Insert(ctx, entity.LandingPages{landingPage})
	} else {
		serviceError = h.landingPageService.Update(ctx, entity.LandingPages{landingPage})
	}

	if serviceError != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, serviceError.Error())
	}

	return c.JSON(http.StatusOK, landingPage)
}

func (h landingPageHandler) insertLandingPage(c echo.Context) error {
	var page entity.LandingPage

	if err := c.Bind(&page); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.landingPageService.Insert(c.Request().Context(), entity.LandingPages{page}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, page)
}

func (h landingPageHandler) updateLandingPages(c echo.Context) error {
	var pages entity.LandingPages

	if err := c.Bind(&pages); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.landingPageService.Update(c.Request().Context(), pages); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, pages)
}

func (h landingPageHandler) updateLandingPage(c echo.Context) error {
	var page entity.LandingPage

	if err := c.Bind(&page); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.landingPageService.Update(c.Request().Context(), entity.LandingPages{page}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, page)
}

func (h landingPageHandler) softDeleteLandingPage(c echo.Context) error {
	id := pubEntity.UUID(c.Param("id"))

	if err := h.landingPageService.SoftDelete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(
		http.StatusOK,
		map[string]string{"message": "Landing page deleted successfully"},
	)
}
