package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/cmd/bot"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
	"go.uber.org/zap"
)

var (
	version     string
	gitRevision string
	buildTime   string
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := bot.NewCommand(
		filepath.Base(os.Args[0]),
		utils.BuildVersion(version, gitRevision, buildTime),
	).ExecuteContext(ctx); err != nil {
		log.Fatal("bot failed", zap.Error(err))
	}
}
