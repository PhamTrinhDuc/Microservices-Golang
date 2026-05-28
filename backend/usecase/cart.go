package usecase

import (
	"context"
	"fmt"

	"backend/domain"
)

type CartUsecase struct {
	repo domain.CartRepository
}

func NewCartUsecase(repo domain.CartRepository) *CartUsecase {
	return &CartUsecase{repo: repo}
}

func (uc *CartUsecase) GetCart(ctx context.Context, userID *int, sessionID *string) ([]*domain.CartItemResponse, error) {
	if userID == nil && (sessionID == nil || *sessionID == "") {
		return nil, fmt.Errorf("either user_id or session_id must be provided")
	}
	return uc.repo.GetCartDetails(ctx, userID, sessionID)
}

func (uc *CartUsecase) AddToCart(ctx context.Context, userID *int, sessionID *string, req *domain.AddToCartRequest) (*domain.CartItem, error) {
	if userID == nil && (sessionID == nil || *sessionID == "") {
		return nil, fmt.Errorf("either user_id or session_id must be provided")
	}

	if req.Quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}

	// 1. Verify variant exists and is active
	exists, err := uc.repo.VerifyVariantExists(ctx, req.VariantID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrVariantNotFound
	}

	// 2. Check if item already in cart
	existingItem, err := uc.repo.FindCartItem(ctx, userID, sessionID, req.VariantID)
	if err != nil {
		return nil, err
	}

	if existingItem != nil {
		// Update quantity
		newQty := existingItem.Quantity + req.Quantity
		return uc.repo.UpdateCartItemQuantity(ctx, existingItem.ID, newQty)
	}

	// Create new cart item
	newItem := &domain.CartItem{
		UserID:    userID,
		SessionID: sessionID,
		VariantID: req.VariantID,
		Quantity:  req.Quantity,
	}

	return uc.repo.CreateCartItem(ctx, newItem)
}

func (uc *CartUsecase) UpdateItemQuantity(ctx context.Context, userID *int, sessionID *string, itemID int, quantity int) (*domain.CartItem, error) {
	if userID == nil && (sessionID == nil || *sessionID == "") {
		return nil, fmt.Errorf("either user_id or session_id must be provided")
	}

	if quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}

	// 1. Fetch item
	item, err := uc.repo.GetCartItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	// 2. Verify ownership
	if err := uc.verifyOwnership(item, userID, sessionID); err != nil {
		return nil, err
	}

	return uc.repo.UpdateCartItemQuantity(ctx, itemID, quantity)
}

func (uc *CartUsecase) RemoveItem(ctx context.Context, userID *int, sessionID *string, itemID int) error {
	if userID == nil && (sessionID == nil || *sessionID == "") {
		return fmt.Errorf("either user_id or session_id must be provided")
	}

	// 1. Fetch item
	item, err := uc.repo.GetCartItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	// 2. Verify ownership
	if err := uc.verifyOwnership(item, userID, sessionID); err != nil {
		return err
	}

	return uc.repo.DeleteCartItem(ctx, itemID)
}

func (uc *CartUsecase) ClearCart(ctx context.Context, userID *int, sessionID *string) error {
	if userID == nil && (sessionID == nil || *sessionID == "") {
		return fmt.Errorf("either user_id or session_id must be provided")
	}
	return uc.repo.ClearCart(ctx, userID, sessionID)
}

func (uc *CartUsecase) MergeCart(ctx context.Context, userID int, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session_id must be provided")
	}

	// 1. List guest items
	guestItems, err := uc.repo.ListCartItems(ctx, nil, &sessionID)
	if err != nil {
		return err
	}
	if len(guestItems) == 0 {
		return nil // Nothing to merge
	}

	// 2. List user items
	userItems, err := uc.repo.ListCartItems(ctx, &userID, nil)
	if err != nil {
		return err
	}

	// Build a map of user variant_id to item
	userItemsMap := make(map[int]*domain.CartItem)
	for _, item := range userItems {
		userItemsMap[item.VariantID] = item
	}

	// 3. Merge
	for _, guestItem := range guestItems {
		if userItem, exists := userItemsMap[guestItem.VariantID]; exists {
			// Update user quantity and delete guest item
			newQty := userItem.Quantity + guestItem.Quantity
			_, err = uc.repo.UpdateCartItemQuantity(ctx, userItem.ID, newQty)
			if err != nil {
				return err
			}
			err = uc.repo.DeleteCartItem(ctx, guestItem.ID)
			if err != nil {
				return err
			}
		} else {
			// Reassign guest item to user
			err = uc.repo.LinkGuestItemToUser(ctx, guestItem.ID, userID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (uc *CartUsecase) verifyOwnership(item *domain.CartItem, userID *int, sessionID *string) error {
	if userID != nil {
		if item.UserID == nil || *item.UserID != *userID {
			return domain.ErrUnauthorized
		}
		return nil
	}

	if sessionID != nil {
		if item.SessionID == nil || *item.SessionID != *sessionID {
			return domain.ErrUnauthorized
		}
		return nil
	}

	return domain.ErrUnauthorized
}
