package tarantool

import (
	"context"
	"fmt"

	"github.com/sku4/ad-api/model"
	modelAd "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/sku4/ad-parser/pkg/ad/street"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

func (ad *Ad) GetStreet(ctx context.Context, id uint64) (*street.Ext, error) {
	return ad.client.StreetGet(ctx, id)
}

func (ad *Ad) GetStreets(ctx context.Context) (*model.StreetsExt, error) {
	var streetsTnt []*modelAd.StreetTnt
	req := tarantool.NewSelectRequest(modelAd.SpaceStreet).
		Index(modelAd.IndexType).
		Iterator(tarantool.IterGt).
		Key(tarantool.UintKey{I: uint(0)}).
		Context(ctx)
	err := ad.conn.Do(req, pool.PreferRO).GetTyped(&streetsTnt)
	if err != nil {
		return nil, fmt.Errorf("get streets: type select %w", err)
	}

	if len(streetsTnt) == 0 {
		return nil, fmt.Errorf("get streets: %w", modelAd.ErrNotFound)
	}

	types, err := ad.client.StreetGetTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("get streets: types %w", err)
	}

	return &model.StreetsExt{
		Streets: streetsTnt,
		Types:   types,
	}, nil
}
