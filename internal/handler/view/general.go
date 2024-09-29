package view

import (
	"context"
	"github.com/Enthreeka/tg-bot-quiz/internal/handler/tgbot"
	"github.com/Enthreeka/tg-bot-quiz/pkg/logger"
	customMsg "github.com/Enthreeka/tg-bot-quiz/pkg/tg_bot_api"
	"github.com/Enthreeka/tg-bot-quiz/pkg/tg_bot_api/markup"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ViewGeneral struct {
	log   *logger.Logger
	tgMsg customMsg.Message
}

func NewViewGeneral(
	log *logger.Logger,
	tgMsg customMsg.Message,
) *ViewGeneral {
	return &ViewGeneral{
		log:   log,
		tgMsg: tgMsg,
	}
}

func (c *ViewGeneral) CallbackStartAdminPanel() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {

		if _, err := c.tgMsg.SendNewMessage(update.FromChat().ID, &markup.StartMenu, "Панель управления"); err != nil {
			return err
		}

		return nil
	}
}
