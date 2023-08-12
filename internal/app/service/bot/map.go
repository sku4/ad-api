package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sku4/ad-api/configs"
	"github.com/sku4/ad-api/model"
	modelAd "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/sku4/ad-parser/pkg/logger"
)

const (
	tmplTgMap  = "map.gohtml"
	mapVersion = 2
)

//nolint:gosec,dupl
func (b *Bot) MapIndex(ctx context.Context, urlQuery string, filterFields *model.AdFilterFields) ([]byte, error) {
	cfg := configs.Get(ctx)

	features := cfg.Features
	featuresSet := make(map[string]struct{})
	for _, f := range features {
		featuresSet[f] = struct{}{}
	}

	valHouse := ""
	if filterFields.House != nil {
		valHouse = *filterFields.House
	}

	jsApp := map[string]any{
		"hasStreet": hasFeature("street_id", featuresSet),
		"hasHouse":  hasFeature("house", featuresSet),
		"valStreet": filterFields.StreetID,
		"valHouse":  valHouse,
		"wnumb": sliderFieldWNumb{
			Decimals: 0,
			Thousand: " ",
			Suffix:   " $",
		},
		"url": map[string]any{
			"query":     urlQuery,
			"ads":       "/ads",
			"locations": "/locations",
			"streets":   "/streets",
		},
	}
	slider := make([]sliderField, 0, sliderFieldsCap)

	if hasFeature("price", featuresSet) {
		var priceFrom, priceTo *float64
		if filterFields.PriceFrom != nil {
			p, _ := filterFields.PriceFrom.Float64()
			priceFrom = &p
		}
		if filterFields.PriceTo != nil {
			p, _ := filterFields.PriceTo.Float64()
			priceTo = &p
		}
		slider = append(slider, sliderField{
			Label: "Цена",
			Code:  "price",
			From:  sliderFieldInput{"от", "0 $", priceFrom},
			To:    sliderFieldInput{"до", "3 000 000 $", priceTo},
			Range: map[string][]int{
				"min": {0, 100},
				"5%":  {10000, 500},
				"70%": {100000, 1000},
				"90%": {200000, 5000},
				"max": {3000000, 10000},
			},
			WNumb: sliderFieldWNumb{
				Decimals: 0,
				Thousand: " ",
				Suffix:   " $",
			},
		})
	}

	if hasFeature("price_m2", featuresSet) {
		var priceM2From, priceM2To *float64
		if filterFields.PriceM2From != nil {
			p, _ := filterFields.PriceM2From.Float64()
			priceM2From = &p
		}
		if filterFields.PriceM2To != nil {
			p, _ := filterFields.PriceM2To.Float64()
			priceM2To = &p
		}
		slider = append(slider, sliderField{
			Label: "Цена за м²",
			Code:  "price_m2",
			From:  sliderFieldInput{"от", "0 $", priceM2From},
			To:    sliderFieldInput{"до", "10 000 $", priceM2To},
			Range: map[string][]int{
				"min": {0, 10},
				"5%":  {500, 10},
				"70%": {3000, 10},
				"90%": {5000, 100},
				"max": {10000, 100},
			},
			WNumb: sliderFieldWNumb{
				Decimals: 0,
				Thousand: " ",
				Suffix:   " $",
			},
		})
	}

	if hasFeature("m2_main", featuresSet) {
		slider = append(slider, sliderField{
			Label: "Общая площадь",
			Code:  "m2_main",
			From:  sliderFieldInput{"от", "0 м²", filterFields.M2MainFrom},
			To:    sliderFieldInput{"до", "1 000 м²", filterFields.M2MainTo},
			Range: map[string][]int{
				"min": {0, 1},
				"80%": {200, 5},
				"90%": {500, 10},
				"max": {1000, 10},
			},
			WNumb: sliderFieldWNumb{
				Decimals: 0,
				Thousand: " ",
				Suffix:   " м²",
			},
		})
	}

	if hasFeature("rooms", featuresSet) {
		var roomsFrom, roomsTo *float64
		if filterFields.RoomsFrom != nil {
			p := float64(*filterFields.RoomsFrom)
			roomsFrom = &p
		}
		if filterFields.RoomsTo != nil {
			p := float64(*filterFields.RoomsTo)
			roomsTo = &p
		}
		slider = append(slider, sliderField{
			Label: "Количество комнат",
			Code:  "rooms",
			From:  sliderFieldInput{"от", "0", roomsFrom},
			To:    sliderFieldInput{"до", "20", roomsTo},
			Range: map[string][]int{
				"min": {0, 1},
				"70%": {10, 1},
				"max": {20, 1},
			},
			WNumb: sliderFieldWNumb{
				Decimals: 0,
			},
		})
	}

	if hasFeature("floor", featuresSet) {
		var floorFrom, floorTo *float64
		if filterFields.FloorFrom != nil {
			p := float64(*filterFields.FloorFrom)
			floorFrom = &p
		}
		if filterFields.FloorTo != nil {
			p := float64(*filterFields.FloorTo)
			floorTo = &p
		}
		slider = append(slider, sliderField{
			Label: "Этаж",
			Code:  "floor",
			From:  sliderFieldInput{"с", "0", floorFrom},
			To:    sliderFieldInput{"по", "40", floorTo},
			Range: map[string][]int{
				"min": {0, 1},
				"70%": {25, 1},
				"max": {40, 1},
			},
			WNumb: sliderFieldWNumb{
				Decimals: 0,
			},
		})
	}

	if hasFeature("floors", featuresSet) {
		var floorsFrom, floorsTo *float64
		if filterFields.FloorsFrom != nil {
			p := float64(*filterFields.FloorsFrom)
			floorsFrom = &p
		}
		if filterFields.FloorsTo != nil {
			p := float64(*filterFields.FloorsTo)
			floorsTo = &p
		}
		slider = append(slider, sliderField{
			Label: "Этажность",
			Code:  "floors",
			From:  sliderFieldInput{"с", "0", floorsFrom},
			To:    sliderFieldInput{"по", "40", floorsTo},
			Range: map[string][]int{
				"min": {0, 1},
				"70%": {25, 1},
				"max": {40, 1},
			},
			WNumb: sliderFieldWNumb{
				Decimals: 0,
			},
		})
	}

	if hasFeature("year", featuresSet) {
		var yearFrom, yearTo *float64
		if filterFields.YearFrom != nil {
			p := float64(*filterFields.YearFrom)
			yearFrom = &p
		}
		if filterFields.YearTo != nil {
			p := float64(*filterFields.YearTo)
			yearTo = &p
		}
		year := time.Now().Year() + yearsAdd
		slider = append(slider, sliderField{
			Label: "Год постройки",
			Code:  "year",
			From:  sliderFieldInput{"с", "1900", yearFrom},
			To:    sliderFieldInput{"по", strconv.Itoa(year), yearTo},
			Range: map[string][]int{
				"min": {1900, 1},
				"10%": {1980, 1},
				"30%": {2000, 1},
				"max": {year, 1},
			},
			WNumb: sliderFieldWNumb{
				Decimals: 0,
			},
		})
	}

	jsApp["slider"] = slider

	jsAppJSON, err := json.Marshal(jsApp)
	if err != nil {
		return nil, errors.Wrap(err, "map index marshal")
	}

	mess := new(bytes.Buffer)
	if err = b.tmpl.ExecuteTemplate(mess, tmplTgMap, map[string]any{
		"hasStreet": hasFeature("street_id", featuresSet),
		"hasHouse":  hasFeature("house", featuresSet),
		"valStreet": filterFields.StreetID,
		"valHouse":  valHouse,
		"jsApp":     template.JS(jsAppJSON),
		"Slider":    slider,
		"Version":   mapVersion,
	}); err != nil {
		return nil, errors.Wrap(err, "map index")
	}

	return mess.Bytes(), nil
}

func (b *Bot) MapFilter(ctx context.Context, fields map[string]any) ([]*model.AdLocation, error) {
	locs := make([]*model.AdLocation, 0)

	locsFilter, err := b.repos.Ad.Filter(ctx, fields)
	if err != nil {
		return nil, errors.Wrap(err, "map filter")
	}

	if locsFilter == nil {
		return locs, nil
	}

	for _, loc := range locsFilter {
		locs = append(locs, &model.AdLocation{
			ID:      loc.ID,
			IDs:     loc.IDs,
			LocLat:  loc.LocLat,
			LocLong: loc.LocLong,
		})
	}

	return locs, nil
}

//nolint:gocyclo
func (b *Bot) MapAds(ctx context.Context, groups []*model.AdFilterGroup) ([]*model.Ad, error) {
	log := logger.Get()
	ads := make([]*model.Ad, 0)
	ids := make([]uint64, 0, len(groups))

	for _, group := range groups {
		ids = append(ids, group.IDs...)
	}

	adsTnt, err := b.repos.Ad.Ads(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, "map ads")
	}

	if adsTnt == nil {
		return ads, nil
	}

	adsTntMap := make(map[uint64]*modelAd.AdTnt, len(adsTnt))
	for _, ad := range adsTnt {
		adsTntMap[ad.ID] = ad
	}

	for _, group := range groups {
		if len(group.IDs) == 0 {
			continue
		}

		var ad *modelAd.AdTnt
		var year *uint16
		var floors *uint8
		var bathroom *string
		var photo *string
		urls := make([]*model.AdURL, 0, 1)
		for _, id := range group.IDs {
			if adTnt, ok := adsTntMap[id]; ok {
				urls = append(urls, &model.AdURL{
					Profile: b.repos.Profile.ProfileGetByID(ctx, adTnt.Profile),
					URL:     adTnt.URL,
				})
				if adTnt.Year != nil && *adTnt.Year > 0 {
					y := *adTnt.Year
					year = &y
				}
				if adTnt.Floors != nil && *adTnt.Floors > 0 {
					f := *adTnt.Floors
					floors = &f
				}
				if adTnt.Bathroom != nil && *adTnt.Bathroom != "" {
					bp := *adTnt.Bathroom
					bathroom = &bp
				}
				if adTnt.Photos != nil && len(adTnt.Photos) > 0 {
					p := adTnt.Photos[0]
					photo = &p
				}
				ad = adTnt
			}
		}

		if ad == nil {
			continue
		}

		ad.Year = year
		ad.Floors = floors
		ad.Bathroom = bathroom

		var created *int64
		if ad.Created != nil {
			crUnix := ad.Created.ToTime().Unix()
			created = &crUnix
		}

		var price, priceM2 *float64
		if ad.Price != nil {
			priceFloat, _ := ad.Price.Float64()
			price = &priceFloat
		}
		if ad.PriceM2 != nil {
			priceM2Float, _ := ad.PriceM2.Float64()
			priceM2 = &priceM2Float
		}

		addressParts := make([]string, 0, countPartsThree)
		if ad.StreetID != nil && *ad.StreetID > 0 {
			streetExt, errStreet := b.repos.Street.GetStreet(ctx, *ad.StreetID)
			if errStreet != nil {
				log.Warnf("map ads: %s", errStreet)
			}
			if streetExt != nil {
				addressParts = append(addressParts, streetExt.Street.Name)
				streetType := ""
				if streetExt.Type.Short != "" {
					streetType = streetExt.Type.Short + "."
				}
				if streetType != "" {
					addressParts = append(addressParts, streetType)
				}
			}
		}

		if ad.House != nil && *ad.House != "" {
			if len(addressParts) == 0 {
				addressParts = append(addressParts, fmt.Sprintf("дом %s", *ad.House))
			} else {
				addressParts = append(addressParts, *ad.House)
			}
		}

		var address *string
		if len(addressParts) > 0 {
			a := strings.Join(addressParts, " ")
			address = &a
		}

		ads = append(ads, &model.Ad{
			ID:        ad.ID,
			Created:   created,
			URLs:      urls,
			Address:   address,
			Price:     price,
			PriceM2:   priceM2,
			Rooms:     ad.Rooms,
			Floor:     ad.Floor,
			Floors:    ad.Floors,
			Year:      ad.Year,
			Photo:     photo,
			M2Main:    ad.M2Main,
			M2Living:  ad.M2Living,
			M2Kitchen: ad.M2Kitchen,
			Bathroom:  ad.Bathroom,
		})
	}

	return ads, nil
}
