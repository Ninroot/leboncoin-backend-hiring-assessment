package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	mock "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao/mock"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

func TestImportAd(t *testing.T) {
	ctx := context.Background()

	var ucReq ImportAdRequest = []CreateAdRequest{
		{
			Title:      "mountain bike",
			PriceCents: 15000,
			PhotoURL:   "https://example.com/bike.jpg",
			OwnerID:    "user-1",
		},
		{
			Title:      "car",
			PriceCents: 30000,
			PhotoURL:   "https://example.com/car.jpg",
			OwnerID:    "user-2",
		},
	}

	var daoReqs adsdao.ImportAdsRequest = []adsdao.CreateAdRequest{
		{
			Title:      "mountain bike",
			PriceCents: 15000,
			PhotoURL:   "https://example.com/bike.jpg",
			OwnerID:    "user-1",
		},
		{
			Title:      "car",
			PriceCents: 30000,
			PhotoURL:   "https://example.com/car.jpg",
			OwnerID:    "user-2",
		},
	}

	t.Run("nominal", func(t *testing.T) {

		m := mock.New(t)
		m.On("ImportAds", ctx, daoReqs).Return([]domain.AdID{1, 2}, nil)

		got, err := NewImportAdFunc(m)(ctx, ucReq)

		require.NoError(t, err)
		require.Equal(t, []domain.AdID{1, 2}, got)
	})

	t.Run("error from dao", func(t *testing.T) {
		m := mock.New(t)
		dbErr := errors.New("connection reset")
		m.On("ImportAds", ctx, daoReqs).Return(nil, dbErr)

		_, err := NewImportAdFunc(m)(ctx, ucReq)

		require.Error(t, err)
		require.ErrorIs(t, err, dbErr)
	})
}
