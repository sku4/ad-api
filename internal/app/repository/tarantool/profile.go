package tarantool

import (
	"context"
)

func (ad *Ad) ProfileGetByID(ctx context.Context, profileID uint16) string {
	return ad.client.ProfileGetByID(ctx, profileID)
}

func (ad *Ad) ProfileGetByCode(ctx context.Context, code string) uint16 {
	return ad.client.ProfileGetByCode(ctx, code)
}
