package tg_bot_api

import (
	"fmt"
	"github.com/Enthreeka/tg-bot-quiz/internal/entity"
	"github.com/Enthreeka/tg-bot-quiz/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Message interface {
	SendNewMessage(chatID int64, markup *tgbotapi.InlineKeyboardMarkup, text string) (int, error)
	SendEditMessage(chatID int64, messageID int, markup *tgbotapi.InlineKeyboardMarkup, text string) (int, error)
	SendDocument(chatID int64, fileName string, fileIDBytes *[]byte, text string) (int, error)
	SendMessageToChannel(username string, quiz *entity.Quiz) error
	SendMessageToUser(chatID int64, quiz *entity.Quiz) (int, error)
}

type TelegramMsg struct {
	log *logger.Logger
	bot *tgbotapi.BotAPI
}

func NewMessageSetting(bot *tgbotapi.BotAPI, log *logger.Logger) *TelegramMsg {
	return &TelegramMsg{
		bot: bot,
		log: log,
	}
}

func (t *TelegramMsg) SendNewMessage(chatID int64, markup *tgbotapi.InlineKeyboardMarkup, text string) (int, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	if markup != nil {
		msg.ReplyMarkup = &markup
	}

	sendMsg, err := t.bot.Send(msg)
	if err != nil {
		t.log.Error("failed to send message", zap.Error(err))
		return 0, err
	}

	return sendMsg.MessageID, nil
}

func (t *TelegramMsg) SendEditMessage(chatID int64, messageID int, markup *tgbotapi.InlineKeyboardMarkup, text string) (int, error) {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = tgbotapi.ModeHTML

	if markup != nil {
		msg.ReplyMarkup = markup
	}

	sendMsg, err := t.bot.Send(msg)
	if err != nil {
		t.log.Error("failed to send msg: %v", err)
		return 0, err
	}

	return sendMsg.MessageID, nil
}

func (t *TelegramMsg) SendDocument(chatID int64, fileName string, fileIDBytes *[]byte, text string) (int, error) {
	msg := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: *fileIDBytes,
	})
	msg.ParseMode = tgbotapi.ModeHTML
	msg.Caption = text

	sendMsg, err := t.bot.Send(msg)
	if err != nil {
		t.log.Error("failed to send msg: %v", err)
		return 0, err
	}

	return sendMsg.MessageID, nil
}

func (t *TelegramMsg) SendMessageToChannel(username string, quiz *entity.Quiz) error {
	if quiz.Question.FileID != nil {
		publicationPhoto := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(*quiz.Question.FileID))
		msg := tgbotapi.NewPhotoToChannel(username, publicationPhoto.Media)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		buttonMarkup := buttonQualifier(quiz.Answer)
		if buttonMarkup != nil {
			msg.ReplyMarkup = &buttonMarkup
		}
		if quiz.Question.QuestionName != "" {
			msg.Caption = quiz.Question.QuestionName
		}

		if _, err := t.bot.Send(msg); err != nil {
			t.log.Error("failed to send message: %v", err)
			return err
		}
		return nil
	}

	msg := tgbotapi.NewMessageToChannel(username, "")
	msg.DisableWebPagePreview = true
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	buttonMarkup := buttonQualifier(quiz.Answer)
	if buttonMarkup != nil {
		msg.ReplyMarkup = &buttonMarkup
	}
	if quiz.Question.QuestionName != "" {
		msg.Text = quiz.Question.QuestionName
	}

	if _, err := t.bot.Send(msg); err != nil {
		t.log.Error("failed to send message", err)
		return err
	}

	return nil
}

func (t *TelegramMsg) SendMessageToUser(chatID int64, quiz *entity.Quiz) (int, error) {
	if quiz.Question.FileID != nil {
		publicationPhotoPhoto := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(*quiz.Question.FileID))
		msg := tgbotapi.NewPhoto(chatID, publicationPhotoPhoto.Media)
		buttonMarkup := buttonQualifier(quiz.Answer)
		if buttonMarkup != nil {
			msg.ReplyMarkup = &buttonMarkup
		}
		if quiz.Question.QuestionName != "" {
			msg.Caption = quiz.Question.QuestionName
		}
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		sendMsg, err := t.bot.Send(msg)
		if err != nil {
			t.log.Error("failed to send message: %v", err)
			return 0, err
		}
		return sendMsg.MessageID, nil
	}

	msg := tgbotapi.NewMessage(chatID, "")
	msg.DisableWebPagePreview = true
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	buttonMarkup := buttonQualifier(quiz.Answer)
	if buttonMarkup != nil {
		msg.ReplyMarkup = &buttonMarkup
	}
	if quiz.Question.QuestionName != "" {
		msg.Text = quiz.Question.QuestionName
	}

	sendMsg, err := t.bot.Send(msg)
	if err != nil {
		t.log.Error("failed to send message", err)
		return 0, err
	}

	return sendMsg.MessageID, nil
}

func buttonQualifier(answers []entity.Answer) *tgbotapi.InlineKeyboardMarkup {
	if len(answers) == 0 {
		return nil
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	buttonsPerRow := 1
	for i, el := range answers {
		btn := tgbotapi.NewInlineKeyboardButtonData(el.Answer, fmt.Sprintf("quiz_answer_%d", el.ID))

		row = append(row, btn)

		if (i+1)%buttonsPerRow == 0 || i == len(answers)-1 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return &markup
}
