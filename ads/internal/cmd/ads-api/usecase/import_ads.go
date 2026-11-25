package usecase

import (
	"context"
	"fmt"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

type (
	ImportAdsFunc    func(ctx context.Context, req ImportAdsRequest) ([]domain.AdID, error)
	ImportAdsRequest []CreateAdRequest

	importAd struct {
		adDAO adsdao.DAO
	}
)

// NewImportAdsFunc returns an ImportAdsFunc backed by the given DAO.
func NewImportAdsFunc(adDAO adsdao.DAO) ImportAdsFunc {
	return importAd{adDAO: adDAO}.importAd
}

func (uc importAd) importAd(ctx context.Context, req ImportAdsRequest) ([]domain.AdID, error) {
	importAdsRequest := make([]adsdao.CreateAdRequest, len(req))

	for i, ad := range req {
		importAdsRequest[i] = adsdao.CreateAdRequest{
			Title:      ad.Title,
			PriceCents: ad.PriceCents,
			PhotoURL:   ad.PhotoURL,
			OwnerID:    ad.OwnerID,
		}
	}

	adID, err := uc.adDAO.ImportAds(ctx, importAdsRequest)
	if err != nil {
		return nil, fmt.Errorf("unable to create import ads: %w", err)
	}

	return adID, nil
}
