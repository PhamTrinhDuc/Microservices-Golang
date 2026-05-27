package route

import (
	"backend/api/middleware"
	"backend/controller"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes sets up user authentication and profile routes
func SetupUserRoutes(router *gin.RouterGroup, uc *controller.UserController, authMiddleware *middleware.AuthMiddleware) {
	// Public authentication endpoints
	auth := router.Group("/auth")
	{
		auth.POST("/register", uc.Register)
		auth.POST("/login", uc.Login)
	}

	// Protected user profiles
	user := router.Group("")
	user.Use(authMiddleware.Handler())
	{
		user.GET("/profile", uc.GetMe)
	}

	// Admin control endpoints
	admin := router.Group("/admin")
	admin.Use(authMiddleware.Handler(), authMiddleware.RequireRole("admin"))
	{
		admin.PUT("/users/:id/lock", uc.LockUser)
	}
}

// SetupAddressRoutes sets up address management routes
func SetupAddressRoutes(router *gin.RouterGroup, ac *controller.AddressController, authMiddleware *middleware.AuthMiddleware) {
	address := router.Group("/addresses")
	address.Use(authMiddleware.Handler())
	{
		address.GET("", ac.List)
		address.POST("", ac.Create)
		address.PUT("/:id/default", ac.SetDefault)
		address.DELETE("/:id", ac.Delete)
	}
}

// SetupCatalogRoutes sets up all category, brand, and product/variant routes
func SetupCatalogRoutes(router *gin.RouterGroup, cc *controller.CatalogController, authMiddleware *middleware.AuthMiddleware) {
	// Public catalog routes
	router.GET("/categories", cc.ListCategories)
	router.GET("/brands", cc.ListBrands)
	router.GET("/products", cc.SearchProducts)
	router.GET("/products/:id", cc.GetProductDetails)

	// Admin catalog management
	admin := router.Group("/admin")
	admin.Use(authMiddleware.Handler(), authMiddleware.RequireRole("admin"))
	{
		admin.POST("/categories", cc.CreateCategory)
		admin.POST("/brands", cc.CreateBrand)
		admin.POST("/products", cc.CreateProduct)
		admin.POST("/option-values", cc.AddOptionValues)
		admin.POST("/products/:id/variants", cc.GenerateVariant)
	}
}
