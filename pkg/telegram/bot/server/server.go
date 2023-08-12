package server

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	client *tgbotapi.BotAPI
}

func NewServer(client *tgbotapi.BotAPI) (*Server, error) {
	return &Server{
		client: client,
	}, nil
}

func (s *Server) Run(ctx context.Context, h IHandler, timeout int) {
	h = Metrics(h)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout

	updates := s.client.GetUpdatesChan(u)
	for update := range updates {
		h.IncomingMessage(ctx, update)
	}
}

func (s *Server) Stop() {
	s.client.StopReceivingUpdates()
}
