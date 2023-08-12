package tarantool

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	client "github.com/sku4/ad-parser/pkg/ad"
	modelAd "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

type Ad struct {
	conn   pool.Pooler
	client *client.Client
}

func NewAd(conn pool.Pooler) *Ad {
	clientTnt := client.NewClient(conn)
	return &Ad{
		conn:   conn,
		client: clientTnt,
	}
}

func (ad *Ad) Filter(ctx context.Context, fields map[string]any) ([]*modelAd.AdLocationTnt, error) {
	return ad.client.AdFilter(ctx, fields)
}

func (ad *Ad) Ads(ctx context.Context, ids []uint64) ([]*modelAd.AdTnt, error) {
	ads := make([]*modelAd.AdTnt, 0, len(ids))

	for _, id := range ids {
		var adsTnt []*modelAd.AdTnt
		req := tarantool.NewSelectRequest(modelAd.SpaceAd).
			Index(modelAd.IndexPrimary).
			Limit(1).
			Iterator(tarantool.IterEq).
			Key(tarantool.UintKey{I: uint(id)}).
			Context(ctx)
		err := ad.conn.Do(req, pool.PreferRO).GetTyped(&adsTnt)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("get ads: primary select %d", id))
		}

		if len(adsTnt) == 0 {
			return nil, fmt.Errorf("get ads %d: %w", id, modelAd.ErrNotFound)
		}

		ads = append(ads, adsTnt[0])
	}

	return ads, nil
}
