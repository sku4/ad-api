package client

import (
	"context"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

//go:generate mockgen -source=client.go -destination=mocks/client.go

const (
	KeyboardButtonTypeSwitch = "switch"
	KeyboardButtonTypeURL    = "url"
	KeyboardButtonTypeWebApp = "web_app"
	randMax                  = 100
	retryAfter               = time.Second * 3
)

type BotClient interface {
	SendMessage(context.Context, string, int64, int) error
	SendMediaPhoto(context.Context, []string, []string, int64, int) error
	SendUploadPhoto(context.Context, []string, []string, int64, int) error
	SendInlineKeyboard(context.Context, []*KeyboardRow, string, int64) error
	SendCallbackQuery(context.Context, []*KeyboardRow, string, int, int64) error
	SendKeyboard(context.Context, []*KeyboardRow, string, int64, bool, bool) error
	AnswerCallbackQuery(ctx context.Context, id, text string) error
	AnswerCallbackQueryWithAlert(ctx context.Context, id, text string) error
}

type Client struct {
	BotClient
	client *tgbotapi.BotAPI
}

type KeyboardRow struct {
	buttons []KeyboardButton
}

type KeyboardButton struct {
	k, v, t string
}

func NewKeyboardRow() *KeyboardRow {
	buttons := make([]KeyboardButton, 0)
	return &KeyboardRow{
		buttons: buttons,
	}
}

func (i *KeyboardRow) Add(k, v string) {
	i.buttons = append(i.buttons, KeyboardButton{
		k, v, "",
	})
}

func (i *KeyboardRow) AddSwitch(k, v string) {
	i.buttons = append(i.buttons, KeyboardButton{
		k, v, KeyboardButtonTypeSwitch,
	})
}

func (i *KeyboardRow) AddURL(k, v string) {
	i.buttons = append(i.buttons, KeyboardButton{
		k, v, KeyboardButtonTypeURL,
	})
}

func (i *KeyboardRow) AddWebApp(k, v string) {
	i.buttons = append(i.buttons, KeyboardButton{
		k, v, KeyboardButtonTypeWebApp,
	})
}

func NewClient(client *tgbotapi.BotAPI) (*Client, error) {
	return &Client{
		client: client,
	}, nil
}

func (c *Client) SendMessage(ctx context.Context, message string, chatID int64, retry int) error {
	_ = ctx

	params := make(tgbotapi.Params)
	err := params.AddFirstValid("chat_id", chatID, "")
	if err != nil {
		return errors.Wrap(err, "send message")
	}
	params.AddNonZero("reply_to_message_id", 0)
	params.AddBool("disable_notification", false)
	params.AddBool("allow_sending_without_reply", false)
	params.AddBool("protect_content", false)
	err = params.AddInterface("reply_markup", nil)
	if err != nil {
		return errors.Wrap(err, "send message")
	}

	params.AddNonEmpty("text", message)
	params.AddBool("disable_web_page_preview", true)
	params.AddNonEmpty("parse_mode", tgbotapi.ModeMarkdown)
	err = params.AddInterface("entities", nil)
	if err != nil {
		return errors.Wrap(err, "send message")
	}

	var resp *tgbotapi.APIResponse
	for i := 0; i < retry; i++ {
		resp, err = c.client.MakeRequest("sendMessage", params)
		if err != nil {
			if resp.Parameters != nil && i != retry-1 {
				time.Sleep(time.Second * time.Duration(resp.Parameters.RetryAfter))
			}
		} else {
			break
		}
	}

	if err != nil {
		return errors.Wrap(err, "send message")
	}

	return nil
}

//nolint:gosec
func (c *Client) SendMediaPhoto(ctx context.Context, urls []string, caption []string, chatID int64, retry int) error {
	_ = ctx

	photos := make([]interface{}, 0, len(urls))
	photosRetry := make([]interface{}, 0, len(urls))
	for i, u := range urls {
		up, err := url.Parse(u)
		if err == nil {
			q := up.Query()
			rr := rand.New(rand.NewSource(time.Now().UnixNano()))
			q.Set("r_retry", strconv.Itoa(rr.Intn(randMax)))
			up.RawQuery = q.Encode()
		}
		photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(u))
		photoRetry := photo
		if up != nil {
			photoRetry = tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(up.String()))
		}
		if len(caption) > i {
			photo.Caption = caption[i]
			photo.ParseMode = tgbotapi.ModeMarkdown
			photoRetry.Caption = caption[i]
			photoRetry.ParseMode = tgbotapi.ModeMarkdown
		}
		photos = append(photos, photo)
		photosRetry = append(photosRetry, photoRetry)
	}
	msg := tgbotapi.NewMediaGroup(chatID, photos)
	msgRetry := tgbotapi.NewMediaGroup(chatID, photosRetry)

	var resp *tgbotapi.APIResponse
	var err error
	for i := 0; i < retry; i++ {
		if i > 0 {
			msg = msgRetry
		}

		params := make(tgbotapi.Params)
		err = params.AddFirstValid("chat_id", chatID, "")
		if err != nil {
			return errors.Wrap(err, "send media photo")
		}
		params.AddBool("disable_notification", false)
		params.AddNonZero("reply_to_message_id", 0)

		err = params.AddInterface("media", msg.Media)
		if err != nil {
			return errors.Wrap(err, "send media photo")
		}

		resp, err = c.client.MakeRequest("sendMediaGroup", params)
		if err != nil {
			if resp.Parameters != nil && i != retry-1 {
				time.Sleep(time.Second * time.Duration(resp.Parameters.RetryAfter))
			}
		} else {
			break
		}
	}

	if err != nil {
		return errors.Wrap(err, "send media photo")
	}

	return nil
}

func (c *Client) SendUploadPhoto(ctx context.Context, urls []string, caption []string, chatID int64, retry int) error {
	_ = ctx

	photos := make([]interface{}, 0, len(urls))
	for i, u := range urls {
		photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(u))
		if len(caption) > i {
			photo.Caption = caption[i]
			photo.ParseMode = tgbotapi.ModeMarkdown
		}
		photos = append(photos, photo)
	}
	msg := tgbotapi.NewMediaGroup(chatID, photos)

	var err error
	for i := 0; i < retry; i++ {
		_, err = c.client.Request(msg)
		if err != nil {
			if i != retry-1 {
				time.Sleep(retryAfter)
			}
		} else {
			break
		}
	}

	if err != nil {
		return errors.Wrap(err, "send upload photo")
	}

	return nil
}

func (c *Client) SendInlineKeyboard(ctx context.Context, keyboardRows []*KeyboardRow,
	message string, chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = getInlineKeyboard(ctx, keyboardRows)

	_, err := c.client.Send(msg)
	if err != nil {
		return errors.Wrap(err, "send inline keyboard")
	}

	return nil
}

func (c *Client) SendCallbackQuery(ctx context.Context, keyboardRows []*KeyboardRow,
	message string, messageID int, chatID int64) error {
	msgEdit := tgbotapi.NewEditMessageText(chatID, messageID, message)
	msgEdit.ParseMode = tgbotapi.ModeMarkdown
	msgEdit.DisableWebPagePreview = true
	msgEdit.ReplyMarkup = getInlineKeyboard(ctx, keyboardRows)

	if _, err := c.client.Send(msgEdit); err != nil {
		return errors.Wrap(err, "send markup callback query")
	}

	return nil
}

func (c *Client) SendKeyboard(ctx context.Context, keyboardRows []*KeyboardRow,
	message string, chatID int64, oneTimeKeyboard bool, resizeKeyboard bool) error {
	msgEdit := tgbotapi.NewMessage(chatID, message)
	msgEdit.ParseMode = tgbotapi.ModeMarkdown
	msgEdit.DisableWebPagePreview = true
	keyboard := getKeyboard(ctx, keyboardRows)
	keyboard.OneTimeKeyboard = oneTimeKeyboard
	keyboard.ResizeKeyboard = resizeKeyboard
	msgEdit.ReplyMarkup = keyboard

	if _, err := c.client.Send(msgEdit); err != nil {
		return errors.Wrap(err, "send keyboard")
	}

	return nil
}

func (c *Client) AnswerCallbackQueryWithAlert(ctx context.Context, id, text string) error {
	_ = ctx

	cfg := tgbotapi.NewCallbackWithAlert(id, text)
	if _, err := c.client.Request(cfg); err != nil {
		return errors.Wrap(err, "send answer callback query")
	}

	return nil
}

func (c *Client) AnswerCallbackQuery(ctx context.Context, id, text string) error {
	_ = ctx

	cfg := tgbotapi.NewCallback(id, text)
	if _, err := c.client.Request(cfg); err != nil {
		return errors.Wrap(err, "send answer callback query")
	}

	return nil
}

func getInlineKeyboard(ctx context.Context, inlineKeyboardRows []*KeyboardRow) *tgbotapi.InlineKeyboardMarkup {
	_ = ctx

	var keyboardButtons [][]tgbotapi.InlineKeyboardButton
	for _, row := range inlineKeyboardRows {
		maxButtons := 8
		inlineRow := make([]tgbotapi.InlineKeyboardButton, 0, maxButtons)
		for _, b := range row.buttons {
			switch b.t {
			case KeyboardButtonTypeSwitch:
				inlineRow = append(inlineRow, tgbotapi.NewInlineKeyboardButtonSwitch(b.k, b.v))
			case KeyboardButtonTypeURL:
				inlineRow = append(inlineRow, tgbotapi.NewInlineKeyboardButtonURL(b.k, b.v))
			case KeyboardButtonTypeWebApp:
				inlineRow = append(inlineRow, tgbotapi.NewInlineKeyboardButtonWebApp(b.k, tgbotapi.WebAppInfo{
					URL: b.v,
				}))
			default:
				inlineRow = append(inlineRow, tgbotapi.NewInlineKeyboardButtonData(b.k, b.v))
			}
			if len(inlineRow) == maxButtons {
				keyboardButtons = append(keyboardButtons, inlineRow)
				inlineRow = make([]tgbotapi.InlineKeyboardButton, 0, maxButtons)
			}
		}
		if len(inlineRow) > 0 {
			keyboardButtons = append(keyboardButtons, inlineRow)
		}
	}

	if len(keyboardButtons) == 0 {
		inlineRow := make([]tgbotapi.InlineKeyboardButton, 0)
		keyboardButtons = append(keyboardButtons, inlineRow)
	}

	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboardButtons,
	}
}

func getKeyboard(ctx context.Context, inlineKeyboardRows []*KeyboardRow) *tgbotapi.ReplyKeyboardMarkup {
	_ = ctx

	var keyboardButtons [][]tgbotapi.KeyboardButton
	for _, row := range inlineKeyboardRows {
		maxButtons := 8
		inlineRow := make([]tgbotapi.KeyboardButton, 0, maxButtons)
		for _, b := range row.buttons {
			switch b.t {
			case KeyboardButtonTypeWebApp:
				inlineRow = append(inlineRow, tgbotapi.NewKeyboardButtonWebApp(b.k, tgbotapi.WebAppInfo{
					URL: b.v,
				}))
			default:
				inlineRow = append(inlineRow, tgbotapi.NewKeyboardButton(b.k))
			}
			if len(inlineRow) == maxButtons {
				keyboardButtons = append(keyboardButtons, inlineRow)
				inlineRow = make([]tgbotapi.KeyboardButton, 0, maxButtons)
			}
		}
		if len(inlineRow) > 0 {
			keyboardButtons = append(keyboardButtons, inlineRow)
		}
	}

	if len(keyboardButtons) == 0 {
		inlineRow := make([]tgbotapi.KeyboardButton, 0)
		keyboardButtons = append(keyboardButtons, inlineRow)
	}

	return &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: keyboardButtons,
	}
}
