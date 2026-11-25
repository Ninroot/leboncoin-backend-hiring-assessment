package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	pkguser "github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg/api"
)

type (
	// GetAdByIDFunc returns a single ad, with the nickname of its owner. It fails with adsdao.ErrNotFound when no ad matches the identifier.
	// The nickname is best effort: when the identity service fails or does not know the owner, the ad is returned without it.
	GetAdByIDFunc func(ctx context.Context, id domain.AdID) (domain.Ad, error)

	getAdByID struct {
		adDAO      adsdao.DAO
		userClient pkguser.UserClient
	}
)

// NewGetAdByIDFunc returns a GetAdByIDFunc backed by the given DAO and identity service client.
func NewGetAdByIDFunc(adDAO adsdao.DAO, userClient pkguser.UserClient) GetAdByIDFunc {
	return getAdByID{adDAO: adDAO, userClient: userClient}.getAdByID
}

func (uc getAdByID) getAdByID(ctx context.Context, id domain.AdID) (domain.Ad, error) {
	found, err := uc.adDAO.GetAdByID(ctx, id)
	if err != nil {
		return domain.Ad{}, fmt.Errorf("unable to get ad %s: %w", id, err)
	}

	// The nickname is only displayed: an identity service outage must not take the ad page down with it.
	user, err := uc.userClient.GetUserByID(ctx, found.UserID)
	switch {
	case err == nil:
		found.UserNickname = user.Nickname
	case errors.Is(err, api.ErrNotFound):
		slog.WarnContext(ctx, "ad owner unknown to the identity service", "ad_id", id, "user_id", found.UserID)
	default:
		slog.ErrorContext(ctx, "unable to get ad owner", "ad_id", id, "user_id", found.UserID, "error", err)
	}

	return found, nil
}
