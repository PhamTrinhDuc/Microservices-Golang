package domain

import "errors"

var (
	// User
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailTaken      = errors.New("email already in use")
	ErrInvalidPassword = errors.New("invalid credentials")
	ErrUnauthorized    = errors.New("unauthorized access")
	ErrAddressNotFound = errors.New("address not found")
	ErrLocked          = errors.New("account is locked")
	// Catalog
	ErrCategoryNotFound      = errors.New("category not found")
	ErrBrandNotFound         = errors.New("brand not found")
	ErrProductNotFound       = errors.New("product not found")
	ErrVariantNotFound       = errors.New("variant not found")
	ErrDuplicateSlug         = errors.New("slug already in use")
	ErrProductOptionNotFound = errors.New("product option not found")
	ErrOptionValueNotFound   = errors.New("product option value not found")
	// Inventory
	ErrStoreNotFound         = errors.New("store not found")
	ErrSupplierNotFound      = errors.New("supplier not found")
	ErrImportInvoiceNotFound = errors.New("import invoice not found")
	ErrInsufficientStock     = errors.New("insufficient stock")
	// Cart
	ErrCartItemNotFound = errors.New("cart item not found")
	ErrInvalidQuantity  = errors.New("quantity must be greater than zero")
	// Vouchers
	ErrPromotionNotFound       = errors.New("promotion not found")
	ErrVoucherNotFound         = errors.New("voucher not found")
	ErrVoucherExpired          = errors.New("voucher outside of validity period")
	ErrVoucherLimitReached     = errors.New("voucher usage limit reached")
	ErrVoucherUserLimitReached = errors.New("voucher limit reached for this user")
	ErrVoucherMinAmountNotMet  = errors.New("order amount does not meet the voucher minimum value")
	// Orders
	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderCannotBeCancelled  = errors.New("order cannot be cancelled in its current state")
	ErrInvalidPaymentMethod    = errors.New("invalid payment method")
	ErrEmptyCart               = errors.New("cannot place order with an empty cart")
	ErrInvalidAddress          = errors.New("invalid receiver address information")
)

