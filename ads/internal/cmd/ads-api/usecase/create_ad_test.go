package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao/mock"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
)

func TestCreateAd(t *testing.T) {
	ctx := context.Background()

	ucReq := usecase.CreateAdRequest{
		Title:      "mountain bike",
		PriceCents: 15000,
		PhotoURL:   "https://example.com/bike.jpg",
		OwnerID:    "user-1",
	}
	daoReq := adsdao.CreateAdRequest{
		Title:      "mountain bike",
		PriceCents: 15000,
		PhotoURL:   "https://example.com/bike.jpg",
		OwnerID:    "user-1",
	}

	t.Run("nominal", func(t *testing.T) {
		m := mock.New(t)
		m.On("CreateAd", ctx, daoReq).Return(domain.AdID(7), nil)

		got, err := usecase.NewCreateAdFunc(m, usecase.Quota{})(ctx, ucReq)

		require.NoError(t, err)
		assert.Equal(t, domain.AdID(7), got)
	})

	t.Run("error from dao", func(t *testing.T) {
		m := mock.New(t)
		dbErr := errors.New("connection reset")
		m.On("CreateAd", ctx, daoReq).Return(domain.AdID(0), dbErr)

		_, err := usecase.NewCreateAdFunc(m, usecase.Quota{})(ctx, ucReq)

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestCreateAdWithQuota(t *testing.T) {
	ctx := context.Background()
	owner := domain.UserID("user-1")
	quota := usecase.Quota{Max: 3, Window: time.Hour}

	ucReq := usecase.CreateAdRequest{Title: "mountain bike", PriceCents: 15000, PhotoURL: "https://example.com/bike.jpg", OwnerID: owner}
	daoReq := adsdao.CreateAdRequest{Title: "mountain bike", PriceCents: 15000, PhotoURL: "https://example.com/bike.jpg", OwnerID: owner}

	t.Run("creates the ad and commits when under quota", func(t *testing.T) {
		m := mock.New(t)
		m.On("Begin", ctx).Return(nil).Once()
		m.On("LockOwner", ctx, owner).Return(nil).Once()
		m.On("CountAdsSince", ctx, owner, time.Hour).Return(2, nil).Once()
		m.On("CreateAd", ctx, daoReq).Return(domain.AdID(7), nil).Once()
		m.On("Commit").Return(nil).Once()
		m.On("Rollback").Return(nil).Once()

		got, err := usecase.NewCreateAdFunc(m, quota)(ctx, ucReq)

		require.NoError(t, err)
		assert.Equal(t, domain.AdID(7), got)
	})

	t.Run("fails with ErrQuotaExceeded and does not insert when quota is reached", func(t *testing.T) {
		m := mock.New(t)
		m.On("Begin", ctx).Return(nil).Once()
		m.On("LockOwner", ctx, owner).Return(nil).Once()
		m.On("CountAdsSince", ctx, owner, time.Hour).Return(3, nil).Once()
		m.On("Rollback").Return(nil).Once()

		_, err := usecase.NewCreateAdFunc(m, quota)(ctx, ucReq)

		require.Error(t, err)
		assert.ErrorIs(t, err, adsdao.ErrQuotaExceeded)
		m.AssertNotCalled(t, "CreateAd", ctx, daoReq)
		m.AssertNotCalled(t, "Commit")
	})

	t.Run("fails when the transaction cannot begin", func(t *testing.T) {
		m := mock.New(t)
		dbErr := errors.New("db down")
		m.On("Begin", ctx).Return(dbErr).Once()

		_, err := usecase.NewCreateAdFunc(m, quota)(ctx, ucReq)

		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("fails and does not count when the owner cannot be locked", func(t *testing.T) {
		m := mock.New(t)
		lockErr := errors.New("lock timeout")
		m.On("Begin", ctx).Return(nil).Once()
		m.On("LockOwner", ctx, owner).Return(lockErr).Once()
		m.On("Rollback").Return(nil).Once()

		_, err := usecase.NewCreateAdFunc(m, quota)(ctx, ucReq)

		assert.ErrorIs(t, err, lockErr)
	})

	t.Run("fails closed when the count fails", func(t *testing.T) {
		m := mock.New(t)
		countErr := errors.New("connection reset")
		m.On("Begin", ctx).Return(nil).Once()
		m.On("LockOwner", ctx, owner).Return(nil).Once()
		m.On("CountAdsSince", ctx, owner, time.Hour).Return(0, countErr).Once()
		m.On("Rollback").Return(nil).Once()

		_, err := usecase.NewCreateAdFunc(m, quota)(ctx, ucReq)

		assert.ErrorIs(t, err, countErr)
	})

	t.Run("fails when the insert fails", func(t *testing.T) {
		m := mock.New(t)
		m.On("Begin", ctx).Return(nil).Once()
		m.On("LockOwner", ctx, owner).Return(nil).Once()
		m.On("CountAdsSince", ctx, owner, time.Hour).Return(0, nil).Once()
		m.On("CreateAd", ctx, daoReq).Return(domain.AdID(0), adsdao.ErrUnknownOwner).Once()
		m.On("Rollback").Return(nil).Once()

		_, err := usecase.NewCreateAdFunc(m, quota)(ctx, ucReq)

		assert.ErrorIs(t, err, adsdao.ErrUnknownOwner)
	})

	t.Run("fails when the commit fails", func(t *testing.T) {
		m := mock.New(t)
		commitErr := errors.New("commit failed")
		m.On("Begin", ctx).Return(nil).Once()
		m.On("LockOwner", ctx, owner).Return(nil).Once()
		m.On("CountAdsSince", ctx, owner, time.Hour).Return(0, nil).Once()
		m.On("CreateAd", ctx, daoReq).Return(domain.AdID(7), nil).Once()
		m.On("Commit").Return(commitErr).Once()
		m.On("Rollback").Return(nil).Once()

		_, err := usecase.NewCreateAdFunc(m, quota)(ctx, ucReq)

		assert.ErrorIs(t, err, commitErr)
	})
}
