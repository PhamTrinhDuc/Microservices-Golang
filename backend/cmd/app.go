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

type Config struct {
	Port     string
	Database bootstrap.DBConfig
}

func loadConfig() Config {
	return Config{
		Port: utils.GetEnvString("PORT", defaultPort),
		Database: bootstrap.DBConfig{
			Host:     utils.GetEnvString("DB_HOST", defaultDBHost),
			Port:     utils.GetEnvInt("DB_PORT", defaultDBPort),
			User:     utils.GetEnvString("DB_USER", "jiyuu_user"),
			Password: utils.GetEnvString("DB_PASSWORD", "jiyuu_password"),
			DBName:   utils.GetEnvString("DB_NAME", "ecommerce_db"),
			SSLMode:  utils.GetEnvString("DB_SSLMODE", "disable"),
			MaxConns: int32(utils.GetEnvInt("DB_MAX_CONNS", 25)),
			MinConns: int32(utils.GetEnvInt("DB_MIN_CONNS", 5)),
		},
	}
}

func main() {
	// Load godotenv first
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: failed to load .env file, reading system environment variables")
	}

	// Create listener context for operating system shutdown signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := loadConfig()

	// 1. Initialize database
	db, err := bootstrap.NewDB(ctx, cfg.Database)
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

	v1 := r.Group("/api/v1")
	route.SetupUserRoutes(v1, container.UserCtl, authMiddleware)
	route.SetupAddressRoutes(v1, container.AddressCtl, authMiddleware)
	route.SetupCatalogRoutes(v1, container.CatalogCtl, authMiddleware)
	route.SetupInventoryRoutes(v1, container.InventoryCtl, authMiddleware)
	route.SetupCartRoutes(v1, container.CartCtl, authMiddleware)
	route.SetupPromotionVoucherRoutes(v1, container.PromotionVoucherCtl, authMiddleware)
	route.SetupOrderRoutes(v1, container.OrderCtl, authMiddleware)

	// 5. Start Server
	serverPort := cfg.Port
	log.Printf("Starting HTTP server on port %s...", serverPort)
	
	// Convert port string to int to pass to Run
	portInt, err := strconv.Atoi(serverPort)
	if err != nil {
		portInt = 8082
	}
	
	r.Run(":" + strconv.Itoa(portInt))
}
