package http

import (
	"encoding/json"
	"net/http"

	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/middleware"
)

type importAdsBody struct {
	Ads []createAdBody `json:"ads"`
}

type importAdsResponse struct {
	IDs []domain.AdID `json:"ids"`
}

func handlerImportAds(importAds usecase.ImportAdsFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var body importAdsBody
		if err := decodeBody(r, &body); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		ownerID := domain.UserID(middleware.UserIDFromRequest(r))

		importAdsRequest := make([]usecase.CreateAdRequest, 0, len(body.Ads))

		for _, item := range body.Ads {
			createAdRequest := usecase.CreateAdRequest{
				Title:      item.Title,
				PriceCents: item.PriceCents,
				PhotoURL:   item.PhotoURL,
				OwnerID:    ownerID,
			}

			importAdsRequest = append(importAdsRequest, createAdRequest)
		}

		ids, err := importAds(r.Context(), importAdsRequest)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(rw).Encode(importAdsResponse{IDs: ids})
	}
}
