package main

import (
	"context"
	"log"
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

func main() {
	// Load godotenv first
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: failed to load .env file, reading system environment variables")
	}

	// Create listener context for operating system shutdown signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Initialize database
	db, err := bootstrap.NewDB(ctx)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("Database connection pool initialized successfully")

	// 2. Initialize RSA keys for tokens
	keysDir := utils.GetEnvString("DEMO_KEYS_DIR", filepath.Join("tmp", "demo-keys"))
	if err := auth.EnsureKeysExist(keysDir); err != nil {
		log.Printf("Warning: Failed to ensure token key files: %v", err)
	} else {
		log.Printf("RSA keys initialized in %s", keysDir)
		// Generate a test token for debugging
		testToken, err := auth.GenerateTokenWithPrivateKey("1", "admin@example.com", "admin")
		if err == nil {
			log.Printf("\n--- TEST TOKEN ---\n%s\n--- END TOKEN ---\n\n", testToken)
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
		log.Printf("[%s] %s -> Status: %d, Duration: %s",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start),
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

	// 5. Start Server
	serverPort := utils.GetEnvString("BACKEND_PORT", defaultPort)
	log.Printf("Starting HTTP server on port %s...", serverPort)

	// Convert port string to int to pass to Run
	portInt, err := strconv.Atoi(serverPort)
	if err != nil {
		portInt = 8082
	}

	r.Run(":" + strconv.Itoa(portInt))
}
