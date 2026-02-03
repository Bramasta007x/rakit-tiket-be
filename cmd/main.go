package main

import (
	"context"
	"fmt"
	"os"

	fileHandler "rakit-tiket-be/internal/app/app_file/handler"
	fileService "rakit-tiket-be/internal/app/app_file/service"

	landingPageHandler "rakit-tiket-be/internal/app/app_landing_page/handler"
	landingPageService "rakit-tiket-be/internal/app/app_landing_page/service"

	"rakit-tiket-be/internal/pkg/client"
	"rakit-tiket-be/pkg/constant"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
	"gitlab.com/threetopia/envgo"
)

func main() {
	envgo.LoadDotEnv("./../.env")

	log := util.MakeLogUtil(
		envgo.GetString(constant.LogEnvironment, constant.LogDevelopment),
		nil,
	)

	// =====================
	// PostgreSQL
	// =====================
	pgClient := client.MakePostgreSQLClientFromEnv()

	if err := pgClient.Migration(); err != nil {
		log.Error(context.Background(), "Migration failed")
		fmt.Fprintf(os.Stderr, "Migration error: %v\n", err)
		os.Exit(1)
	}
	log.Info(context.Background(), "Migration success")

	sqlDB := pgClient.GetSQLDB()

	// =====================
	// HTTP Server
	// =====================
	e := echo.New()

	// =====================
	// Services
	// =====================
	landingPageSvc := landingPageService.MakeLandingPageService(sqlDB)
	fileSvc := fileService.MakeFileService(log, sqlDB)

	// =====================
	// Handlers / Adapters
	// =====================
	landingPageAdapter := landingPageHandler.MakeHttpAdapter(landingPageSvc)
	fileAdapter := fileHandler.MakeFileAdapter(log, fileSvc)

	// =====================
	// Register Routes
	// =====================
	apiGroup := e.Group("/api")

	landingPageAdapter.RegisterRoute(apiGroup)
	fileAdapter.RegisterRouter(apiGroup)

	// =====================
	// Start Server
	// =====================
	port := envgo.GetString("PORT", "8000")
	log.Info(context.Background(), "Starting HTTP server on port "+port)
	e.Logger.Fatal(e.Start(":" + port))
}
