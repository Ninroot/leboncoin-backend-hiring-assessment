package usecase

import (
	"context"
	"fmt"
	"time"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

type (
	// CreateAdFunc creates an ad owned by the calling user. It fails with adsdao.ErrQuotaExceeded when the owner already reached their quota.
	CreateAdFunc func(ctx context.Context, req CreateAdRequest) (domain.AdID, error)

	// CreateAdRequest describes the ad to create. OwnerID is the authenticated user: a caller can never create an ad on behalf of someone else.
	CreateAdRequest struct {
		Title      string
		PriceCents int64
		PhotoURL   string
		OwnerID    domain.UserID
	}

	// Quota limits how many ads an owner can create during a sliding window. The zero value disables the limit.
	Quota struct {
		// Max is the number of ads allowed per window. Zero or less disables the limit.
		Max int
		// Window is the sliding period the limit applies to.
		Window time.Duration
	}

	createAd struct {
		adDAO adsdao.DAO
		quota Quota
	}
)

// NewCreateAdFunc returns a CreateAdFunc backed by the given DAO and enforcing the given quota.
func NewCreateAdFunc(adDAO adsdao.DAO, quota Quota) CreateAdFunc {
	return createAd{adDAO: adDAO, quota: quota}.createAd
}

func (uc createAd) createAd(ctx context.Context, req CreateAdRequest) (domain.AdID, error) {
	daoReq := adsdao.CreateAdRequest{
		Title:      req.Title,
		PriceCents: req.PriceCents,
		PhotoURL:   req.PhotoURL,
		OwnerID:    req.OwnerID,
	}

	if uc.quota.Max <= 0 {
		adID, err := uc.adDAO.CreateAd(ctx, daoReq)
		if err != nil {
			return 0, fmt.Errorf("unable to create ad: %w", err)
		}

		return adID, nil
	}

	return uc.createAdWithinQuota(ctx, daoReq)
}

// createAdWithinQuota checks the quota and inserts the ad in one transaction. The owner lock makes concurrent requests from the same owner wait for each other, so they cannot all pass the check before any of them inserts.
func (uc createAd) createAdWithinQuota(ctx context.Context, req adsdao.CreateAdRequest) (domain.AdID, error) {
	tx, err := uc.adDAO.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("unable to create ad: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := tx.LockOwner(ctx, req.OwnerID); err != nil {
		return 0, fmt.Errorf("unable to create ad: %w", err)
	}

	count, err := tx.CountAdsSince(ctx, req.OwnerID, uc.quota.Window)
	if err != nil {
		return 0, fmt.Errorf("unable to create ad: %w", err)
	}

	if count >= uc.quota.Max {
		return 0, fmt.Errorf("unable to create ad: %w", adsdao.ErrQuotaExceeded)
	}

	adID, err := tx.CreateAd(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("unable to create ad: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("unable to create ad: %w", err)
	}

	return adID, nil
}
