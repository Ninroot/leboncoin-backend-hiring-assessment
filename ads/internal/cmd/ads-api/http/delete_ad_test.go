package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/middleware"
)

func newDeleteAdHandler(fn usecase.DeleteAdByIDFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("DELETE /ads/{adID}", middleware.RequireAuth(handlerDeleteAd(fn)))

	return httptest.NewServer(mux)
}

func TestHandlerDeleteAd(t *testing.T) {
	t.Run("returns 204 on success", func(t *testing.T) {
		srv := newDeleteAdHandler(func(_ context.Context, _ domain.AdID) error {
			return nil
		})
		defer srv.Close()

		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/ads/1", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "user-1")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("returns 404 when ad not found", func(t *testing.T) {
		srv := newDeleteAdHandler(func(_ context.Context, _ domain.AdID) error {
			return dao.ErrNotFound
		})
		defer srv.Close()

		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/ads/1", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "user-1")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("returns 500 on internal error", func(t *testing.T) {
		srv := newDeleteAdHandler(func(_ context.Context, _ domain.AdID) error {
			return errors.New("internal error")
		})
		defer srv.Close()

		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/ads/1", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "user-1")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
