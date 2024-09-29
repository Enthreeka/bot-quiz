package callback

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Enthreeka/tg-bot-quiz/internal/entity"
	"github.com/Enthreeka/tg-bot-quiz/internal/handler/tgbot"
	service "github.com/Enthreeka/tg-bot-quiz/internal/usecase"
	customErr "github.com/Enthreeka/tg-bot-quiz/pkg/bot_error"
	"github.com/Enthreeka/tg-bot-quiz/pkg/excel"
	store "github.com/Enthreeka/tg-bot-quiz/pkg/local_storage"
	"github.com/Enthreeka/tg-bot-quiz/pkg/logger"
	customMsg "github.com/Enthreeka/tg-bot-quiz/pkg/tg_bot_api"
	"github.com/Enthreeka/tg-bot-quiz/pkg/tg_bot_api/markup"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	QuestionGET    = "get"
	QuestionDELETE = "delete"
)

const jsonExample = "{\n \"варианты_ответы\": [\n  {\n   \"ответ\": \"Answer 123124\",\n   \"цена_ответа\": 10\n  },\n  {\n   \"ответ\": \"Answer e23fsdf\",\n   \"цена_ответа\": 20\n  },\n  {\n   \"ответ\": \"Answer 33249w8ueryfsd\",\n   \"цена_ответа\": 30\n  }\n ]\n}"

type CallbackQuiz interface {
	CallbackShowQuizSetting() tgbot.ViewFunc
	CallbackCreateQuizQuestion() tgbot.ViewFunc
	CallbackListQuestion() tgbot.ViewFunc
	CallbackDeleteQuestion() tgbot.ViewFunc
	CallbackGetQuestion() tgbot.ViewFunc
	CallbackDeleteByIDQuestion() tgbot.ViewFunc
	CallbackCheckQuiz() tgbot.ViewFunc
	CallbackCreateAnswer() tgbot.ViewFunc
	CallbackUserResponse() tgbot.ViewFunc
	CallbackSendQuizToChannel() tgbot.ViewFunc
	CallbackShowChannels() tgbot.ViewFunc
	CallbackAddImage() tgbot.ViewFunc
	CallbackUpdateQuestion() tgbot.ViewFunc
	CallbackCancelUpdate() tgbot.ViewFunc
	CallbackUpdateAnswers() tgbot.ViewFunc
	CallbackGetUserResultExcelFile() tgbot.ViewFunc
	CallbackResetRating() tgbot.ViewFunc
}

type callbackQuiz struct {
	quizService    service.QuizService
	channelService service.ChannelService
	log            *logger.Logger
	store          store.LocalStorage
	excel          *excel.Excel
	tgMsg          customMsg.Message

	mu sync.RWMutex
}

func NewCallbackQuiz(
	quizService service.QuizService,
	channelService service.ChannelService,
	log *logger.Logger,
	store store.LocalStorage,
	tgMsg customMsg.Message,
	excel *excel.Excel,
) (CallbackQuiz, error) {
	if store == nil {
		return nil, errors.New("store is nil")
	}
	if log == nil {
		return nil, errors.New("logger is nil")
	}
	if quizService == nil {
		return nil, errors.New("quizService is nil")
	}
	if tgMsg == nil {
		return nil, errors.New("tgMsg is nil")
	}
	if channelService == nil {
		return nil, errors.New("channelService is nil")
	}
	if excel == nil {
		return nil, errors.New("excel is nil")
	}

	return &callbackQuiz{
		quizService:    quizService,
		channelService: channelService,
		log:            log,
		store:          store,
		tgMsg:          tgMsg,
		excel:          excel,
	}, nil
}

// CallbackShowQuizSetting - quiz_setting
func (c *callbackQuiz) CallbackShowQuizSetting() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {

		text := "Управление ботом"

		if _, err := c.tgMsg.SendEditMessage(update.FromChat().ID,
			update.CallbackQuery.Message.MessageID,
			&markup.QuizSetting,
			text); err != nil {
			return err
		}

		return nil
	}
}

// CallbackCreateQuizQuestion - create_question
func (c *callbackQuiz) CallbackCreateQuizQuestion() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {

		text := "Отправьте вопрос"
		sentMsg, err := c.tgMsg.SendNewMessage(update.FromChat().ID,
			//update.CallbackQuery.Message.MessageID,
			nil,
			text)
		if err != nil {
			return err
		}

		c.store.Set(&store.Data{
			CurrentMsgID:  sentMsg,
			PreferMsgID:   update.CallbackQuery.Message.MessageID,
			OperationType: store.QuizCreate,
		}, update.FromChat().ID)

		return nil
	}
}

// CallbackListQuestion - list_question
func (c *callbackQuiz) CallbackListQuestion() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {

		questionMarkup, err := c.quizService.GetQuestionMarkup(ctx, QuestionGET)
		if err != nil {
			return err
		}

		text := "Список вопросов"
		_, err = c.tgMsg.SendEditMessage(update.FromChat().ID,
			update.CallbackQuery.Message.MessageID,
			questionMarkup,
			text)
		if err != nil {
			return err
		}

		return nil
	}
}

// CallbackDeleteQuestion - delete_question
func (c *callbackQuiz) CallbackDeleteQuestion() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {

		questionMarkup, err := c.quizService.GetQuestionMarkup(ctx, QuestionDELETE)
		if err != nil {
			return err
		}

		text := "Список вопросов"
		_, err = c.tgMsg.SendEditMessage(update.FromChat().ID,
			update.CallbackQuery.Message.MessageID,
			questionMarkup,
			text)
		if err != nil {
			return err
		}

		return nil
	}
}

// CallbackGetQuestion - question_get_{question_id}
func (c *callbackQuiz) CallbackGetQuestion() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		id := GetThirdValue(update.CallbackData())
		if id == 0 {
			c.log.Error("GetThirdValue: failed to get id from  button")
			return customErr.ErrNotFound
		}

		question, err := c.quizService.GetQuestionByID(ctx, id)
		if err != nil {
			c.log.Error("failed to get question by id: %v", err)
			return err
		}

		questionSetting := markup.QuestionSetting(id)
		text := question.QuestionName
		_, err = c.tgMsg.SendEditMessage(update.FromChat().ID,
			update.CallbackQuery.Message.MessageID,
			&questionSetting,
			text)
		if err != nil {
			return err
		}

		return nil
	}
}

// CallbackDeleteByIDQuestion - question_delete_{question_id}
func (c *callbackQuiz) CallbackDeleteByIDQuestion() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		id := GetThirdValue(update.CallbackData())
		if id == 0 {
			c.log.Error("GetThirdValue: failed to get id from  button")
			return customErr.ErrNotFound
		}

		err := c.quizService.DeleteQuestion(ctx, id)
		if err != nil {
			c.log.Error("failed to get question by id: %v", err)
			return err
		}

		text := "Управление ботом"
		if _, err := c.tgMsg.SendEditMessage(update.FromChat().ID,
			update.CallbackQuery.Message.MessageID,
			&markup.QuizSetting,
			text); err != nil {
			return err
		}

		return nil
	}
}

// CallbackCheckQuiz - quiz_check_{question_id}
func (c *callbackQuiz) CallbackCheckQuiz() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		id := GetThirdValue(update.CallbackData())
		if id == 0 {
			c.log.Error("GetThirdValue: failed to get id from  button")
			return customErr.ErrNotFound
		}

		quiz, err := c.quizService.GetQuizByQuestionID(ctx, id)
		if err != nil {
			c.log.Error("failed to get quiz by id: %v", err)
			return err
		}

		if _, err = c.tgMsg.SendMessageToUser(update.FromChat().ID, quiz); err != nil {
			return err
		}

		return nil
	}
}

// CallbackUpdateAnswer - add_answers
func (c *callbackQuiz) CallbackCreateAnswer() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		id := GetThirdValue(update.CallbackData())
		if id == 0 {
			c.log.Error("GetThirdValue: failed to get id from  button")
			return customErr.ErrNotFound
		}

		isAnswerExist, err := c.quizService.IsAnswerExists(ctx, id)
		if err != nil {
			c.log.Error("failed to get quiz by id: %v", err)
			return err
		}

		if isAnswerExist {
			text := "Ответы уже существуют, если что-то нужно изменить, то нажмите на кнопку - Обновить ответы"
			if _, err = c.tgMsg.SendNewMessage(update.FromChat().ID, nil, text); err != nil {
				return err
			}
			return nil
		}
		text := "Отправьте JSON для создания ответов:"
		cancelCommand := markup.CancelCommandQuestion(id)
		_, err = c.tgMsg.SendNewMessage(update.FromChat().ID,
			nil,
			text)
		if err != nil {
			return err
		}

		sentMsg, err := c.tgMsg.SendNewMessage(update.FromChat().ID,
			//update.CallbackQuery.Message.MessageID,
			&cancelCommand,
			jsonExample)
		if err != nil {
			return err
		}

		c.store.Set(&store.Data{
			QuestionID:    id,
			CurrentMsgID:  sentMsg,
			PreferMsgID:   update.CallbackQuery.Message.MessageID,
			OperationType: store.QuizUpdateAnswer,
		}, update.FromChat().ID)

		return nil
	}
}

// CallbackUserResponse -  quiz_answer
func (c *callbackQuiz) CallbackUserResponse() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		id := GetThirdValue(update.CallbackData())
		if id == 0 {
			c.log.Error("GetThirdValue: failed to get id from  button")
			return nil
		}

		isUserAnswerDomain := &entity.IsUserAnswer{AnswerID: id, UserID: update.CallbackQuery.From.ID}

		isAnswerExist, err := c.quizService.IsUserAnswerExists(ctx, isUserAnswerDomain)
		if err != nil {
			c.log.Error("failed to get user answer exists: %v", err)
			return nil
		}

		var text string
		switch isAnswerExist {
		case true:
			text = "На данный вопрос вы уже отвечали!"
		case false:
			costOfResponse, err := c.quizService.UpdateUserResult(ctx, id, update.CallbackQuery.From.ID)
			if err != nil {
				c.log.Error("failed to update user result: %v, update.Callback: %s", err, update.CallbackData())
				return nil
			}

			go func(ctx context.Context, isUserAnswerDomain *entity.IsUserAnswer) {
				if err := c.quizService.CreateBooleanUserAnswer(ctx, isUserAnswerDomain); err != nil {
					c.log.Error("failed to create user answer: %v", err)
				}
			}(ctx, isUserAnswerDomain)

			text = fmt.Sprintf("За выбранный ответ вы получили баллов: %d", costOfResponse)
		}

		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, text)
		if _, err := bot.Request(callback); err != nil {
			c.log.Error("failed to send callback message: %v", err)
			return nil
		}

		return nil
	}
}

// CallbackShowChannels - show_channels_{questionID}
func (c *callbackQuiz) CallbackShowChannels() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		id := GetThirdValue(update.CallbackData())
		if id == 0 {
			c.log.Error("GetThirdValue: failed to get id from  button")
			return nil
		}

		channelsButton, err := c.channelService.GetAllAdminChannel(ctx, id)
		if err != nil {
			c.log.Error("failed to get all admin channels: %v", err)
			return err
		}

		text := "Выберите канал для отправки вопроса"
		_, err = c.tgMsg.SendEditMessage(update.FromChat().ID,
			update.CallbackQuery.Message.MessageID,
			channelsButton,
			text)
		if err != nil {
			c.log.Error("failed to send update result message: %v", err)
			return err
		}

		return nil
	}
}

// CallbackSendQuizToChannel - channel_get_{URLQuery}
func (c *callbackQuiz) CallbackSendQuizToChannel() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		query := GetThirdValueString(update.CallbackData())
		if query == "" {
			c.log.Error("GetThirdValue: failed to get id from  button")
			return nil
		}

		urlValues, err := url.ParseQuery(query)
		if err != nil {
			c.log.Error("failed to parse url query: %v", err)
			return err
		}

		questionID, err := strconv.Atoi(urlValues.Get("questionID"))
		if err != nil {
			c.log.Error("failed to parse questionID: %v", err)
			return err
		}
		channelID, err := strconv.Atoi(urlValues.Get("channelID"))
		if err != nil {
			c.log.Error("failed to parse channelID: %v", err)
			return err
		}

		quiz, err := c.quizService.GetQuizByQuestionID(ctx, questionID)
		if err != nil {
			c.log.Error("failed to get quiz by id: %v", err)
			return err
		}

		channel, err := c.channelService.GetByID(ctx, channelID)
		if err != nil {
			c.log.Error("failed to get channel by id: %v", err)
			return err
		}

		if _, err = c.tgMsg.SendMessageToUser(channel.TgID, quiz); err != nil {
			return err
		}

		if err = c.quizService.SetSendStatus(ctx, questionID); err != nil {
			c.log.Error("failed to set quiz status: %v", err)
			return err
		}

		return nil
	}
}

// CallbackAddImage - add_image_{question_id}
func (c *callbackQuiz) CallbackAddImage() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		id := GetThirdValue(update.CallbackData())
		if id == 0 {
			c.log.Error("GetThirdValue: failed to get id from  button")
			return customErr.ErrNotFound
		}

		text := "Отправьте изображение"
		cancelCommand := markup.CancelCommandQuestion(id)
		sentMsg, err := c.tgMsg.SendEditMessage(update.FromChat().ID,
			update.CallbackQuery.Message.MessageID,
			&cancelCommand,
			text)
		if err != nil {
			return err
		}

		c.store.Set(&store.Data{
			QuestionID:    id,
			CurrentMsgID:  sentMsg,
			PreferMsgID:   update.CallbackQuery.Message.MessageID,
			OperationType: store.QuizUpdateImage,
		}, update.FromChat().ID)

		return nil
	}
}

// CallbackUpdateQuestion - update_question_{question_id}
func (c *callbackQuiz) CallbackUpdateQuestion() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		id := GetThirdValue(update.CallbackData())
		if id == 0 {
			c.log.Error("GetThirdValue: failed to get id from  button")
			return customErr.ErrNotFound
		}

		text := "Отправьте новый вопрос"
		cancelCommand := markup.CancelCommandQuestion(id)
		sentMsg, err := c.tgMsg.SendNewMessage(update.FromChat().ID,
			//update.CallbackQuery.Message.MessageID,
			&cancelCommand,
			text)
		if err != nil {
			return err
		}

		c.store.Set(&store.Data{
			QuestionID:    id,
			CurrentMsgID:  sentMsg,
			PreferMsgID:   update.CallbackQuery.Message.MessageID,
			OperationType: store.QuizUpdateQuestion,
		}, update.FromChat().ID)

		return nil
	}
}

// CallbackCancelUpdate - cancel_update_{question_id}
func (c *callbackQuiz) CallbackCancelUpdate() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		id := GetThirdValue(update.CallbackData())
		if id == 0 {
			c.log.Error("GetThirdValue: failed to get id from  button")
			return customErr.ErrNotFound
		}

		c.store.Delete(update.FromChat().ID)

		question, err := c.quizService.GetQuestionByID(ctx, id)
		if err != nil {
			c.log.Error("failed to get question by id: %v", err)
			return err
		}

		questionSetting := markup.QuestionSetting(id)
		text := question.QuestionName
		_, err = c.tgMsg.SendEditMessage(update.FromChat().ID,
			update.CallbackQuery.Message.MessageID,
			&questionSetting,
			text)
		if err != nil {
			return err
		}

		return nil
	}
}

// CallbackUpdateAnswers - update_answers_{question_id}
func (c *callbackQuiz) CallbackUpdateAnswers() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		id := GetThirdValue(update.CallbackData())
		if id == 0 {
			c.log.Error("GetThirdValue: failed to get id from  button")
			return customErr.ErrNotFound
		}

		quiz, err := c.quizService.GetQuizByQuestionID(ctx, id)
		if err != nil {
			c.log.Error("failed to get quiz by id: %v", err)
			return err
		}

		bytesArgs, err := json.MarshalIndent(AnswerToArgsModel(quiz.Answer), "", " ")
		if err != nil {
			c.log.Error("failed to marshal args: %v", err)
			return err
		}

		if _, err := c.tgMsg.SendNewMessage(update.FromChat().ID, nil, string(bytesArgs)); err != nil {
			return err
		}

		text := "Отправьте новые ответы"
		cancelCommand := markup.CancelCommandQuestion(id)
		sentMsg, err := c.tgMsg.SendNewMessage(update.FromChat().ID,
			//update.CallbackQuery.Message.MessageID,
			&cancelCommand,
			text)
		if err != nil {
			return err
		}

		c.store.Set(&store.Data{
			QuestionID:    id,
			CurrentMsgID:  sentMsg,
			PreferMsgID:   update.CallbackQuery.Message.MessageID,
			OperationType: store.QuizUpdateOldAnswer,
		}, update.FromChat().ID)

		return nil
	}
}

// CallbackGetUserResultExcelFile - downloading_rating
func (c *callbackQuiz) CallbackGetUserResultExcelFile() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		userResult, err := c.quizService.GetAllUserResults(ctx)
		if err != nil {
			c.log.Error("quizService.GetAllUserResults: failed to get contest: %v", err)
			return err
		}

		c.mu.Lock()
		fileName, err := c.excel.GenerateUserResultsExcelFile(userResult, update.CallbackQuery.From.UserName)
		if err != nil {
			c.log.Error("Excel.GenerateExcelFile: failed to generate excel file: %v", err)
			return err
		}

		fileIDBytes, err := c.excel.GetExcelFile(fileName)
		if err != nil {
			c.log.Error("Excel.GetExcelFile: failed to get excel file: %v", err)
			return err
		}
		c.mu.Unlock()

		if fileIDBytes == nil {
			c.log.Error("fileIDBytes is nil")
			return errors.New("nil file")
		}

		if _, err := c.tgMsg.SendDocument(update.FromChat().ID,
			fileName,
			fileIDBytes,
			"Рейтинг пользователей",
		); err != nil {
			return err
		}

		return nil
	}
}

// CallbackResetRating - reset_rating
func (c *callbackQuiz) CallbackResetRating() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {

		err := c.quizService.ResetAllUserResult(ctx)
		if err != nil {
			c.log.Error("failed to reset user result: %v", err)
			return err
		}

		go func(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
			newCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			if err = c.CallbackGetUserResultExcelFile()(newCtx, bot, update); err != nil {
				c.log.Error("failed to reset user result: %v", err)

			}
		}(bot, update)

		return nil
	}
}
