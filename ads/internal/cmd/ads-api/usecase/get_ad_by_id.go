package usecase

import (
	"context"
	"fmt"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	pkguser "github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg"
)

type (
	// GetAdByIDFunc returns a single ad. It fails with adsdao.ErrNotFound when no ad matches the identifier.
	GetAdByIDFunc func(ctx context.Context, id domain.AdID) (domain.Ad, error)

	getAdByID struct {
		adDAO      adsdao.DAO
		userClient pkguser.UserClient
	}
)

// NewGetAdByIDFunc returns a GetAdByIDFunc backed by the given DAO.
func NewGetAdByIDFunc(adDAO adsdao.DAO, userClient pkguser.UserClient) GetAdByIDFunc {
	return getAdByID{adDAO: adDAO, userClient: userClient}.getAdByID
}

func (uc getAdByID) getAdByID(ctx context.Context, id domain.AdID) (domain.Ad, error) {
	found, err := uc.adDAO.GetAdByID(ctx, id)
	if err != nil {
		return domain.Ad{}, fmt.Errorf("unable to get ad %s: %w", id, err)
	}

	user, err := uc.userClient.GetUserByID(ctx, found.UserID)
	if err != nil {
		return domain.Ad{}, fmt.Errorf("unable to get user %s: %w", found.UserID, err)
	}

	found.UserNickname = user.Nickname

	return found, nil
}
