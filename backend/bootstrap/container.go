package bootstrap

import (
	"context"
	"log"
	"log/slog"
	"os"

	"backend/controller"
	"backend/internal/payment"
	"backend/internal/shipping"
	"backend/internal/storage"
	"backend/internal/utils"
	"backend/repository"
	"backend/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	"indexing/chunker"
	"indexing/db"
	"indexing/embedder"
	"indexing/ingestion"
)

type Container struct {
	// Repositories
	UserRepo             *repository.UserRepository
	AddressRepo          *repository.AddressRepository
	CatalogRepo          *repository.CatalogRepository
	InventoryRepo        *repository.InventoryRepository
	CartRepo             *repository.CartRepository
	PromotionVoucherRepo *repository.PromotionVoucherRepository
	OrderRepo            *repository.OrderRepository
	BannerRepo           *repository.BannerRepository
	WishlistRepo         *repository.WishlistRepository
	AnalyticsRepo        *repository.AnalyticsRepository
	PolicyRepo           *repository.PolicyRepository

	// Usecases
	UserUC             *usecase.UserUsecase
	AddressUC          *usecase.AddressUsecase
	CatalogUC          *usecase.CatalogUsecase
	InventoryUC        *usecase.InventoryUsecase
	CartUC             *usecase.CartUsecase
	PromotionVoucherUC *usecase.PromotionVoucherUsecase
	OrderUC            *usecase.OrderUsecase
	BannerUC           *usecase.BannerUsecase
	WishlistUC         *usecase.WishlistUsecase
	AnalyticsUC        *usecase.AnalyticsUsecase
	PolicyUC           *usecase.PolicyUsecase

	// Controllers
	UserCtl             *controller.UserController
	AddressCtl          *controller.AddressController
	CatalogCtl          *controller.CatalogController
	InventoryCtl        *controller.InventoryController
	CartCtl             *controller.CartController
	PromotionVoucherCtl *controller.PromotionVoucherController
	OrderCtl            *controller.OrderController
	BannerCtl           *controller.BannerController
	UploadCtl           *controller.UploadController
	WishlistCtl         *controller.WishlistController
	AnalyticsCtl        *controller.AnalyticsController
	PolicyCtl           *controller.PolicyController
}

func NewContainer(pool *pgxpool.Pool) *Container {
	// External Integration Clients
	payosClient := payment.NewPayOSClient()
	ghnClient := shipping.NewGHNClient()

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	addressRepo := repository.NewAddressRepository(pool)
	catalogRepo := repository.NewCatalogRepository(pool)
	inventoryRepo := repository.NewInventoryRepository(pool)
	cartRepo := repository.NewCartRepository(pool)
	promotionVoucherRepo := repository.NewPromotionVoucherRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)
	bannerRepo := repository.NewBannerRepository(pool)
	wishlistRepo := repository.NewWishlistRepository(pool)
	analyticsRepo := repository.NewAnalyticsRepository(pool)
	policyRepo := repository.NewPolicyRepository(pool)

	// Initialize Indexing (RAG) library dependencies
	embedderBaseURL := utils.GetEnvString("OPENAI_BASE_URL", "https://api.openai.com")
	embedderAPIKey := utils.GetEnvString("OPENAI_API_KEY", "")
	embedderModel := utils.GetEnvString("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small")
	embedderDims := utils.GetEnvInt("OPENAI_EMBEDDING_DIMENSIONS", 1024)

	emb := embedder.New(embedderBaseURL, embedderAPIKey, embedderModel)
	chk := chunker.New(500, 50)

	connString := utils.GetEnvString("CONN_POSTGRES", "postgres://jiyuu_user:jiyuu_password@localhost:5433/ecommerce_db")
	indexingStore, err := db.NewStore(context.Background(), connString, embedderDims)
	if err != nil {
		log.Printf("Warning: failed to initialize indexing database store: %v", err)
	}

	// Slog logger
	slogHandler := slog.NewJSONHandler(os.Stdout, nil)
	slogLogger := slog.New(slogHandler)

	syncManager := ingestion.NewSyncManager(indexingStore, emb, chk, slogLogger)

	// Usecases
	userUC := usecase.NewUserUsecase(userRepo)
	addressUC := usecase.NewAddressUsecase(addressRepo)
	catalogUC := usecase.NewCatalogUsecase(catalogRepo)
	inventoryUC := usecase.NewInventoryUsecase(inventoryRepo)
	cartUC := usecase.NewCartUsecase(cartRepo)
	promotionVoucherUC := usecase.NewPromotionVoucherUsecase(promotionVoucherRepo)
	orderUC := usecase.NewOrderUsecase(orderRepo, cartRepo, addressRepo, payosClient, ghnClient)
	bannerUC := usecase.NewBannerUsecase(bannerRepo)
	wishlistUC := usecase.NewWishlistUsecase(wishlistRepo)
	analyticsUC := usecase.NewAnalyticsUsecase(analyticsRepo)
	policyUC := usecase.NewPolicyUsecase(policyRepo, emb, syncManager)

	// Initialize Storage Provider
	storageType := utils.GetEnvString("STORAGE_PROVIDER", "local")
	baseURL := utils.GetEnvString("APP_BASE_URL", "http://localhost:8082")

	var storageProvider storage.StorageProvider
	if storageType == "s3" {
		bucket := utils.GetEnvString("S3_BUCKET", "")
		region := utils.GetEnvString("S3_REGION", "us-east-1")
		accessKey := utils.GetEnvString("S3_ACCESS_KEY", "")
		secretKey := utils.GetEnvString("S3_SECRET_KEY", "")
		endpoint := utils.GetEnvString("S3_ENDPOINT", "")

		sp, err := storage.NewS3StorageProvider(bucket, region, accessKey, secretKey, endpoint)
		if err != nil {
			log.Fatalf("failed to initialize S3 storage provider: %v", err)
		}
		storageProvider = sp
	} else {
		staticDir := utils.GetEnvString("LOCAL_STATIC_DIR", "./static/uploads")
		storageProvider = storage.NewLocalStorageProvider(staticDir, baseURL)
	}

	// Controllers
	userCtl := controller.NewUserController(userUC)
	addressCtl := controller.NewAddressController(addressUC)
	catalogCtl := controller.NewCatalogController(catalogUC)
	inventoryCtl := controller.NewInventoryController(inventoryUC)
	cartCtl := controller.NewCartController(cartUC)
	promotionVoucherCtl := controller.NewPromotionVoucherController(promotionVoucherUC)
	orderCtl := controller.NewOrderController(orderUC, payosClient)
	bannerCtl := controller.NewBannerController(bannerUC)
	uploadCtl := controller.NewUploadController(storageProvider)
	wishlistCtl := controller.NewWishlistController(wishlistUC)
	analyticsCtl := controller.NewAnalyticsController(analyticsUC)
	policyCtl := controller.NewPolicyController(policyUC)

	return &Container{
		UserRepo:             userRepo,
		AddressRepo:          addressRepo,
		CatalogRepo:          catalogRepo,
		InventoryRepo:        inventoryRepo,
		CartRepo:             cartRepo,
		PromotionVoucherRepo: promotionVoucherRepo,
		OrderRepo:            orderRepo,
		BannerRepo:           bannerRepo,
		WishlistRepo:         wishlistRepo,
		AnalyticsRepo:        analyticsRepo,
		PolicyRepo:           policyRepo,

		UserUC:             userUC,
		AddressUC:          addressUC,
		CatalogUC:          catalogUC,
		InventoryUC:        inventoryUC,
		CartUC:             cartUC,
		PromotionVoucherUC: promotionVoucherUC,
		OrderUC:            orderUC,
		BannerUC:           bannerUC,
		WishlistUC:         wishlistUC,
		AnalyticsUC:        analyticsUC,
		PolicyUC:           policyUC,

		UserCtl:             userCtl,
		AddressCtl:          addressCtl,
		CatalogCtl:          catalogCtl,
		InventoryCtl:        inventoryCtl,
		CartCtl:             cartCtl,
		PromotionVoucherCtl: promotionVoucherCtl,
		OrderCtl:            orderCtl,
		BannerCtl:           bannerCtl,
		UploadCtl:           uploadCtl,
		WishlistCtl:         wishlistCtl,
		AnalyticsCtl:        analyticsCtl,
		PolicyCtl:           policyCtl,
	}
}
