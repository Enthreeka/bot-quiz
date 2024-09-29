package tgbot

import (
	"context"
	store "github.com/Enthreeka/tg-bot-quiz/pkg/local_storage"
	"github.com/Enthreeka/tg-bot-quiz/pkg/tg_bot_api/markup"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	success = "Операция выполнена успешно. "
)

func (b *Bot) response(storeData *store.Data, update *tgbotapi.Update) {
	var (
		messageId int
		userID    = update.FromChat().ID
	)

	if update.Message != nil {
		messageId = update.Message.MessageID
	} else if update.CallbackQuery != nil {
		messageId = update.CallbackQuery.Message.MessageID
	}

	if resp, err := b.bot.Request(tgbotapi.NewDeleteMessage(userID, messageId)); nil != err || !resp.Ok {
		b.log.Error("failed to delete message id %d (%s): %v", storeData.CurrentMsgID, string(resp.Result), err)
	}

	if resp, err := b.bot.Request(tgbotapi.NewDeleteMessage(userID, storeData.PreferMsgID)); nil != err || !resp.Ok {
		b.log.Error("failed to delete message id %d (%s): %v", storeData.PreferMsgID, string(resp.Result), err)
	}

	text, markup := b.responseText(storeData)
	if _, err := b.tgMsg.SendEditMessage(userID, storeData.CurrentMsgID, markup, text); err != nil {
		b.log.Error("failed to send telegram message: ", err)
	}
}

func (b *Bot) responseText(storeData *store.Data) (string, *tgbotapi.InlineKeyboardMarkup) {
	switch storeData.OperationType {
	case store.AdminCreate:
		return success + "Пользователь получил администраторские права.", &markup.UserSetting
	case store.AdminDelete:
		return success + "Пользователь лишился администраторских прав.", &markup.UserSetting
	case store.QuizCreate:
		return success, &markup.QuizSetting
	case store.QuizUpdateAnswer, store.QuizUpdateImage, store.QuizUpdateQuestion, store.QuizUpdateOldAnswer:
		question, err := b.quizService.GetQuestionByID(context.Background(), storeData.QuestionID)
		if err != nil {
			b.log.Error("failed to get question by id: %v", err)
			return "", nil
		}

		questionSetting := markup.QuestionSetting(storeData.QuestionID)
		text := question.QuestionName
		return text, &questionSetting
	}
	return success, nil
}
