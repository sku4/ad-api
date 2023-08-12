package bot

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	dec "github.com/shopspring/decimal"
	"github.com/sku4/ad-api/configs"
	"github.com/sku4/ad-api/internal/app"
	"github.com/sku4/ad-api/pkg/telegram/bot/client"
	in "github.com/sku4/ad-api/pkg/telegram/inline"
	modelAd "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/sku4/ad-parser/pkg/logger"
	"github.com/tarantool/go-tarantool/v2/decimal"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	tmplTgListEmpty   = "list_empty.gohtml"
	tmplTgSubNotFound = "sub_not_found.gohtml"
	countSubsInline   = 10
	countSubsMax      = 500
	countPBInline     = 2
	countPartsTwo     = 2
	countPartsThree   = 3
	roundNumber       = 100
)

func (b *Bot) ManageSubscriptionsPage(ctx context.Context, chatID int64, messageID int,
	page *in.PaginationInline) error {
	subsTgIDTnt, err := b.repos.Subscription.GetByTgID(ctx, chatID, countSubsMax, "")
	if err != nil {
		return errors.Wrap(err, "manage subscriptions page: get")
	}

	if page.Current.PageID == 1 && len(subsTgIDTnt.Subscriptions) == 0 {
		mess := new(bytes.Buffer)
		if err = b.tmpl.ExecuteTemplate(mess, tmplTgListEmpty, map[string]any{
			"AddSub": app.AddSub,
		}); err != nil {
			return errors.Wrap(err, "manage subscriptions page")
		}
		b.sendAnswer(ctx, chatID, messageID, mess.String())

		return nil
	}

	inlineKeyboardRows := make([]*client.KeyboardRow, 0, countSubsInline+countPBInline)
	countSubs := countSubsInline * (page.Current.PageID - 1)
	severalPages := subsTgIDTnt.All > countSubsInline
	cntSubsShow := 0
	for i, sub := range subsTgIDTnt.Subscriptions {
		if i < countSubs || i >= countSubs+countSubsInline {
			continue
		}
		inline := in.NewInlineExt(app.ManageSubsEntity, app.CommandOpen, in.NewManageSubInlineExt(sub.ID, page))
		inlineKeyboardRow := client.NewKeyboardRow()
		title := strings.Join(b.buildTitle(ctx, sub, false), ", ")
		if severalPages {
			title = fmt.Sprintf("%d. %s", i+1, title)
		}
		inlineKeyboardRow.Add(title, inline.Serialize())
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)
		cntSubsShow++
	}

	firstPage := page.Current.PageID == 1
	lastPage := !(subsTgIDTnt.All > int64(countSubs+cntSubsShow))
	inlineKeyboardRow := client.NewKeyboardRow()

	if !firstPage {
		current := in.NewPageExt(page.Current.PageID - 1)
		inline := in.NewInlineExt(app.ManageSubsEntity, app.CommandPage,
			in.NewPaginationInlineExt(current))
		inlineKeyboardRow.Add("<< Назад", inline.Serialize())
	}

	if !lastPage {
		current := in.NewPageExt(page.Current.PageID + 1)
		inline := in.NewInlineExt(app.ManageSubsEntity, app.CommandPage,
			in.NewPaginationInlineExt(current))
		inlineKeyboardRow.Add("Далее >>", inline.Serialize())
	}

	if severalPages {
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)
	}

	if messageID > 0 {
		err = b.client.SendCallbackQuery(ctx, inlineKeyboardRows, "Управление подписками", messageID, chatID)
		if err != nil {
			return errors.Wrap(err, "manage subscriptions page: send callback query")
		}
	} else {
		err = b.client.SendInlineKeyboard(ctx, inlineKeyboardRows, "Управление подписками", chatID)
		if err != nil {
			return errors.Wrap(err, "manage subscriptions page: send inline keyboard")
		}
	}

	return nil
}

func (b *Bot) ManageSubscriptionsOpen(ctx context.Context, chatID int64, messageID int,
	sub *in.ManageSubInline) error {
	subTnt, err := b.repos.Subscription.GetByID(ctx, sub.SubscriptionID)
	if err != nil {
		if errors.Is(err, modelAd.ErrNotFound) {
			mess := new(bytes.Buffer)
			if err = b.tmpl.ExecuteTemplate(mess, tmplTgSubNotFound, map[string]any{
				"ListSubs": app.ManageSubsEntity,
			}); err != nil {
				return errors.Wrap(err, "manage subscriptions open")
			}
			b.sendAnswer(ctx, chatID, messageID, mess.String())

			return nil
		}

		return errors.Wrap(err, "manage subscriptions open: get")
	}

	showOnMapURL, err := buildShowOnMapURL(ctx, subTnt)
	if err != nil {
		return errors.Wrap(err, "manage subscriptions open: build show on map url")
	}

	title := strings.Join(b.buildTitle(ctx, subTnt, true), "\n")
	back := in.NewInlineExt(app.ManageSubsEntity, app.CommandPage, in.NewPaginationInlineExt(sub.Back.Current))
	del := in.NewInlineExt(app.ManageSubsEntity, app.CommandDelete, in.NewManageSubInlineExt(sub.SubscriptionID, sub.Back))
	inlineKeyboardRows := make([]*client.KeyboardRow, 0, countPBInline)
	inlineKeyboardRow := client.NewKeyboardRow()
	inlineKeyboardRow.AddURL("Показать на карте  🌎", showOnMapURL)
	inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)
	inlineKeyboardRow = client.NewKeyboardRow()
	inlineKeyboardRow.Add("<< Назад", back.Serialize())
	inlineKeyboardRow.Add("Удалить  ❌", del.Serialize())
	inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)

	err = b.client.SendCallbackQuery(ctx, inlineKeyboardRows, fmt.Sprintf("Параметры подписки\n\n%s", title),
		messageID, chatID)
	if err != nil {
		return errors.Wrap(err, "manage subscriptions open: send callback query")
	}

	return nil
}

func (b *Bot) ManageSubscriptionsDelete(ctx context.Context, chatID int64, messageID int,
	queryID string, sub *in.ManageSubInline) error {
	log := logger.Get()

	subTnt, err := b.repos.Subscription.GetByID(ctx, sub.SubscriptionID)
	if err != nil {
		errSend := b.client.AnswerCallbackQuery(ctx, queryID, "Подписка не найдена")
		if errSend != nil {
			log.Warnf("manage subscriptions delete: send message error: %s", errSend)
		}

		return errors.Wrap(err, "manage subscriptions delete: get")
	}

	err = b.repos.Subscription.DeleteByID(ctx, sub.SubscriptionID)
	if err != nil {
		errSend := b.client.AnswerCallbackQuery(ctx, queryID, "Ошибка удаления")
		if errSend != nil {
			log.Warnf("manage subscriptions delete: send message error: %s", errSend)
		}

		return errors.Wrap(err, "manage subscriptions delete")
	}

	title := strings.Join(b.buildTitle(ctx, subTnt, false), ", ")
	err = b.client.SendMessage(ctx, fmt.Sprintf("Подписка удалена: %s", title), chatID, retryCount)
	if err != nil {
		log.Warnf("manage subscriptions delete: send message error: %s", err)
	}

	b.SetUsersMetrics()

	return b.ManageSubscriptionsPage(ctx, chatID, messageID, in.NewPaginationInline())
}

func (b *Bot) sendAnswer(ctx context.Context, chatID int64, messageID int, message string) {
	log := logger.Get()

	if messageID > 0 {
		err := b.client.SendCallbackQuery(ctx, nil, message, messageID, chatID)
		if err != nil {
			log.Warnf("manage subscriptions: send callback query error: %s", err)
		}
	} else {
		err := b.client.SendMessage(ctx, message, chatID, retryCount)
		if err != nil {
			log.Warnf("manage subscriptions: send message error: %s", err)
		}
	}
}

func (b *Bot) buildTitle(ctx context.Context, sub *modelAd.SubscriptionTnt, ext bool) []string {
	log := logger.Get()
	title := make([]string, 0)

	if sub == nil {
		return title
	}

	address := make([]string, 0, countPartsThree)
	if sub.StreetID != nil && *sub.StreetID > 0 {
		streetExt, err := b.repos.Street.GetStreet(ctx, *sub.StreetID)
		if err != nil {
			log.Warnf("build title: %s", err)
		}
		if streetExt != nil {
			address = append(address, streetExt.Street.Name)
			streetType := ""
			if streetExt.Type.Short != "" {
				streetType = streetExt.Type.Short + "."
			}
			if streetType != "" {
				address = append(address, streetType)
			}
		}
	}

	if sub.House != nil && *sub.House != "" {
		if len(address) == 0 {
			address = append(address, fmt.Sprintf("дом %s", *sub.House))
		} else {
			address = append(address, *sub.House)
		}
	}

	if len(address) > 0 {
		if ext {
			title = append(title, fmt.Sprintf("Адрес: %s", strings.Join(address, " ")))
		} else {
			title = append(title, strings.Join(address, " "))
		}
	}

	price := buildPrice(sub.PriceFrom, sub.PriceTo)
	if price != "" {
		if ext {
			title = append(title, fmt.Sprintf("Цена: %s", price))
		} else {
			title = append(title, price)
		}
	}

	priceM2 := buildPrice(sub.PriceM2From, sub.PriceM2To)
	if priceM2 != "" {
		if ext {
			title = append(title, fmt.Sprintf("Цена м²: %s", priceM2))
		} else {
			title = append(title, fmt.Sprintf("%s / м²", priceM2))
		}
	}

	m2Main := buildFloat(sub.M2MainFrom, sub.M2MainTo, "от %s", "до %s")
	if m2Main != "" {
		if ext {
			title = append(title, fmt.Sprintf("Общая площадь: %s м²", m2Main))
		} else {
			title = append(title, fmt.Sprintf("%s м²", m2Main))
		}
	}

	room := buildUint(sub.RoomsFrom, sub.RoomsTo, "от %d", "до %d")
	if room != "" {
		if ext {
			title = append(title, fmt.Sprintf("Комнат: %s", room))
		} else {
			title = append(title, fmt.Sprintf("%s комн", room))
		}
	}

	floor := buildUint(sub.FloorFrom, sub.FloorTo, "с %d", "по %d")
	if floor != "" {
		if ext {
			title = append(title, fmt.Sprintf("Этаж: %s", floor))
		} else {
			title = append(title, fmt.Sprintf("%s этаж", floor))
		}
	}

	year := buildUint(sub.YearFrom, sub.YearTo, "с %d", "по %d")
	if year != "" {
		if ext {
			title = append(title, fmt.Sprintf("Год постройки: %s", year))
		} else {
			title = append(title, fmt.Sprintf("%s г.п.", year))
		}
	}

	return title
}

func buildUint[B uint8 | uint16](from, to *B, fromPattern, toPattern string) string {
	numWords := make([]string, 0, countPartsTwo)
	num := make([]string, 0, countPartsTwo)
	if from != nil && *from > 0 {
		numWords = append(numWords, fmt.Sprintf(fromPattern, *from))
		num = append(num, fmt.Sprintf("%d", *from))
	}
	if to != nil && *to > 0 {
		numWords = append(numWords, fmt.Sprintf(toPattern, *to))
		num = append(num, fmt.Sprintf("%d", *to))
	}

	if len(num) == countPartsTwo {
		return strings.Join(num, " - ")
	} else if len(numWords) > 0 {
		return strings.Join(numWords, " ")
	}

	return ""
}

func buildPrice(from, to *decimal.Decimal) string {
	printFr := message.NewPrinter(language.French)
	priceWords := make([]string, 0, countPartsTwo)
	prices := make([]string, 0, countPartsTwo)
	priceFromFloat := float64(0)
	if from != nil {
		priceFromFloat, _ = from.Float64()
	}
	priceToFloat := float64(0)
	if to != nil {
		priceToFloat, _ = to.Float64()
	}
	if priceFromFloat > 0 {
		priceWords = append(priceWords, printFr.Sprintf("от %.0f", priceFromFloat))
		prices = append(prices, printFr.Sprintf("%.0f", priceFromFloat))
	}
	if priceToFloat > 0 {
		priceWords = append(priceWords, printFr.Sprintf("до %.0f", priceToFloat))
		prices = append(prices, printFr.Sprintf("%.0f", priceToFloat))
	}

	if len(prices) == countPartsTwo {
		return fmt.Sprintf("%s$", strings.Join(prices, " - "))
	} else if len(priceWords) > 0 {
		return fmt.Sprintf("%s$", strings.Join(priceWords, " "))
	}

	return ""
}

func buildFloat(from, to *float64, fromPattern, toPattern string) string {
	numWords := make([]string, 0, countPartsTwo)
	num := make([]string, 0, countPartsTwo)
	if from != nil && *from > 0 {
		numWords = append(numWords, getFloatM2(*from, fromPattern))
		num = append(num, getFloatM2(*from, "%s"))
	}
	if to != nil && *to > 0 {
		numWords = append(numWords, getFloatM2(*to, toPattern))
		num = append(num, getFloatM2(*to, "%s"))
	}

	if len(num) == countPartsTwo {
		return strings.Join(num, " - ")
	} else if len(numWords) > 0 {
		return strings.Join(numWords, " ")
	}

	return ""
}

func getFloatM2(f float64, pattern string) string {
	if int(f*roundNumber)%roundNumber > 0 {
		return fmt.Sprintf(pattern, fmt.Sprintf("%.1f", f))
	}
	return fmt.Sprintf(pattern, fmt.Sprintf("%.0f", f))
}

func buildShowOnMapURL(ctx context.Context, subTnt *modelAd.SubscriptionTnt) (string, error) {
	cfg := configs.Get(ctx)

	showOnMapURL, err := url.Parse(cfg.App.HostURL)
	if err != nil {
		return "", err
	}

	q := showOnMapURL.Query()
	if subTnt.StreetID != nil && *subTnt.StreetID > 0 {
		q.Set("street_id", strconv.FormatUint(*subTnt.StreetID, 10))
	}
	if subTnt.House != nil && *subTnt.House != "" {
		q.Set("house", fmt.Sprintf("\"%s\"", *subTnt.House))
	}
	decZero := dec.New(0, 0)
	if subTnt.PriceFrom != nil && subTnt.PriceFrom.Cmp(decZero) > 0 {
		f, _ := subTnt.PriceFrom.Float64()
		q.Set("price_from", strconv.FormatFloat(f, 'f', 0, 64))
	}
	if subTnt.PriceTo != nil && subTnt.PriceTo.Cmp(decZero) > 0 {
		f, _ := subTnt.PriceTo.Float64()
		q.Set("price_to", strconv.FormatFloat(f, 'f', 0, 64))
	}
	if subTnt.PriceM2From != nil && subTnt.PriceM2From.Cmp(decZero) > 0 {
		f, _ := subTnt.PriceM2From.Float64()
		q.Set("price_m2_from", strconv.FormatFloat(f, 'f', 0, 64))
	}
	if subTnt.PriceM2To != nil && subTnt.PriceM2To.Cmp(decZero) > 0 {
		f, _ := subTnt.PriceM2To.Float64()
		q.Set("price_m2_to", strconv.FormatFloat(f, 'f', 0, 64))
	}
	if subTnt.RoomsFrom != nil && *subTnt.RoomsFrom > 0 {
		q.Set("rooms_from", strconv.Itoa(int(*subTnt.RoomsFrom)))
	}
	if subTnt.RoomsTo != nil && *subTnt.RoomsTo > 0 {
		q.Set("rooms_to", strconv.Itoa(int(*subTnt.RoomsTo)))
	}
	if subTnt.FloorFrom != nil && *subTnt.FloorFrom > 0 {
		q.Set("floor_from", strconv.Itoa(int(*subTnt.FloorFrom)))
	}
	if subTnt.FloorTo != nil && *subTnt.FloorTo > 0 {
		q.Set("floor_to", strconv.Itoa(int(*subTnt.FloorTo)))
	}
	if subTnt.YearFrom != nil && *subTnt.YearFrom > 0 {
		q.Set("year_from", strconv.Itoa(int(*subTnt.YearFrom)))
	}
	if subTnt.YearTo != nil && *subTnt.YearTo > 0 {
		q.Set("year_to", strconv.Itoa(int(*subTnt.YearTo)))
	}
	if subTnt.M2MainFrom != nil && *subTnt.M2MainFrom > 0 {
		f := *subTnt.M2MainFrom
		q.Set("m2_main_from", strconv.FormatFloat(f, 'f', 0, 64))
	}
	if subTnt.M2MainTo != nil && *subTnt.M2MainTo > 0 {
		f := *subTnt.M2MainTo
		q.Set("m2_main_to", strconv.FormatFloat(f, 'f', 0, 64))
	}

	showOnMapURL.RawQuery = q.Encode()

	return showOnMapURL.String(), nil
}
