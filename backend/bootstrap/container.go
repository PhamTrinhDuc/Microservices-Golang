package bootstrap

import (
	"backend/controller"
	"backend/repository"
	"backend/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	// Repositories
	UserRepo    *repository.UserRepository
	AddressRepo *repository.AddressRepository
	CatalogRepo *repository.CatalogRepository

	// Usecases
	UserUC    *usecase.UserUsecase
	AddressUC *usecase.AddressUsecase
	CatalogUC *usecase.CatalogUsecase

	// Controllers
	UserCtl    *controller.UserController
	AddressCtl *controller.AddressController
	CatalogCtl *controller.CatalogController
}

func NewContainer(pool *pgxpool.Pool) *Container {
	// Repositories
	userRepo := repository.NewUserRepository(pool)
	addressRepo := repository.NewAddressRepository(pool)
	catalogRepo := repository.NewCatalogRepository(pool)

	// Usecases
	userUC := usecase.NewUserUsecase(userRepo)
	addressUC := usecase.NewAddressUsecase(addressRepo)
	catalogUC := usecase.NewCatalogUsecase(catalogRepo)

	// Controllers
	userCtl := controller.NewUserController(userUC)
	addressCtl := controller.NewAddressController(addressUC)
	catalogCtl := controller.NewCatalogController(catalogUC)

	return &Container{
		UserRepo:    userRepo,
		AddressRepo: addressRepo,
		CatalogRepo: catalogRepo,

		UserUC:    userUC,
		AddressUC: addressUC,
		CatalogUC: catalogUC,

		UserCtl:    userCtl,
		AddressCtl: addressCtl,
		CatalogCtl: catalogCtl,
	}
}
