package bot

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sku4/ad-api/configs"
	"github.com/sku4/ad-api/internal/app"
	"github.com/sku4/ad-api/pkg/secret"
	"github.com/sku4/ad-api/pkg/telegram/bot/client"
)

const (
	tmplTgHelp         = "help.gohtml"
	tmplTgStart        = "start.gohtml"
	tmplTgFeedback     = "feedback.gohtml"
	tmplTgFeedbackMess = "feedback_mess.gohtml"
	retryCount         = 2
	countSABInline     = 1
)

func (b *Bot) Start(ctx context.Context, chatID int64) error {
	mess := new(bytes.Buffer)
	if err := b.tmpl.ExecuteTemplate(mess, tmplTgStart, map[string]any{}); err != nil {
		return errors.Wrap(err, "start")
	}

	err := b.client.SendUploadPhoto(ctx, []string{"web/static/img/logo-start.png"},
		[]string{mess.String()}, chatID, retryCount)
	if err != nil {
		return errors.Wrap(err, "start")
	}

	return nil
}

func (b *Bot) Help(ctx context.Context, chatID int64) error {
	mess := new(bytes.Buffer)
	if err := b.tmpl.ExecuteTemplate(mess, tmplTgHelp, map[string]any{
		"AddSub": app.AddSub,
		"Map":    app.Map,
	}); err != nil {
		return errors.Wrap(err, "help")
	}

	err := b.client.SendMessage(ctx, mess.String(), chatID, retryCount)
	if err != nil {
		return errors.Wrap(err, "help")
	}

	return nil
}

func (b *Bot) AddSubscription(ctx context.Context, chatID int64, isPrivate bool) error {
	cfg := configs.Get(ctx)

	addSubURL, err := url.Parse(cfg.App.HostURL)
	if err != nil {
		return errors.Wrap(err, "subscription add: parse url")
	}
	addSubURL.Path = fmt.Sprintf("/%s", app.AddSub)

	inlineKeyboardRows := make([]*client.KeyboardRow, 0, countSABInline)
	inlineKeyboardRow := client.NewKeyboardRow()

	q := addSubURL.Query()
	q.Set(app.ParamChatID, strconv.FormatInt(chatID, 10))

	if isPrivate {
		q.Set(app.ParamIsPrivate, app.ParamValueTrue)
		q.Set(app.ParamHash, secret.Hash(cfg.Telegram.BotToken, q.Encode()))
		addSubURL.RawQuery = q.Encode()
		inlineKeyboardRow.AddWebApp("Открыть форму добавления подписки", addSubURL.String())
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)
		err = b.client.SendKeyboard(ctx, inlineKeyboardRows, "Для добавления подписки откройте форму по кнопке ниже",
			chatID, false, true)
		if err != nil {
			return errors.Wrap(err, "subscription add: send keyboard")
		}
	} else {
		q.Set(app.ParamHash, secret.Hash(cfg.Telegram.BotToken, q.Encode()))
		addSubURL.RawQuery = q.Encode()
		inlineKeyboardRow.AddURL("Открыть форму добавления подписки", addSubURL.String())
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)
		err = b.client.SendInlineKeyboard(ctx, inlineKeyboardRows, "Добавить подписку", chatID)
		if err != nil {
			return errors.Wrap(err, "subscription add: send inline keyboard")
		}
	}

	return nil
}

func (b *Bot) Feedback(ctx context.Context, chatID int64) error {
	mess := new(bytes.Buffer)
	if err := b.tmpl.ExecuteTemplate(mess, tmplTgFeedback, map[string]any{
		"Cancel": app.Cancel,
	}); err != nil {
		return errors.Wrap(err, "feedback")
	}

	err := b.client.SendMessage(ctx, mess.String(), chatID, retryCount)
	if err != nil {
		return errors.Wrap(err, "feedback")
	}

	return nil
}

func (b *Bot) FeedbackMessage(ctx context.Context, chatID int64, text string, info []any) error {
	cfg := configs.Get(ctx)

	mess := new(bytes.Buffer)
	if err := b.tmpl.ExecuteTemplate(mess, tmplTgFeedbackMess, map[string]any{
		"Text": text,
		"Info": info,
	}); err != nil {
		return errors.Wrap(err, "feedback message")
	}

	err := b.client.SendMessage(ctx, mess.String(), cfg.Telegram.FeedbackChatID, retryCount)
	if err != nil {
		return errors.Wrap(err, "feedback message")
	}

	err = b.client.SendMessage(ctx, "Сообщение отправлено разработчику сервиса", chatID, retryCount)
	if err != nil {
		return errors.Wrap(err, "feedback message")
	}

	return nil
}

func (b *Bot) Map(ctx context.Context, chatID int64) error {
	cfg := configs.Get(ctx)

	mapURL, err := url.Parse(cfg.App.HostURL)
	if err != nil {
		return errors.Wrap(err, "map: parse url")
	}

	inlineKeyboardRows := make([]*client.KeyboardRow, 0, countSABInline)
	inlineKeyboardRow := client.NewKeyboardRow()
	inlineKeyboardRow.AddURL("Открыть карту объявлений", mapURL.String())
	inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)
	err = b.client.SendInlineKeyboard(ctx, inlineKeyboardRows, "Карта объявлений", chatID)
	if err != nil {
		return errors.Wrap(err, "map: send inline keyboard")
	}

	return nil
}
