package bootstrap

import (
	"log"

	"backend/controller"
	"backend/internal/payment"
	"backend/internal/shipping"
	"backend/internal/storage"
	"backend/internal/utils"
	"backend/repository"
	"backend/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
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
	}
}
