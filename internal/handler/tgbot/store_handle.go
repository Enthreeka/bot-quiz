package tgbot

import (
	"context"
	"github.com/Enthreeka/tg-bot-quiz/internal/entity"
	store "github.com/Enthreeka/tg-bot-quiz/pkg/local_storage"
	"github.com/Enthreeka/tg-bot-quiz/pkg/tg_bot_api/coverter"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) isStateExist(userID int64) (*store.Data, bool) {
	data, exist := b.store.Read(userID)
	return data, exist
}

func (b *Bot) isStoreProcessing(ctx context.Context, update *tgbotapi.Update) (bool, error) {
	userID := update.Message.From.ID
	storeData, isExist := b.isStateExist(userID)
	if !isExist || storeData == nil {
		return false, nil
	}
	defer b.store.Delete(userID)

	return b.switchStoreData(ctx, update, storeData)
}

func (b *Bot) switchStoreData(ctx context.Context, update *tgbotapi.Update, storeData *store.Data) (bool, error) {
	var (
		err error
	)

	switch storeData.OperationType {
	case store.AdminCreate:
		err = b.userService.UpdateRoleByUsername(ctx, entity.AdminType, update.Message.Text)
		if err != nil {
			b.log.Error("isStoreExist::store.AdminCreate:UpdateRoleByUsername: %v", err)
		}
	case store.AdminDelete:
		err = b.userService.UpdateRoleByUsername(ctx, entity.UserType, update.Message.Text)
		if err != nil {
			b.log.Error("isStoreExist::store.AdminDelete:userRepo.UpdateRoleByUsername: %v", err)
		}

	case store.QuizCreate:
		if _, err = b.quizService.CreateQuestion(ctx, nil, &entity.Question{QuestionName: coverter.ConvertToMarkdownV2(update.Message.Text, update.Message.Entities),
			CreatedByUser: update.FromChat().ID}); err != nil {
			b.log.Error("isStoreExist::store.QuizCreate:CreateQuestion: %v", err)
		}
	case store.QuizUpdateAnswer:
		if err = b.quizService.QuizUpdateAnswer(ctx, update.Message.Text, storeData.QuestionID); err != nil {
			b.log.Error("isStoreExist::store.QuizUpdateAnswer: %v", err)
		}

	case store.QuizUpdateImage:
		if err = b.quizService.UpdateImage(ctx, storeData.QuestionID, update.Message.Photo[len(update.Message.Photo)-1].FileID); err != nil {
			b.log.Error("isStoreExist::store.QuizUpdateImage: %v", err)
		}

	case store.QuizUpdateQuestion:
		if err = b.quizService.UpdateQuestion(ctx, storeData.QuestionID, update.Message.Text); err != nil {
			b.log.Error("isStoreExist::store.QuizUpdateQuestion: %v", err)
		}
	case store.QuizUpdateOldAnswer:
		if err = b.quizService.QuizUpdateOldAnswer(ctx, coverter.ConvertToMarkdownV2(update.Message.Text, update.Message.Entities), storeData.QuestionID); err != nil {
			b.log.Error("isStoreExist::store.QuizUpdateQuestion: %v", err)
		}
	default:
		return false, nil
	}

	if err == nil {
		b.response(storeData, update)
	}
	return true, err
}
