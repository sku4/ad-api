package telegram

import (
	"hash/crc32"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/sku4/ad-api/internal/app/service"
	"github.com/sku4/ad-api/pkg/telegram/bot/server"
	"github.com/sku4/ad-parser/pkg/logger"
)

const (
	lruCacheSize = 1000
)

type Handler struct {
	services service.Service
	feedback *lru.Cache[uint32, any]
	crcTable *crc32.Table
}

func NewHandler(services *service.Service) server.IHandler {
	log := logger.Get()
	cache, err := lru.New[uint32, any](lruCacheSize)
	if err != nil {
		log.Fatalf("error init lru cache: %s", err)
	}

	return &Handler{
		services: *services,
		feedback: cache,
		crcTable: crc32.MakeTable(crc32.IEEE),
	}
}
