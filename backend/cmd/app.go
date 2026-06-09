package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"backend/api/middleware"
	"backend/api/route"
	"backend/bootstrap"
	"backend/controller"
	"backend/internal/auth"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const (
	defaultPort   = "8082"
	defaultDBHost = "localhost"
	defaultDBPort = 5433
)

func run(log *slog.Logger) error {
	// Load godotenv first
	if err := godotenv.Load(); err != nil {
		slog.Warn("Warning: failed to load .env file, reading system environment variables")
	}

	// Create listener context for operating system shutdown signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Initialize database
	db, err := bootstrap.NewDB(ctx)
	if err != nil {
		log.Error("Database connection failed", "error", err.Error())
		return err
	}
	defer db.Close()
	log.Info("Database connection pool initialized successfully")

	// 2. Initialize RSA keys for tokens
	keysDir := utils.GetEnvString("DEMO_KEYS_DIR", filepath.Join("tmp", "demo-keys"))
	if err := auth.EnsureKeysExist(keysDir); err != nil {
		log.Warn("Warning: Failed to ensure token key files", "error", err.Error())
	} else {
		log.Info("RSA keys initialized in", "keysDir", keysDir)
		// Generate a test token for debugging
		testToken, err := auth.GenerateTokenWithPrivateKey("1", "admin@example.com", "admin")
		if err == nil {
			log.Info("\n--- TEST TOKEN ---", "token", testToken, "--- END TOKEN ---\n\n")
		}
	}

	// 3. Setup Gin Engine
	r := gin.Default()

	// Đăng ký CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Phục vụ static files cho local storage uploads
	r.Static("/static", "./static")

	// Request logging middleware
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("Request",
			"method", c.Request.Method,
			"url", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start),
		)
	})

	// 4. Initialize Dependency Container & Router wiring
	container := bootstrap.NewContainer(db.GetPool())
	authMiddleware := middleware.NewAuthMiddleware()
	locationCtl := controller.NewLocationController()

	v1 := r.Group("/api/v1")
	route.SetupUserRoutes(v1, container.UserCtl, authMiddleware)
	route.SetupAddressRoutes(v1, container.AddressCtl, authMiddleware)
	route.SetupCatalogRoutes(v1, container.CatalogCtl, authMiddleware)
	route.SetupInventoryRoutes(v1, container.InventoryCtl, authMiddleware)
	route.SetupCartRoutes(v1, container.CartCtl, authMiddleware)
	route.SetupPromotionVoucherRoutes(v1, container.PromotionVoucherCtl, authMiddleware)
	route.SetupOrderRoutes(v1, container.OrderCtl, authMiddleware)
	route.SetupBannerRoutes(v1, container.BannerCtl, authMiddleware)
	route.SetupUploadRoutes(v1, container.UploadCtl, authMiddleware)
	route.SetupLocationRoutes(v1, locationCtl)
	route.SetupWishlistRoutes(v1, container.WishlistCtl, authMiddleware)
	route.SetupAnalyticsRoutes(v1, container.AnalyticsCtl, authMiddleware)
	route.SetupPolicyRoutes(v1, container.PolicyCtl, authMiddleware)

	// 5. Start Server
	serverPort := utils.GetEnvString("BACKEND_PORT", defaultPort)
	log.Info("Starting HTTP server on port", "port", serverPort)

	// Convert port string to int to pass to Run
	portInt, err := strconv.Atoi(serverPort)
	if err != nil {
		portInt = 8082
	}

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(portInt),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("Listening on port", "port", portInt)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		log.Error("Server error", "err", err.Error())
		return err
	case <-ctx.Done():
		log.Info("Shutting down gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("Graceful shutdown failed", "error", err.Error())
			return err
		}
		log.Info("Server stopped")
		return nil
	}
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}
