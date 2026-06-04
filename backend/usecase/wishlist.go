package usecase

import (
	"context"
	"errors"
	"fmt"

	"backend/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

type WishlistUsecase struct {
	wishlistRepo domain.WishlistRepository
}

func NewWishlistUsecase(wishlistRepo domain.WishlistRepository) *WishlistUsecase {
	return &WishlistUsecase{
		wishlistRepo: wishlistRepo,
	}
}

func (u *WishlistUsecase) AddToWishlist(ctx context.Context, userID int, variantID int) (*domain.WishlistItem, error) {
	item, err := u.wishlistRepo.Create(ctx, userID, variantID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign key violation
			return nil, fmt.Errorf("product variant with id %d not found: %w", variantID, domain.ErrVariantNotFound)
		}
		return nil, err
	}
	return item, nil
}

func (u *WishlistUsecase) RemoveFromWishlist(ctx context.Context, userID int, variantID int) error {
	return u.wishlistRepo.Delete(ctx, userID, variantID)
}

func (u *WishlistUsecase) GetWishlist(ctx context.Context, userID int) ([]*domain.WishlistItemResponse, error) {
	return u.wishlistRepo.GetByUserID(ctx, userID)
}
