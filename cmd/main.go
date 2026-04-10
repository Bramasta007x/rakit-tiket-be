package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	authHandler "rakit-tiket-be/internal/app/app_auth/handler"
	authService "rakit-tiket-be/internal/app/app_auth/service"

	fileHandler "rakit-tiket-be/internal/app/app_file/handler"
	fileService "rakit-tiket-be/internal/app/app_file/service"

	landingPageHandler "rakit-tiket-be/internal/app/app_landing_page/handler"
	landingPageService "rakit-tiket-be/internal/app/app_landing_page/service"

	ticketHandler "rakit-tiket-be/internal/app/app_ticket/handler"
	ticketService "rakit-tiket-be/internal/app/app_ticket/service"

	regHandler "rakit-tiket-be/internal/app/app_registrant/handler"
	regService "rakit-tiket-be/internal/app/app_registrant/service"

	orderHandler "rakit-tiket-be/internal/app/app_order/handler"
	orderService "rakit-tiket-be/internal/app/app_order/service"

	artistHandler "rakit-tiket-be/internal/app/app_artist/handler"
	artistService "rakit-tiket-be/internal/app/app_artist/service"

	eventHandler "rakit-tiket-be/internal/app/app_event/handler"
	eventService "rakit-tiket-be/internal/app/app_event/service"

	paymentHandler "rakit-tiket-be/internal/app/app_payment/handler"
	paymentService "rakit-tiket-be/internal/app/app_payment/service"

	"rakit-tiket-be/config"
	"rakit-tiket-be/internal/pkg/client"
	"rakit-tiket-be/internal/pkg/email"
	"rakit-tiket-be/internal/pkg/middleware"
	"rakit-tiket-be/internal/pkg/payment"
	"rakit-tiket-be/pkg/constant"
	"rakit-tiket-be/pkg/util"

	"github.com/labstack/echo/v4"
	"github.com/rs/cors"
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

	clientOrigin := envgo.GetString("CLIENT_ORIGIN_URL", "*")
	corsOptions := config.CorsOptions(clientOrigin)
	e.Use(echo.WrapMiddleware(cors.New(corsOptions).Handler))

	// Middleware
	authMiddleware := middleware.MakeAuthMiddleware(log)

	smtpHost := envgo.GetString("SMTP_HOST", "")
	smtpPort, err := strconv.Atoi(envgo.GetString("SMTP_PORT", ""))
	if err != nil || smtpHost == "" || smtpPort == 0 {
		log.Error(context.Background(), "Missing required SMTP configuration: SMTP_HOST and SMTP_PORT must be set")
		os.Exit(1)
	}
	smtpUser := envgo.GetString("SMTP_USER", "")
	smtpPass := envgo.GetString("SMTP_PASS", "")
	senderName := envgo.GetString("SENDER_NAME", "")
	senderEmail := envgo.GetString("SENDER_EMAIL", "")
	if smtpUser == "" || smtpPass == "" || senderName == "" || senderEmail == "" {
		log.Error(context.Background(), "Missing required SMTP configuration: SMTP_USER, SMTP_PASS, SENDER_NAME, and SENDER_EMAIL must be set")
		os.Exit(1)
	}

	// Midtrans / Payment Factory
	midtransServerKey := envgo.GetString("MIDTRANS_SERVER_KEY", "")
	midtransEnvironment := envgo.GetString("MIDTRANS_ENVIRONMENT", "false")
	if midtransServerKey == "" {
		log.Error(context.Background(), "Missing required configuration: MIDTRANS_SERVER_KEY must be set")
		os.Exit(1)
	}
	midtransIsProduction := midtransEnvironment == "true"
	paymentFactory := payment.NewPaymentFactory(midtransServerKey, midtransIsProduction)
	emailSvc := email.MakeEmailService(log, smtpHost, smtpPort, smtpUser, smtpPass, senderName, senderEmail)

	// Service
	landingPageService := landingPageService.MakeLandingPageService(sqlDB)
	fileService := fileService.MakeFileService(log, sqlDB)
	authSvc := authService.MakeAuthService(log, sqlDB)

	ticketSvc := ticketService.MakeTicketService(log, sqlDB)
	regService := regService.MakeRegistrantService(log, sqlDB, paymentFactory)
	ordService := orderService.MakeOrderService(log, sqlDB, paymentFactory, emailSvc)
	eventSvc := eventService.MakeEventService(log, sqlDB)
	artistSvc := artistService.MakeArtistService(log, sqlDB)

	bankAccountSvc := paymentService.MakeBankAccountService(log, sqlDB)
	manualTransferSvc := paymentService.MakeManualTransferService(log, sqlDB, emailSvc)

	// Adapter
	landingPageAdapter := landingPageHandler.MakeHttpAdapter(landingPageService, fileService, authMiddleware)
	fileAdapter := fileHandler.MakeFileAdapter(log, fileService)
	authAdapter := authHandler.MakeHttpAdapter(log, authSvc)
	ticketAdapter := ticketHandler.MakeHttpAdapter(ticketSvc, authMiddleware)
	registrantHttpHandler := regHandler.MakeHttpAdapter(regService, authMiddleware)
	orderHttpHandler := orderHandler.MakeHttpAdapter(log, ordService)
	eventAdapter := eventHandler.MakeHttpAdapter(eventSvc, authMiddleware)
	artistAdapter := artistHandler.MakeHttpAdapter(artistSvc, fileService, authMiddleware)
	paymentAdapter := paymentHandler.MakeHttpAdapter(log, bankAccountSvc, manualTransferSvc, fileService, authMiddleware)

	// Register Routes
	apiGroup := e.Group("/api")

	landingPageAdapter.RegisterRoute(apiGroup)
	fileAdapter.RegisterRouter(apiGroup)
	authAdapter.RegisterRoute(apiGroup)
	ticketAdapter.RegisterRoute(apiGroup)
	registrantHttpHandler.RegisterRoute(apiGroup)
	orderHttpHandler.RegisterRoute(apiGroup)
	eventAdapter.RegisterRoute(apiGroup)
	artistAdapter.RegisterRoute(apiGroup)
	paymentAdapter.RegisterRoute(apiGroup)

	// Start Cron Scheduler
	// scheduler := cron.NewScheduler(ordService, log)
	// if err := scheduler.Start(); err != nil {
	// 	log.Error(context.Background(), "Failed to start cron scheduler")
	// 	os.Exit(1)
	// }

	// // Graceful Shutdown
	// go func() {
	// 	quit := make(chan os.Signal, 1)
	// 	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 	<-quit
	// 	log.Info(context.Background(), "Shutting down server...")
	// 	scheduler.Stop()
	// 	os.Exit(0)
	// }()

	// Start Server
	port := envgo.GetString("PORT", "8001")
	log.Info(context.Background(), "Starting HTTP server on port "+port)
	e.Logger.Fatal(e.Start(":" + port))
}
