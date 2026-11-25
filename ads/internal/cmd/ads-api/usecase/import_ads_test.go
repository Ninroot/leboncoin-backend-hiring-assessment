package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao/mock"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
)

func TestImportAds(t *testing.T) {
	ctx := context.Background()

	ucReq := usecase.ImportAdsRequest{
		{Title: "mountain bike", PriceCents: 15000, PhotoURL: "https://example.com/bike.jpg", OwnerID: "user-1"},
		{Title: "car", PriceCents: 30000, PhotoURL: "https://example.com/car.jpg", OwnerID: "user-1"},
	}
	bikeReq := adsdao.CreateAdRequest{Title: "mountain bike", PriceCents: 15000, PhotoURL: "https://example.com/bike.jpg", OwnerID: "user-1"}
	carReq := adsdao.CreateAdRequest{Title: "car", PriceCents: 30000, PhotoURL: "https://example.com/car.jpg", OwnerID: "user-1"}

	t.Run("creates every ad in one transaction and returns the IDs in order", func(t *testing.T) {
		m := mock.New(t)
		m.On("Begin", ctx).Return(nil).Once()
		m.On("CreateAd", ctx, bikeReq).Return(domain.AdID(1), nil).Once()
		m.On("CreateAd", ctx, carReq).Return(domain.AdID(2), nil).Once()
		m.On("Commit").Return(nil).Once()
		m.On("Rollback").Return(nil).Once()

		got, err := usecase.NewImportAdsFunc(m)(ctx, ucReq)

		require.NoError(t, err)
		assert.Equal(t, []domain.AdID{1, 2}, got)
	})

	t.Run("returns an empty, non-nil slice when there is nothing to import", func(t *testing.T) {
		m := mock.New(t)
		m.On("Begin", ctx).Return(nil).Once()
		m.On("Commit").Return(nil).Once()
		m.On("Rollback").Return(nil).Once()

		got, err := usecase.NewImportAdsFunc(m)(ctx, usecase.ImportAdsRequest{})

		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("does not apply the creation quota", func(t *testing.T) {
		// The mock fails the test on any call it does not expect, so a LockOwner or CountAdsSince call would make this test fail.
		m := mock.New(t)
		m.On("Begin", ctx).Return(nil).Once()
		m.On("CreateAd", ctx, bikeReq).Return(domain.AdID(1), nil).Once()
		m.On("CreateAd", ctx, carReq).Return(domain.AdID(2), nil).Once()
		m.On("Commit").Return(nil).Once()
		m.On("Rollback").Return(nil).Once()

		_, err := usecase.NewImportAdsFunc(m)(ctx, ucReq)

		require.NoError(t, err)
	})

	t.Run("fails when the transaction cannot begin", func(t *testing.T) {
		m := mock.New(t)
		dbErr := errors.New("db down")
		m.On("Begin", ctx).Return(dbErr).Once()

		_, err := usecase.NewImportAdsFunc(m)(ctx, ucReq)

		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("rolls back and stops at the first ad that fails", func(t *testing.T) {
		m := mock.New(t)
		m.On("Begin", ctx).Return(nil).Once()
		m.On("CreateAd", ctx, bikeReq).Return(domain.AdID(0), adsdao.ErrUnknownOwner).Once()
		m.On("Rollback").Return(nil).Once()

		got, err := usecase.NewImportAdsFunc(m)(ctx, ucReq)

		require.Error(t, err)
		assert.ErrorIs(t, err, adsdao.ErrUnknownOwner)
		assert.Nil(t, got)
		m.AssertNotCalled(t, "CreateAd", ctx, carReq)
		m.AssertNotCalled(t, "Commit")
	})

	t.Run("fails when the commit fails", func(t *testing.T) {
		m := mock.New(t)
		commitErr := errors.New("commit failed")
		m.On("Begin", ctx).Return(nil).Once()
		m.On("CreateAd", ctx, bikeReq).Return(domain.AdID(1), nil).Once()
		m.On("CreateAd", ctx, carReq).Return(domain.AdID(2), nil).Once()
		m.On("Commit").Return(commitErr).Once()
		m.On("Rollback").Return(nil).Once()

		got, err := usecase.NewImportAdsFunc(m)(ctx, ucReq)

		assert.ErrorIs(t, err, commitErr)
		assert.Nil(t, got)
	})
}
