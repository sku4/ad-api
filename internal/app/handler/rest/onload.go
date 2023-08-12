package rest

import (
	"context"
	"encoding/json"
	"hash/crc32"
	"net/http"
	"strconv"
	"time"

	"github.com/sku4/ad-api/configs"
	"github.com/sku4/ad-parser/pkg/logger"
)

func (h *Handler) OnLoad(ctx context.Context) {
	h.HotCache(ctx)
	h.SetMetrics(ctx)
}

func (h *Handler) HotCache(ctx context.Context) {
	log := logger.Get()
	cfg := configs.Get(ctx)

	go func(ctx context.Context) {
		for {
			var err error
			locs, err := h.services.MapFilter(ctx, map[string]any{})
			if err != nil {
				log.Warnf("hot cache map filter err: %s", err)
			}

			var locsJSON []byte
			if len(locs) > 0 {
				locsJSON, err = json.Marshal(locsResp{
					Status:  http.StatusOK,
					Result:  locs,
					Message: "OK",
				})
				if err != nil {
					log.Warnf("hot cache marshal err: %s", err)
				}
			}

			if len(locsJSON) > 0 {
				cacheKey := strconv.Itoa(int(crc32.Checksum([]byte("{}"), h.crcTable)))
				h.cache.Set(cacheKey, locsJSON, cfg.App.Map.CacheDuration)
			}

			timer := time.NewTimer(cfg.App.Map.CacheDuration - time.Second*10)
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
			}
		}
	}(ctx)
}

func (h *Handler) SetMetrics(ctx context.Context) {
	_ = ctx
	h.services.SetUsersMetrics()
}
