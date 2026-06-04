package usecase

import (
	"context"

	"backend/domain"
)

type BannerUsecase struct {
	repo domain.BannerRepository
}

func NewBannerUsecase(repo domain.BannerRepository) *BannerUsecase {
	return &BannerUsecase{repo: repo}
}

func (u *BannerUsecase) Create(ctx context.Context, req *domain.CreateBannerRequest) (*domain.Banner, error) {
	b := &domain.Banner{
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Tag:         req.Tag,
		LinkURL:     req.LinkURL,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
		CategoryID:  req.CategoryID,
	}
	return u.repo.Create(ctx, b)
}

func (u *BannerUsecase) GetByID(ctx context.Context, id int) (*domain.Banner, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *BannerUsecase) List(ctx context.Context, onlyActive bool, categoryID *int) ([]*domain.Banner, error) {
	return u.repo.List(ctx, onlyActive, categoryID)
}

func (u *BannerUsecase) Update(ctx context.Context, id int, req *domain.UpdateBannerRequest) (*domain.Banner, error) {
	b, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	b.Title = req.Title
	b.Subtitle = req.Subtitle
	b.Description = req.Description
	b.ImageURL = req.ImageURL
	b.Tag = req.Tag
	b.LinkURL = req.LinkURL
	b.SortOrder = req.SortOrder
	b.IsActive = req.IsActive
	b.CategoryID = req.CategoryID

	return u.repo.Update(ctx, b)
}

func (u *BannerUsecase) Delete(ctx context.Context, id int) error {
	return u.repo.Delete(ctx, id)
}
