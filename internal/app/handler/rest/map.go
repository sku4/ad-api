package rest

import (
	"encoding/json"
	"hash/crc32"
	"net/http"
	"strconv"

	"github.com/joncalhoun/qson"
	"github.com/sku4/ad-api/configs"
	"github.com/sku4/ad-api/model"
	"github.com/sku4/ad-parser/pkg/logger"
)

func (h *Handler) Map(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()

	var filterFields *model.AdFilterFields
	if err := qson.Unmarshal(&filterFields, r.URL.RawQuery); err != nil {
		if err != qson.ErrInvalidParam {
			JSONHandleError(w, err)
			return
		}
		filterFields = new(model.AdFilterFields)
	}

	body, err := h.services.Bot.MapIndex(r.Context(), r.URL.RawQuery, filterFields)
	if err != nil {
		log.Warnf("map: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {
		log.Warnf("map: write error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type locsResp struct {
	Status  int                 `json:"status"`
	Message string              `json:"message"`
	Result  []*model.AdLocation `json:"result"`
}

func (h *Handler) Locations(w http.ResponseWriter, r *http.Request) {
	cfg := configs.Get(r.Context())

	var filterFields *model.AdFilterFields
	if err := json.NewDecoder(r.Body).Decode(&filterFields); err != nil {
		JSONHandleError(w, err)
		return
	}

	filterJSON, err := json.Marshal(filterFields)
	if err != nil {
		JSONHandleError(w, err)
		return
	}

	cacheKey := strconv.Itoa(int(crc32.Checksum(filterJSON, h.crcTable)))
	cacheData := h.cache.Get(cacheKey)
	if cacheData != nil && !cacheData.Expired() {
		JSONHandleResp(w, cacheData.Value().([]byte))
		return
	}

	adTuple, err := filterFields.ConvertToTuple()
	if err != nil {
		JSONHandleError(w, err)
		return
	}

	locs, err := h.services.MapFilter(r.Context(), adTuple)
	if err != nil {
		JSONHandleError(w, err)
		return
	}

	locsJSON, err := json.Marshal(locsResp{
		Status:  http.StatusOK,
		Result:  locs,
		Message: "OK",
	})
	if err != nil {
		JSONHandleError(w, err)
		return
	}

	h.cache.Set(cacheKey, locsJSON, cfg.App.Map.CacheDuration)

	JSONHandleResp(w, locsJSON)
}

type adsResp struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Result  []*model.Ad `json:"result"`
}

func (h *Handler) Ads(w http.ResponseWriter, r *http.Request) {
	var adsFilter *model.AdsFilter
	if err := json.NewDecoder(r.Body).Decode(&adsFilter); err != nil {
		JSONHandleError(w, err)
		return
	}

	ads, err := h.services.MapAds(r.Context(), adsFilter.Groups)
	if err != nil {
		JSONHandleError(w, err)
		return
	}

	adsJSON, err := json.Marshal(adsResp{
		Status:  http.StatusOK,
		Result:  ads,
		Message: "OK",
	})
	if err != nil {
		JSONHandleError(w, err)
		return
	}

	JSONHandleResp(w, adsJSON)
}
