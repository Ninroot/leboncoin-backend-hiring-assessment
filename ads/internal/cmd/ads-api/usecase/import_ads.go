package usecase

import (
	"context"
	"fmt"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

type (
	// ImportAdsFunc creates several ads, all or nothing: if one of them cannot be created, none is. It returns the generated identifiers in the order of the request. The ad creation quota does not apply.
	ImportAdsFunc func(ctx context.Context, req ImportAdsRequest) ([]domain.AdID, error)

	// ImportAdsRequest lists the ads to create.
	ImportAdsRequest []CreateAdRequest

	importAds struct {
		adDAO adsdao.DAO
	}
)

// NewImportAdsFunc returns an ImportAdsFunc backed by the given DAO.
func NewImportAdsFunc(adDAO adsdao.DAO) ImportAdsFunc {
	return importAds{adDAO: adDAO}.importAds
}

// importAds inserts every ad in one transaction.
func (uc importAds) importAds(ctx context.Context, req ImportAdsRequest) ([]domain.AdID, error) {
	tx, err := uc.adDAO.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to import ads: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	ids := make([]domain.AdID, 0, len(req))

	for i, ad := range req {
		id, err := tx.CreateAd(ctx, adsdao.CreateAdRequest{
			Title:      ad.Title,
			PriceCents: ad.PriceCents,
			PhotoURL:   ad.PhotoURL,
			OwnerID:    ad.OwnerID,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to import ads: ad %d: %w", i, err)
		}

		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("unable to import ads: %w", err)
	}

	return ids, nil
}
