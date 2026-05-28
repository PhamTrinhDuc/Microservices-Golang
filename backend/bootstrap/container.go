package bootstrap

import (
	"backend/controller"
	"backend/internal/payment"
	"backend/internal/shipping"
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

	// Usecases
	UserUC             *usecase.UserUsecase
	AddressUC          *usecase.AddressUsecase
	CatalogUC          *usecase.CatalogUsecase
	InventoryUC        *usecase.InventoryUsecase
	CartUC             *usecase.CartUsecase
	PromotionVoucherUC *usecase.PromotionVoucherUsecase
	OrderUC            *usecase.OrderUsecase

	// Controllers
	UserCtl             *controller.UserController
	AddressCtl          *controller.AddressController
	CatalogCtl          *controller.CatalogController
	InventoryCtl        *controller.InventoryController
	CartCtl             *controller.CartController
	PromotionVoucherCtl *controller.PromotionVoucherController
	OrderCtl            *controller.OrderController
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

	// Usecases
	userUC := usecase.NewUserUsecase(userRepo)
	addressUC := usecase.NewAddressUsecase(addressRepo)
	catalogUC := usecase.NewCatalogUsecase(catalogRepo)
	inventoryUC := usecase.NewInventoryUsecase(inventoryRepo)
	cartUC := usecase.NewCartUsecase(cartRepo)
	promotionVoucherUC := usecase.NewPromotionVoucherUsecase(promotionVoucherRepo)
	orderUC := usecase.NewOrderUsecase(orderRepo, cartRepo, addressRepo, payosClient, ghnClient)

	// Controllers
	userCtl := controller.NewUserController(userUC)
	addressCtl := controller.NewAddressController(addressUC)
	catalogCtl := controller.NewCatalogController(catalogUC)
	inventoryCtl := controller.NewInventoryController(inventoryUC)
	cartCtl := controller.NewCartController(cartUC)
	promotionVoucherCtl := controller.NewPromotionVoucherController(promotionVoucherUC)
	orderCtl := controller.NewOrderController(orderUC, payosClient)

	return &Container{
		UserRepo:             userRepo,
		AddressRepo:          addressRepo,
		CatalogRepo:          catalogRepo,
		InventoryRepo:        inventoryRepo,
		CartRepo:             cartRepo,
		PromotionVoucherRepo: promotionVoucherRepo,
		OrderRepo:            orderRepo,

		UserUC:             userUC,
		AddressUC:          addressUC,
		CatalogUC:          catalogUC,
		InventoryUC:        inventoryUC,
		CartUC:             cartUC,
		PromotionVoucherUC: promotionVoucherUC,
		OrderUC:            orderUC,

		UserCtl:             userCtl,
		AddressCtl:          addressCtl,
		CatalogCtl:          catalogCtl,
		InventoryCtl:        inventoryCtl,
		CartCtl:             cartCtl,
		PromotionVoucherCtl: promotionVoucherCtl,
		OrderCtl:            orderCtl,
	}
}
