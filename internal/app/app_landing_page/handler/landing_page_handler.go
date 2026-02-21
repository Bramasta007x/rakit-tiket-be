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
	restricted.PUT("/landing-page/:id", h.updateLandingPage)

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

	// Jika request adalah Multipart (Upload File)
	if strings.Contains(contentType, "multipart/form-data") {
		return h.handleMultipartLandingPage(c)
	}

	// Fallback ke JSON biasa
	if c.Request().Method == http.MethodPost {
		return h.insertLandingPage(c)
	}

	return h.updateLandingPage(c)
}

func (h landingPageHandler) handleMultipartLandingPage(c echo.Context) error {
	ctx := c.Request().Context()

	// Ambil JSON String dari key "data"
	jsonData := c.FormValue("data")

	var landingPage entity.LandingPage

	if err := json.Unmarshal([]byte(jsonData), &landingPage); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format in 'data' field: "+err.Error())
	}

	// Generate ID di awal jika create baru, agar bisa direlasikan dengan File
	if landingPage.ID == "" {
		// Jika ada param ID di URL (PUT), pakai itu
		if ParamID := c.Param("id"); ParamID != "" {
			// Jika ada param ID di URL (PUT), pakai itu
			landingPage.ID = pubEntity.UUID(ParamID)
		} else {
			// Jika POST baru
			landingPage.ID = pubEntity.UUID(pubEntity.MakeUUID())
		}
	}

	// Helper function untuk proses upload
	processUpload := func(formKey string, relationSource string) (*string, error) {
		fileHeader, err := c.FormFile(formKey)
		if err != nil {
			if err == http.ErrMissingFile {
				return nil, nil // Tidak ada file diupload, skip
			}
			return nil, err
		}

		// Buka file
		src, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer src.Close()

		fileBytes, err := io.ReadAll(src)
		if err != nil {
			return nil, err
		}

		// Buat entity File
		fileEntity := &fileEntity.FileEntity{
			Name:        fileHeader.Filename,
			Description: fmt.Sprintf("Uploaded for Landing Page %s", relationSource),
			Data:        fileBytes,
			RelationEntity: pubEntity.MakeRelationEntity(
				landingPage.ID, // Merelasikan ke Landing Page ID
				relationSource, // Contoh: "landing_page_banner"
			),
		}

		// Simpan via FileService (Ini akan menyimpan ke disk & DB File)
		if err := h.fileService.Insert(ctx, fileEntity); err != nil {
			return nil, err
		}

		// Return ID file yang baru dibuat
		filePath := fmt.Sprintf("/%s/%s/%s.ref", fileEntity.RelationSource, fileEntity.RelationID.String(), fileEntity.ID.String())

		return &filePath, nil
	}

	// Proses Upload Banner Image (Jika ada)
	bannerFileID, err := processUpload("banner_image_file", "landing_page_banner")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed uploading banner: "+err.Error())
	}

	if bannerFileID != nil {
		landingPage.BannerImage = bannerFileID // Update struct dengan ID file baru
	}

	// Proses Upload Venue Image (jika ada)
	venueFileID, err := processUpload("venue_image_file", "landing_page_venue")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed uploading venue: "+err.Error())
	}

	if venueFileID != nil {
		landingPage.VenueImage = venueFileID // Update struct dengan ID file baru
	}

	// Proses Upload Logo Image (jika ada)
	logoFileID, err := processUpload("logo_image_file", "landing_page_logo")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed uploading logo: "+err.Error())
	}
	if logoFileID != nil {
		landingPage.LogoImage = logoFileID // Update struct dengan ID file baru
	}

	// Simpan Data Landing Page ke DB
	// Menggunakan Upsert, Insert jika tidak ada data, Update jika ada data
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
