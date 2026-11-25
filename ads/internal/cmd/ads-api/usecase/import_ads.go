package usecase

import (
	"context"
	"fmt"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

type (
	ImportAdFunc    func(ctx context.Context, req ImportAdRequest) ([]domain.AdID, error)
	ImportAdRequest []CreateAdRequest

	importAd struct {
		adDAO adsdao.DAO
	}
)

// NewImportAdFunc returns an ImportAdFunc backed by the given DAO.
func NewImportAdFunc(adDAO adsdao.DAO) ImportAdFunc {
	return importAd{adDAO: adDAO}.importAd
}

func (uc importAd) importAd(ctx context.Context, req ImportAdRequest) ([]domain.AdID, error) {
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
		return nil, fmt.Errorf("unable to create ad: %w", err)
	}

	return adID, nil
}
