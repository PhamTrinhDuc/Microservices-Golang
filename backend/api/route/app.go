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
		auth.POST("/google", uc.GoogleAuth)
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
		admin.GET("/users", uc.ListUsers)
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
		admin.PUT("/categories/:id", cc.UpdateCategory)
		admin.DELETE("/categories/:id", cc.DeleteCategory)

		admin.POST("/brands", cc.CreateBrand)
		admin.PUT("/brands/:id", cc.UpdateBrand)
		admin.DELETE("/brands/:id", cc.DeleteBrand)

		admin.POST("/products", cc.CreateProduct)
		admin.PUT("/products/:id", cc.UpdateProduct)
		admin.DELETE("/products/:id", cc.DeleteProduct)

		admin.POST("/option-values", cc.AddOptionValues)
		admin.POST("/products/:id/variants", cc.GenerateVariant)
	}
}

// SetupInventoryRoutes sets up store, suppliers, and inventory routes
func SetupInventoryRoutes(router *gin.RouterGroup, ic *controller.InventoryController, authMiddleware *middleware.AuthMiddleware) {
	// Public store list
	router.GET("/stores", ic.ListStores)

	// Admin control endpoints
	admin := router.Group("/admin")
	admin.Use(authMiddleware.Handler(), authMiddleware.RequireRole("admin"))
	{
		// Store management
		admin.POST("/stores", ic.CreateStore)
		admin.GET("/stores/:id", ic.GetStoreByID)
		admin.PUT("/stores/:id", ic.UpdateStore)
		admin.DELETE("/stores/:id", ic.DeactivateStore)

		// Suppliers
		admin.POST("/suppliers", ic.CreateSupplier)
		admin.GET("/suppliers", ic.ListSuppliers)
		admin.PUT("/suppliers/:id", ic.UpdateSupplier)
		admin.DELETE("/suppliers/:id", ic.DeleteSupplier)

		// Inventory & Logs
		admin.GET("/stores/:id/inventory", ic.ListStoreInventory)
		admin.PUT("/stores/:id/inventory", ic.AdjustInventory)
		admin.POST("/inventory/import", ic.ImportGoods)
		admin.GET("/inventory/imports", ic.ListImportInvoices)
		admin.GET("/inventory/imports/:id", ic.GetImportInvoiceDetails)
		admin.GET("/inventory/low-stock", ic.GetLowStockAlerts)
		admin.GET("/inventory/logs", ic.GetInventoryLogs)
	}
}

// SetupCartRoutes sets up the cart management endpoints
func SetupCartRoutes(router *gin.RouterGroup, cc *controller.CartController, authMiddleware *middleware.AuthMiddleware) {
	// Guest & Authenticated Cart Management (Optional Token)
	cart := router.Group("/cart")
	cart.Use(authMiddleware.OptionalHandler())
	{
		cart.GET("", cc.GetCart)
		cart.POST("", cc.AddToCart)
		cart.PUT("/items/:id", cc.UpdateItemQuantity)
		cart.DELETE("/items/:id", cc.RemoveItem)
		cart.DELETE("", cc.ClearCart)
	}

	// Authenticated Cart Merging (Required Token)
	authenticatedCart := router.Group("/cart")
	authenticatedCart.Use(authMiddleware.Handler())
	{
		authenticatedCart.POST("/merge", cc.MergeCart)
	}
}

// SetupPromotionVoucherRoutes sets up promotions and vouchers endpoints
func SetupPromotionVoucherRoutes(router *gin.RouterGroup, pvc *controller.PromotionVoucherController, authMiddleware *middleware.AuthMiddleware) {
	// Public / customer endpoints
	router.GET("/vouchers", pvc.ListActiveVouchers)

	// Apply voucher requires authenticated user
	customer := router.Group("")
	customer.Use(authMiddleware.Handler())
	{
		customer.POST("/vouchers/apply", pvc.ApplyVoucher)
	}

	// Admin endpoints
	admin := router.Group("/admin")
	admin.Use(authMiddleware.Handler(), authMiddleware.RequireRole("admin"))
	{
		// Promotions CRUD
		admin.POST("/promotions", pvc.CreatePromotion)
		admin.GET("/promotions", pvc.ListPromotions)
		admin.GET("/promotions/:id", pvc.GetPromotionByID)
		admin.PUT("/promotions/:id", pvc.UpdatePromotion)
		admin.DELETE("/promotions/:id", pvc.DeletePromotion)

		// Vouchers CRUD
		admin.POST("/vouchers", pvc.CreateVoucher)
		admin.GET("/vouchers", pvc.ListVouchers)
		admin.GET("/vouchers/:id", pvc.GetVoucherByID)
		admin.PUT("/vouchers/:id", pvc.UpdateVoucher)
		admin.DELETE("/vouchers/:id", pvc.DeleteVoucher)
	}
}

// SetupOrderRoutes sets up order creation, management and payment webhook routes
func SetupOrderRoutes(router *gin.RouterGroup, oc *controller.OrderController, authMiddleware *middleware.AuthMiddleware) {
	// Public webhook for PayOS payments
	router.POST("/payments/webhook", oc.ConfirmPaymentWebhook)

	// Protected customer routes
	customer := router.Group("")
	customer.Use(authMiddleware.Handler())
	{
		customer.POST("/orders/checkout", oc.Checkout)
		customer.GET("/orders", oc.ListMyOrders)
		customer.GET("/orders/:id", oc.GetMyOrderDetails)
		customer.POST("/orders/:id/cancel", oc.CancelMyOrder)
	}

	// Protected admin routes
	admin := router.Group("/admin")
	admin.Use(authMiddleware.Handler(), authMiddleware.RequireRole("admin"))
	{
		admin.GET("/orders", oc.AdminListOrders)
		admin.PUT("/orders/:id/status", oc.AdminUpdateStatus)
	}
}

// SetupBannerRoutes sets up routes for banner management
func SetupBannerRoutes(router *gin.RouterGroup, bc *controller.BannerController, authMiddleware *middleware.AuthMiddleware) {
	// Public routes
	router.GET("/banners", bc.ListPublic)

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(authMiddleware.Handler(), authMiddleware.RequireRole("admin"))
	{
		admin.POST("/banners", bc.Create)
		admin.GET("/banners", bc.ListAdmin)
		admin.PUT("/banners/:id", bc.Update)
		admin.DELETE("/banners/:id", bc.Delete)
	}
}

// SetupUploadRoutes sets up file upload routing endpoints
func SetupUploadRoutes(router *gin.RouterGroup, uc *controller.UploadController, authMiddleware *middleware.AuthMiddleware) {
	admin := router.Group("/admin")
	admin.Use(authMiddleware.Handler(), authMiddleware.RequireRole("admin"))
	{
		admin.POST("/upload", uc.UploadImage)
	}
}

