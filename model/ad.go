package model

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/tarantool/go-tarantool/v2/decimal"
)

type AdLocation struct {
	ID      *uint64  `json:"i,omitempty"`
	IDs     []uint64 `json:"s,omitempty"`
	LocLat  float64  `json:"t"`
	LocLong float64  `json:"g"`
}

type AdFilterFields struct {
	StreetID    *uint64          `json:"street_id,omitempty"`
	House       *string          `json:"house,omitempty"`
	PriceFrom   *decimal.Decimal `json:"price_from,omitempty"`
	PriceTo     *decimal.Decimal `json:"price_to,omitempty"`
	PriceM2From *decimal.Decimal `json:"price_m2_from,omitempty"`
	PriceM2To   *decimal.Decimal `json:"price_m2_to,omitempty"`
	RoomsFrom   *uint8           `json:"rooms_from,omitempty"`
	RoomsTo     *uint8           `json:"rooms_to,omitempty"`
	FloorFrom   *uint8           `json:"floor_from,omitempty"`
	FloorTo     *uint8           `json:"floor_to,omitempty"`
	FloorsFrom  *uint8           `json:"floors_from,omitempty"`
	FloorsTo    *uint8           `json:"floors_to,omitempty"`
	YearFrom    *uint16          `json:"year_from,omitempty"`
	YearTo      *uint16          `json:"year_to,omitempty"`
	M2MainFrom  *float64         `json:"m2_main_from,omitempty"`
	M2MainTo    *float64         `json:"m2_main_to,omitempty"`
	Profiles    []uint16         `json:"profiles,omitempty"`
}

func (a AdFilterFields) ConvertToTuple() (map[string]any, error) {
	adJSON, errM := json.Marshal(a)
	if errM != nil {
		return nil, errors.Wrap(errM, "convert to tuple")
	}

	var adTuple map[string]any
	errM = json.Unmarshal(adJSON, &adTuple)
	if errM != nil {
		return nil, errors.Wrap(errM, "convert to tuple")
	}

	if a.PriceFrom != nil {
		adTuple["price_from"] = a.PriceFrom
	}
	if a.PriceTo != nil {
		adTuple["price_to"] = a.PriceTo
	}
	if a.PriceM2From != nil {
		adTuple["price_m2_from"] = a.PriceM2From
	}
	if a.PriceM2To != nil {
		adTuple["price_m2_to"] = a.PriceM2To
	}

	return adTuple, nil
}

type AdsFilter struct {
	Groups []*AdFilterGroup `json:"groups,omitempty"`
}

type AdFilterGroup struct {
	IDs []uint64 `json:"ids,omitempty"`
}

type Ad struct {
	ID        uint64   `json:"id"`
	Created   *int64   `json:"c_time,omitempty"`
	URLs      []*AdURL `json:"urls"`
	Address   *string  `json:"address,omitempty"`
	Price     *float64 `json:"price,omitempty"`
	PriceM2   *float64 `json:"price_m2,omitempty"`
	Rooms     *uint8   `json:"rooms,omitempty"`
	Floor     *uint8   `json:"floor,omitempty"`
	Floors    *uint8   `json:"floors,omitempty"`
	Year      *uint16  `json:"year,omitempty"`
	Photo     *string  `json:"photo,omitempty"`
	M2Main    *float64 `json:"m2_main,omitempty"`
	M2Living  *float64 `json:"m2_living,omitempty"`
	M2Kitchen *float64 `json:"m2_kitchen,omitempty"`
	Bathroom  *string  `json:"bathroom,omitempty"`
}

type AdURL struct {
	Profile string `json:"p"`
	URL     string `json:"u"`
}
