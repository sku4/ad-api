package server

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type IHandler interface {
	IncomingMessage(context.Context, tgbotapi.Update)
}

type Func func(context.Context, tgbotapi.Update)

func (f Func) IncomingMessage(ctx context.Context, update tgbotapi.Update) {
	f(ctx, update)
}
