package service

import (
	"context"
	"html/template"

	"github.com/sku4/ad-api/internal/app/repository"
	"github.com/sku4/ad-api/internal/app/service/bot"
	"github.com/sku4/ad-api/model"
	"github.com/sku4/ad-api/pkg/telegram/bot/client"
	"github.com/sku4/ad-api/pkg/telegram/inline"
	modelAd "github.com/sku4/ad-parser/pkg/ad/model"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go

type Bot interface {
	ManageSubscriptionsPage(ctx context.Context, chatID int64, messageID int, page *inline.PaginationInline) error
	ManageSubscriptionsOpen(ctx context.Context, chatID int64, messageID int, sub *inline.ManageSubInline) error
	ManageSubscriptionsDelete(ctx context.Context, chatID int64, messageID int, queryID string,
		sub *inline.ManageSubInline) error
	AddSubscription(context.Context, int64, bool) error
	Help(context.Context, int64) error
	Start(context.Context, int64) error
	Feedback(context.Context, int64) error
	FeedbackMessage(context.Context, int64, string, []any) error
	Map(context.Context, int64) error
	WebAppAddSubscriptionIndex(context.Context, bool, string) ([]byte, error)
	WebAppStreets(context.Context) ([]*model.Street, error)
	WebAppAddSubscription(context.Context, *modelAd.SubscriptionTnt) error
	MapIndex(context.Context, string, *model.AdFilterFields) ([]byte, error)
	MapAds(context.Context, []*model.AdFilterGroup) ([]*model.Ad, error)
	MapFilter(context.Context, map[string]any) ([]*model.AdLocation, error)
	SetUsersMetrics()
}

type Service struct {
	Bot
}

func NewService(repos *repository.Repository, client client.BotClient, tmpl *template.Template) *Service {
	return &Service{
		Bot: bot.NewBot(repos, client, tmpl),
	}
}
