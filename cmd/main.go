package main

import (
	"context"
	"fmt"
	"os"

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

	// PostgreSQL
	pgClient := client.MakePostgreSQLClientFromEnv()

	if err := pgClient.Migration(); err != nil {
		log.Error(context.Background(), "Migration failed")
		fmt.Fprintf(os.Stderr, "Migration error: %v\n", err)
		os.Exit(1)
	}
	log.Info(context.Background(), "Migration success")

	// --- HTTP Server ---
	e := echo.New()

	// =====================
	// Service & Handler
	// =====================
	sqlDB := pgClient.GetSQLDB()

	landingPageSvc := landingPageService.MakeLandingPageService(sqlDB)
	httpHandler := landingPageHandler.MakeHttpAdapter(landingPageSvc)

	// =====================
	// Register Routes
	// =====================
	apiGroup := e.Group("/api")
	httpHandler.RegisterRoute(apiGroup)

	// =====================
	// Start Server
	// =====================
	port := envgo.GetString("PORT", "8000")
	log.Info(context.Background(), "Starting HTTP server on port "+port)
	e.Logger.Fatal(e.Start(":" + port))
}
