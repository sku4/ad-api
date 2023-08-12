package repository

import (
	"context"

	tnt "github.com/sku4/ad-api/internal/app/repository/tarantool"
	"github.com/sku4/ad-api/model"
	modelAd "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/sku4/ad-parser/pkg/ad/street"
	"github.com/sku4/ad-parser/pkg/ad/subscription"
	"github.com/tarantool/go-tarantool/v2/pool"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

type Subscription interface {
	Put(ctx context.Context, modelSub *modelAd.SubscriptionTnt) error
	GetByTgID(ctx context.Context, tgID int64, limit int, after string) (*subscription.GetByTgIDTnt, error)
	GetByID(ctx context.Context, id uint64) (*modelAd.SubscriptionTnt, error)
	DeleteByID(ctx context.Context, id uint64) error
	GetUsersStat(ctx context.Context) (*model.UsersStat, error)
}

type Street interface {
	GetStreet(ctx context.Context, id uint64) (*street.Ext, error)
	GetStreets(ctx context.Context) (*model.StreetsExt, error)
}

type Ad interface {
	Ads(ctx context.Context, ids []uint64) ([]*modelAd.AdTnt, error)
	Filter(ctx context.Context, fields map[string]any) ([]*modelAd.AdLocationTnt, error)
}

type Profile interface {
	ProfileGetByID(ctx context.Context, profileID uint16) string
	ProfileGetByCode(ctx context.Context, code string) uint16
}

type Repository struct {
	Ad
	Street
	Subscription
	Profile
}

func NewRepository(conn pool.Pooler) *Repository {
	ad := tnt.NewAd(conn)
	return &Repository{
		Subscription: ad,
		Street:       ad,
		Ad:           ad,
		Profile:      ad,
	}
}
