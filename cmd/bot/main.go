package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/clients/tg"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/expense/inmemory"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		cancel()
	}()

	cfg, err := config.New()
	if err != nil {
		log.Fatal("config init failed: ", err)
	}

	tgClient, err := tg.New(cfg)
	if err != nil {
		log.Fatal("tg client init failed: ", err)
	}

	expenses := inmemory.New()

	bot := model.NewBot(tgClient, expenses)

	tgClient.ListenUpdates(ctx, bot)
}
