package usecase

import (
	"context"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

type (
	// DeleteAdByIDFunc deletes an ad by its identifier. It fails with adsdao.ErrNotFound when no ad matches the identifier.
	DeleteAdByIDFunc func(ctx context.Context, id domain.AdID) error

	deleteAd struct {
		adDAO adsdao.DAO
	}
)

func NewDeleteAdByIDFunc(adDAO adsdao.DAO) DeleteAdByIDFunc {
	return deleteAd{adDAO: adDAO}.deleteAdByID
}

func (uc deleteAd) deleteAdByID(ctx context.Context, id domain.AdID) error {
	return uc.adDAO.DeleteAdByID(ctx, id)
}
