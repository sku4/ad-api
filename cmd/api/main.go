package main

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/sku4/ad-api/configs"
	"github.com/sku4/ad-api/internal/app/handler/rest"
	"github.com/sku4/ad-api/internal/app/handler/telegram"
	"github.com/sku4/ad-api/internal/app/repository"
	"github.com/sku4/ad-api/internal/app/service"
	"github.com/sku4/ad-api/pkg/telegram/bot/client"
	"github.com/sku4/ad-api/pkg/telegram/bot/server"
	"github.com/sku4/ad-parser/pkg/logger"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

func main() {
	// init config
	log := logger.Get()
	cfg, err := configs.Init()
	if err != nil {
		log.Errorf("error init config: %s", err)
		return
	}

	// init tarantool
	conn, err := pool.Connect(cfg.Tarantool.Servers, tarantool.Opts{
		Timeout:   cfg.Tarantool.Timeout,
		Reconnect: cfg.Tarantool.ReconnectInterval,
	})
	if err != nil {
		log.Errorf("error tarantool connection refused: %s", err)
		return
	}
	defer func() {
		errs := conn.Close()
		for _, e := range errs {
			log.Errorf("error close connection pool: %s", e)
		}
	}()

	// init context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx = configs.Set(ctx, cfg)

	tgClient, tgServer, err := initTelegramBot(cfg)
	if err != nil {
		log.Errorf("error init telegram bot: %s", err)
		return
	}

	tmpl, err := initTemplates(cfg)
	if err != nil {
		log.Errorf("error init templates: %s", err)
		return
	}

	repos := repository.NewRepository(conn)
	services := service.NewService(repos, tgClient, tmpl)
	tgHandlers := telegram.NewHandler(services)
	restHandlers := rest.NewHandler(services)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// run telegram server
	go func() {
		log.Info("Telegram listening for messages")
		tgServer.Run(ctx, tgHandlers, cfg.Telegram.Timeout)
		quit <- nil
	}()

	// run rest server
	routes := restHandlers.InitRoutes()
	restServer := rest.NewServer(ctx, routes)
	go func() {
		log.Info(fmt.Sprintf("Rest server is running on: %d", cfg.Rest.Port))
		if errRest := restServer.ListenAndServe(); errRest != nil {
			log.Infof("Rest server %s", errRest)
			quit <- nil
		}
	}()

	// create hot cache and etc
	restHandlers.OnLoad(ctx)

	log.Infof("App Started")

	// graceful shutdown
	log.Infof("Got signal %v, attempting graceful shutdown", <-quit)
	cancel()
	log.Info("Context is stopped")

	err = restServer.Shutdown(ctx)
	if err != nil {
		log.Errorf("error rest server shutdown: %s", err)
	}

	tgServer.Stop()
	log.Info("Telegram listening stopped")

	errs := conn.CloseGraceful()
	for _, e := range errs {
		log.Errorf("error close graceful connection pool: %s", e)
	}

	log.Info("App Shutting Down")
}

func initTelegramBot(cfg *configs.Config) (tgClient *client.Client, tgServer *server.Server, err error) {
	tgBot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error init telegram bot")
	}

	tgClient, err = client.NewClient(tgBot)
	if err != nil {
		return nil, nil, errors.Wrap(err, "telegram client init failed")
	}

	tgServer, err = server.NewServer(tgBot)
	if err != nil {
		return nil, nil, errors.Wrap(err, "telegram server init failed")
	}

	return
}

func initTemplates(cfg *configs.Config) (*template.Template, error) {
	var err error
	tmpl := template.New("")

	for _, folder := range cfg.Template.Folders {
		tmpl, err = tmpl.ParseGlob(fmt.Sprintf("%s/*.gohtml", folder))
		if err != nil {
			return nil, err
		}
	}

	return tmpl, nil
}
