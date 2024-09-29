package main

import (
	"context"
	"github.com/Enthreeka/tg-bot-quiz/internal/bot"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	tgBot := bot.NewBot()

	tgBot.Run(ctx)
}
