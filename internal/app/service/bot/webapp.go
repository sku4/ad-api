package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	dec "github.com/shopspring/decimal"
	"github.com/sku4/ad-api/configs"
	"github.com/sku4/ad-api/internal/app"
	"github.com/sku4/ad-api/model"
	modelAd "github.com/sku4/ad-parser/pkg/ad/model"
)

const (
	tmplTgAddSub        = "add_sub.gohtml"
	addressPartsCap     = 2
	webAppAddSubVersion = 1
)

const (
	sliderFieldsCap = 6
	yearsAdd        = 2
)

type sliderField struct {
	Label string           `json:"label"`
	Code  string           `json:"code"`
	From  sliderFieldInput `json:"from"`
	To    sliderFieldInput `json:"to"`
	Range map[string][]int `json:"range"`
	WNumb sliderFieldWNumb `json:"wnumb"`
}

type sliderFieldInput struct {
	Label       string   `json:"label"`
	Placeholder string   `json:"placeholder"`
	Start       *float64 `json:"start,omitempty"`
}

type sliderFieldWNumb struct {
	Decimals int    `json:"decimals"`
	Thousand string `json:"thousand,omitempty"`
	Suffix   string `json:"suffix,omitempty"`
}

//nolint:gosec
func (b *Bot) WebAppAddSubscriptionIndex(ctx context.Context, isPrivate bool, urlQuery string) ([]byte, error) {
	cfg := configs.Get(ctx)

	features := cfg.Features
	featuresSet := make(map[string]struct{})
	for _, f := range features {
		featuresSet[f] = struct{}{}
	}

	jsApp := map[string]any{
		"hasStreet": hasFeature("street_id", featuresSet),
		"hasHouse":  hasFeature("house", featuresSet),
		"isPrivate": isPrivate,
		"url": map[string]any{
			"query":   urlQuery,
			"streets": "/streets",
			"sub_add": fmt.Sprintf("/%s/subscription/add", app.AddSub),
		},
	}
	slider := make([]sliderField, 0, sliderFieldsCap)

	if hasFeature("price", featuresSet) {
		slider = append(slider, sliderField{
			Label: "Цена",
			Code:  "price",
			From:  sliderFieldInput{Label: "от", Placeholder: "0 $"},
			To:    sliderFieldInput{Label: "до", Placeholder: "3 000 000 $"},
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
		slider = append(slider, sliderField{
			Label: "Цена за м²",
			Code:  "price_m2",
			From:  sliderFieldInput{Label: "от", Placeholder: "0 $"},
			To:    sliderFieldInput{Label: "до", Placeholder: "10 000 $"},
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
			From:  sliderFieldInput{Label: "от", Placeholder: "0 м²"},
			To:    sliderFieldInput{Label: "до", Placeholder: "1 000 м²"},
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
		slider = append(slider, sliderField{
			Label: "Количество комнат",
			Code:  "rooms",
			From:  sliderFieldInput{Label: "от", Placeholder: "0"},
			To:    sliderFieldInput{Label: "до", Placeholder: "20"},
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
		slider = append(slider, sliderField{
			Label: "Этаж",
			Code:  "floor",
			From:  sliderFieldInput{Label: "с", Placeholder: "0"},
			To:    sliderFieldInput{Label: "по", Placeholder: "40"},
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
		year := time.Now().Year() + yearsAdd
		slider = append(slider, sliderField{
			Label: "Год постройки",
			Code:  "year",
			From:  sliderFieldInput{Label: "с", Placeholder: "1900"},
			To:    sliderFieldInput{Label: "по", Placeholder: strconv.Itoa(year)},
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
		return nil, errors.Wrap(err, "web app add sub marshal")
	}

	mess := new(bytes.Buffer)
	if err = b.tmpl.ExecuteTemplate(mess, tmplTgAddSub, map[string]any{
		"hasStreet": hasFeature("street_id", featuresSet),
		"hasHouse":  hasFeature("house", featuresSet),
		"isPrivate": isPrivate,
		"jsApp":     template.JS(jsAppJSON),
		"Slider":    slider,
		"Version":   webAppAddSubVersion,
	}); err != nil {
		return nil, errors.Wrap(err, "web app add sub")
	}

	return mess.Bytes(), nil
}

func (b *Bot) WebAppStreets(ctx context.Context) ([]*model.Street, error) {
	cfg := configs.Get(ctx)

	b.mu.RLock()
	cacheTime := b.streetCacheTime
	streets := make([]*model.Street, 0)
	if b.streetsCache != nil {
		streets = b.streetsCache
	}
	b.mu.RUnlock()

	if cacheTime.Add(cfg.App.Street.CacheDuration).Unix() < time.Now().Unix() {
		streetsExt, err := b.repos.Street.GetStreets(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "streets")
		}

		if streetsExt == nil {
			return streets, nil
		}

		for _, streetTnt := range streetsExt.Streets {
			addressParts := make([]string, 0, addressPartsCap)
			addressParts = append(addressParts, streetTnt.Name)
			streetType := streetsExt.Types[streetTnt.Type]
			if streetType.Short != "" {
				addressParts = append(addressParts, streetType.Short+".")
			}

			streets = append(streets, &model.Street{
				ID:   streetTnt.ID,
				Name: strings.Join(addressParts, " "),
			})
		}

		sort.Slice(streets, func(i, j int) bool { return streets[i].Name < streets[j].Name })

		b.mu.Lock()
		b.streetsCache = streets
		b.streetCacheTime = time.Now()
		b.mu.Unlock()
	}

	return streets, nil
}

func (b *Bot) WebAppAddSubscription(ctx context.Context, modelSub *modelAd.SubscriptionTnt) error {
	modelSub.ID = 0
	if modelSub.StreetID != nil && *modelSub.StreetID == 0 {
		modelSub.StreetID = nil
	}
	if modelSub.House != nil && *modelSub.House == "" {
		modelSub.House = nil
	}
	decZero := dec.New(0, 0)
	if modelSub.PriceFrom != nil && modelSub.PriceFrom.Cmp(decZero) == 0 {
		modelSub.PriceFrom = nil
	}
	if modelSub.PriceTo != nil && modelSub.PriceTo.Cmp(decZero) == 0 {
		modelSub.PriceTo = nil
	}
	if modelSub.PriceM2From != nil && modelSub.PriceM2From.Cmp(decZero) == 0 {
		modelSub.PriceM2From = nil
	}
	if modelSub.PriceM2To != nil && modelSub.PriceM2To.Cmp(decZero) == 0 {
		modelSub.PriceM2To = nil
	}
	if modelSub.RoomsFrom != nil && *modelSub.RoomsFrom == 0 {
		modelSub.RoomsFrom = nil
	}
	if modelSub.RoomsTo != nil && *modelSub.RoomsTo == 0 {
		modelSub.RoomsTo = nil
	}
	if modelSub.FloorFrom != nil && *modelSub.FloorFrom == 0 {
		modelSub.FloorFrom = nil
	}
	if modelSub.FloorTo != nil && *modelSub.FloorTo == 0 {
		modelSub.FloorTo = nil
	}
	if modelSub.YearFrom != nil && *modelSub.YearFrom == 0 {
		modelSub.YearFrom = nil
	}
	if modelSub.YearTo != nil && *modelSub.YearTo == 0 {
		modelSub.YearTo = nil
	}
	if modelSub.M2MainFrom != nil && *modelSub.M2MainFrom == 0 {
		modelSub.M2MainFrom = nil
	}
	if modelSub.M2MainTo != nil && *modelSub.M2MainTo == 0 {
		modelSub.M2MainTo = nil
	}

	err := b.repos.Subscription.Put(ctx, modelSub)
	if err != nil {
		return errors.Wrap(err, "sub add")
	}

	b.SetUsersMetrics()

	return nil
}

func hasFeature(feature string, features map[string]struct{}) bool {
	_, ok := features[feature]
	return ok
}
