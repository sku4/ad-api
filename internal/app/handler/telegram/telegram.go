package telegram

import (
	"context"
	"fmt"
	"hash/crc32"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sku4/ad-api/internal/app"
	in "github.com/sku4/ad-api/pkg/telegram/inline"
	"github.com/sku4/ad-parser/pkg/logger"
)

const (
	infoCap = 7
)

func (h *Handler) IncomingMessage(ctx context.Context, update tgbotapi.Update) {
	log := logger.Get()

	var err error
	var botMessage *tgbotapi.Message
	if update.Message != nil {
		botMessage = update.Message
	} else if update.ChannelPost != nil {
		botMessage = update.ChannelPost
	}

	if botMessage != nil {
		// ignore supergroup type messages
		if botMessage.Chat.IsSuperGroup() {
			return
		}

		isFeedback := false
		var feedbackID uint32
		if botMessage.From != nil {
			feedbackID = crc32.Checksum([]byte(
				fmt.Sprintf("%d/%d", botMessage.Chat.ID, botMessage.From.ID)), h.crcTable)
		} else {
			feedbackID = crc32.Checksum([]byte(fmt.Sprintf("%d", botMessage.Chat.ID)), h.crcTable)
		}

		if botMessage.IsCommand() {
			switch botMessage.Command() {
			case app.Start:
				err = h.services.Bot.Start(ctx, botMessage.Chat.ID)
			case app.Help:
				err = h.services.Bot.Help(ctx, botMessage.Chat.ID)
			case app.Feedback:
				err = h.services.Bot.Feedback(ctx, botMessage.Chat.ID)
				if err == nil {
					isFeedback = true
				}
			case app.AddSub:
				err = h.services.Bot.AddSubscription(ctx, botMessage.Chat.ID, botMessage.Chat.IsPrivate())
			case app.ManageSubsEntity:
				err = h.services.Bot.ManageSubscriptionsPage(ctx, botMessage.Chat.ID, 0,
					in.NewPaginationInline())
			case app.Map:
				err = h.services.Bot.Map(ctx, botMessage.Chat.ID)
			}
		} else {
			if _, ok := h.feedback.Get(feedbackID); ok {
				info := make([]any, 0, infoCap)
				if botMessage.Chat != nil {
					info = append(info, fmt.Sprintf("Chat id: %d", botMessage.Chat.ID))
					if botMessage.Chat.Type != "" {
						info = append(info, fmt.Sprintf("Chat type: %s", botMessage.Chat.Type))
					}
					if botMessage.Chat.Title != "" {
						info = append(info, fmt.Sprintf("Chat title: %s", botMessage.Chat.Title))
					}
				}
				if botMessage.From != nil {
					info = append(info, fmt.Sprintf("From id: %d", botMessage.From.ID))
					if botMessage.From.FirstName != "" {
						info = append(info, fmt.Sprintf("First name: %s", botMessage.From.FirstName))
					}
					if botMessage.From.LastName != "" {
						info = append(info, fmt.Sprintf("Last name: %s", botMessage.From.LastName))
					}
					if botMessage.From.UserName != "" {
						info = append(info, fmt.Sprintf("User name: @%s", botMessage.From.UserName))
					}
				}
				err = h.services.Bot.FeedbackMessage(ctx, botMessage.Chat.ID, botMessage.Text, info)
			}
		}

		if isFeedback {
			h.feedback.Add(feedbackID, struct{}{})
		} else {
			h.feedback.Remove(feedbackID)
		}
	} else if update.CallbackQuery != nil {
		inline, errInline := in.NewInline().UnSerialize(update.CallbackQuery.Data)
		if errInline != nil {
			log.Warnf("error processing message: %s", errInline)
			return
		}

		switch inline.Entity {
		case app.ManageSubsEntity:
			switch inline.Command {
			case app.CommandPage:
				err = h.services.Bot.ManageSubscriptionsPage(ctx,
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					inline.Data.(*in.PaginationInline))
			case app.CommandOpen:
				err = h.services.Bot.ManageSubscriptionsOpen(ctx,
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					inline.Data.(*in.ManageSubInline))
			case app.CommandDelete:
				err = h.services.Bot.ManageSubscriptionsDelete(ctx,
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					update.CallbackQuery.ID,
					inline.Data.(*in.ManageSubInline))
			default:
				log.Warnf("incoming message: command %s not found", inline.Command)
				return
			}
		default:
			log.Warnf("incoming message: entity %s not found", inline.Entity)
			return
		}
	}

	if err != nil {
		log.Warnf("error processing message: %s", err)
	}
}
