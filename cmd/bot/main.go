package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/cmd/bot"
	"go.uber.org/zap"
)

var (
	version     string
	gitRevision string
	buildTime   string
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

	cmd := bot.NewCommand(filepath.Base(os.Args[0]), buildVersion())
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatal("command failed", zap.Error(err))
	}
}

func buildVersion() string {
	builder := strings.Builder{}

	if len(version) == 0 {
		builder.WriteString(gitRevision)
	} else {
		builder.WriteString(fmt.Sprintf("%s (%s)", version, gitRevision))
	}

	builder.WriteString(fmt.Sprintf(" %s", buildTime))

	return builder.String()
}
