package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	mockdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao/mock"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	pkguser "github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg"
	mockuser "github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg/mock"
)

func TestGetAdByID(t *testing.T) {
	ctx := context.Background()

	t.Run("nominal", func(t *testing.T) {
		d := mockdao.New(t)
		u := mockuser.New(t)

		const adID = 42
		want := domain.Ad{ID: adID, Title: "bike", Price: 5000, UserID: "user-1", UserNickname: "alice"}

		d.On("GetAdByID", ctx, domain.AdID(adID)).Return(want, nil)
		u.On("GetUserByID", ctx, want.UserID).Return(pkguser.User{UserID: want.UserID, Nickname: "alice"}, nil)

		got, err := usecase.NewGetAdByIDFunc(d, u)(ctx, adID)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("error no ad found", func(t *testing.T) {
		d := mockdao.New(t)
		u := mockuser.New(t)

		d.On("GetAdByID", ctx, mock.Anything).Return(domain.Ad{}, adsdao.ErrNotFound)

		_, err := usecase.NewGetAdByIDFunc(d, u)(ctx, 99)

		require.Error(t, err)
		assert.ErrorIs(t, err, adsdao.ErrNotFound)
	})

	t.Run("error from user client", func(t *testing.T) {
		d := mockdao.New(t)
		u := mockuser.New(t)

		const adID = 42
		ad := domain.Ad{ID: adID, Title: "bike", Price: 5000, UserID: "user-1"}

		d.On("GetAdByID", ctx, domain.AdID(adID)).Return(ad, nil)

		clientErr := errors.New("connection reset")
		u.On("GetUserByID", ctx, ad.UserID).Return(pkguser.User{}, clientErr)

		_, err := usecase.NewGetAdByIDFunc(d, u)(ctx, adID)

		require.Error(t, err)
		assert.ErrorIs(t, err, clientErr)
	})

	t.Run("error from dao", func(t *testing.T) {
		d := mockdao.New(t)
		u := mockuser.New(t)

		dbErr := errors.New("connection reset")
		d.On("GetAdByID", ctx, mock.Anything).Return(domain.Ad{}, dbErr)

		_, err := usecase.NewGetAdByIDFunc(d, u)(ctx, 1)

		assert.ErrorIs(t, err, dbErr)
	})
}
