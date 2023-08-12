package bot

import (
	"html/template"
	"sync"
	"time"

	"github.com/sku4/ad-api/internal/app/repository"
	"github.com/sku4/ad-api/model"
	"github.com/sku4/ad-api/pkg/telegram/bot/client"
)

type Bot struct {
	repos           *repository.Repository
	client          client.BotClient
	tmpl            *template.Template
	mu              sync.RWMutex
	streetCacheTime time.Time
	streetsCache    []*model.Street
}

func NewBot(repos *repository.Repository, client client.BotClient, tmpl *template.Template) *Bot {
	return &Bot{
		repos:  repos,
		client: client,
		tmpl:   tmpl,
	}
}
