package tarantool

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sku4/ad-api/model"
	modelAd "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/sku4/ad-parser/pkg/ad/subscription"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/datetime"
	"github.com/tarantool/go-tarantool/v2/pool"
)

var (
	usersCountQuery = fmt.Sprintf(
		`SELECT COUNT(DISTINCT "%s") as "users_count", COUNT("%s") as "sub_count" FROM "%s"`,
		modelAd.SpaceSubTgID, modelAd.SpaceSubID, modelAd.SpaceSubscription)
)

func (ad *Ad) Put(ctx context.Context, modelSub *modelAd.SubscriptionTnt) error {
	_ = ctx

	var streetID *uint
	if modelSub.StreetID != nil {
		streetUint := uint(*modelSub.StreetID)
		streetID = &streetUint
	}
	var subTnt []modelAd.SubscriptionTnt
	uniqSelect := tarantool.NewSelectRequest(modelAd.SpaceSubscription).
		Index(modelAd.IndexUniq).
		Limit(1).
		Iterator(tarantool.IterEq).
		Key([]interface{}{
			modelSub.TelegramID,
			streetID,
			modelSub.House,
			modelSub.PriceFrom,
			modelSub.PriceTo,
			modelSub.PriceM2From,
			modelSub.PriceM2To,
			modelSub.RoomsFrom,
			modelSub.RoomsTo,
			modelSub.FloorFrom,
			modelSub.FloorTo,
			modelSub.YearFrom,
			modelSub.YearTo,
			modelSub.M2MainFrom,
			modelSub.M2MainTo,
		})
	err := ad.conn.Do(uniqSelect, pool.PreferRW).GetTyped(&subTnt)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("sub put: uniq select %d %d %v %v %v %v %v %d %d %d %d %d %d %d %d",
			modelSub.TelegramID, streetID, modelSub.House, modelSub.PriceFrom, modelSub.PriceTo, modelSub.PriceM2From,
			modelSub.PriceM2To, modelSub.RoomsFrom, modelSub.RoomsTo, modelSub.FloorFrom, modelSub.FloorTo,
			modelSub.YearFrom, modelSub.YearTo, modelSub.M2MainFrom, modelSub.M2MainTo))
	}

	if len(subTnt) > 0 {
		return errors.Wrap(model.ErrSubAlreadyExists, "sub put")
	}

	// put sub
	created, err := datetime.NewDatetime(time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, "sub put: time convert to datetime")
	}
	modelSub.Created = created

	subInsert := tarantool.NewInsertRequest(modelAd.SpaceSubscription).Tuple(modelSub.ConvertToInsertTuple())
	_, err = ad.conn.Do(subInsert, pool.RW).Get()
	if err != nil {
		return errors.Wrap(err, "sub put: call")
	}

	return nil
}

func (ad *Ad) GetByTgID(ctx context.Context, tgID int64, limit int, after string) (*subscription.GetByTgIDTnt, error) {
	return ad.client.SubscriptionGetByTgID(ctx, tgID, limit, after)
}

func (ad *Ad) GetByID(ctx context.Context, id uint64) (*modelAd.SubscriptionTnt, error) {
	var subsTnt []*modelAd.SubscriptionTnt
	req := tarantool.NewSelectRequest(modelAd.SpaceSubscription).
		Index(modelAd.IndexPrimary).
		Limit(1).
		Iterator(tarantool.IterEq).
		Key(tarantool.UintKey{I: uint(id)}).
		Context(ctx)
	err := ad.conn.Do(req, pool.PreferRO).GetTyped(&subsTnt)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("get sub: primary select %d", id))
	}

	if len(subsTnt) == 0 {
		return nil, fmt.Errorf("get sub id %d: %w", id, modelAd.ErrNotFound)
	}

	return subsTnt[0], nil
}

func (ad *Ad) DeleteByID(ctx context.Context, id uint64) error {
	req := tarantool.NewDeleteRequest(modelAd.SpaceSubscription).
		Index(modelAd.IndexPrimary).
		Key(tarantool.UintKey{I: uint(id)}).
		Context(ctx)
	_, err := ad.conn.Do(req, pool.RW).Get()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("del sub: primary %d", id))
	}

	return nil
}

func (ad *Ad) GetUsersStat(ctx context.Context) (*model.UsersStat, error) {
	var usersStat []*model.UsersStat
	req := tarantool.NewExecuteRequest(usersCountQuery).
		Context(ctx)
	err := ad.conn.Do(req, pool.PreferRO).GetTyped(&usersStat)
	if err != nil {
		return nil, errors.Wrap(err, "users stat")
	}

	if len(usersStat) == 0 {
		return nil, errors.Wrap(model.ErrResultNotFound, "users stat")
	}

	return usersStat[0], nil
}
