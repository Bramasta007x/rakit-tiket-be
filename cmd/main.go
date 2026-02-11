package main

import (
	"context"
	"fmt"
	"os"

	authHandler "rakit-tiket-be/internal/app/app_auth/handler"
	authService "rakit-tiket-be/internal/app/app_auth/service"

	fileHandler "rakit-tiket-be/internal/app/app_file/handler"
	fileService "rakit-tiket-be/internal/app/app_file/service"

	landingPageHandler "rakit-tiket-be/internal/app/app_landing_page/handler"
	landingPageService "rakit-tiket-be/internal/app/app_landing_page/service"

	ticketHandler "rakit-tiket-be/internal/app/app_ticket/handler"
	ticketService "rakit-tiket-be/internal/app/app_ticket/service"

	"rakit-tiket-be/internal/pkg/client"
	"rakit-tiket-be/internal/pkg/middleware"
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

	sqlDB := pgClient.GetSQLDB()

	// HTTP Server
	e := echo.New()

	// Middleware
	authMiddleware := middleware.MakeAuthMiddleware(log)

	// Services
	landingPageService := landingPageService.MakeLandingPageService(sqlDB)
	fileService := fileService.MakeFileService(log, sqlDB)
	authSvc := authService.MakeAuthService(log, sqlDB)
	ticketSvc := ticketService.MakeTicketService(sqlDB)

	// Handlers / Adapters
	landingPageAdapter := landingPageHandler.MakeHttpAdapter(landingPageService, fileService, authMiddleware)
	fileAdapter := fileHandler.MakeFileAdapter(log, fileService)
	authAdapter := authHandler.MakeHttpAdapter(log, authSvc)
	ticketAdapter := ticketHandler.MakeHttpAdapter(ticketSvc, authMiddleware)

	// Register Routes
	apiGroup := e.Group("/api")

	landingPageAdapter.RegisterRoute(apiGroup)
	fileAdapter.RegisterRouter(apiGroup)
	authAdapter.RegisterRoute(apiGroup)
	ticketAdapter.RegisterRoute(apiGroup)

	// Start Server
	port := envgo.GetString("PORT", "8000")
	log.Info(context.Background(), "Starting HTTP server on port "+port)
	e.Logger.Fatal(e.Start(":" + port))
}
