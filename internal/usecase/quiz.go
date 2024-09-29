package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Enthreeka/tg-bot-quiz/internal/entity"
	"github.com/Enthreeka/tg-bot-quiz/internal/repo"
	"github.com/Enthreeka/tg-bot-quiz/pkg/logger"
	"github.com/Enthreeka/tg-bot-quiz/pkg/serialize"
	"github.com/Enthreeka/tg-bot-quiz/pkg/tg_bot_api/button"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"unicode/utf8"
)

type QuizService interface {
	CreateQuestion(ctx context.Context, tx pgx.Tx, question *entity.Question) (int, error)
	GetQuestionByID(ctx context.Context, id int) (*entity.Question, error)
	UpdateQuestion(ctx context.Context, questionID int, question string) error
	DeleteQuestion(ctx context.Context, id int) error
	GetQuestionMarkup(ctx context.Context, method string) (*tgbotapi.InlineKeyboardMarkup, error)
	UpdateImage(ctx context.Context, questionID int, image string) error
	SetSendStatus(ctx context.Context, id int) error

	GetQuizByQuestionID(ctx context.Context, id int) (*entity.Quiz, error)
	IsAnswerExists(ctx context.Context, questionID int) (bool, error)
	QuizUpdateOldAnswer(ctx context.Context, text string, questionID int) error

	UpdateUserResult(ctx context.Context, answerID int, userID int64) (int, error)
	CreateUserResult(ctx context.Context, userResult *entity.UserResult) error
	GetAllUserResults(ctx context.Context) ([]entity.UserResult, error)
	ResetAllUserResult(ctx context.Context) error

	CreateBooleanUserAnswer(ctx context.Context, answer *entity.IsUserAnswer) error
	IsUserAnswerExists(ctx context.Context, userAnswer *entity.IsUserAnswer) (bool, error)

	QuizUpdateAnswer(ctx context.Context, text string, questionID int) error
}

type quizService struct {
	quizRepo repo.QuizRepo
	log      *logger.Logger
}

func NewQuizService(quizRepo repo.QuizRepo, log *logger.Logger) (QuizService, error) {
	if quizRepo == nil {
		return nil, errors.New("nil quizRepo")
	}
	if log == nil {
		return nil, errors.New("nil logger")
	}

	return &quizService{
		quizRepo: quizRepo,
		log:      log,
	}, nil
}

func (q *quizService) SetSendStatus(ctx context.Context, id int) error {
	return q.quizRepo.SetSendStatus(ctx, id)
}

func (q *quizService) IsAnswerExists(ctx context.Context, questionID int) (bool, error) {
	return q.quizRepo.IsAnswerExists(ctx, questionID)
}

func (q *quizService) UpdateImage(ctx context.Context, questionID int, image string) error {
	return q.quizRepo.UpdateImage(ctx, questionID, image)
}

func (q *quizService) GetQuestionByID(ctx context.Context, id int) (*entity.Question, error) {
	return q.quizRepo.GetQuestionByID(ctx, id)
}

func (q *quizService) UpdateQuestion(ctx context.Context, questionID int, question string) error {
	return q.quizRepo.UpdateQuestion(ctx, questionID, question)
}

func (q *quizService) DeleteQuestion(ctx context.Context, id int) error {
	return q.quizRepo.DeleteQuestion(ctx, id)
}

func (q *quizService) GetQuizByQuestionID(ctx context.Context, id int) (*entity.Quiz, error) {
	return q.quizRepo.GetQuizByQuestionID(ctx, id)
}

func (q *quizService) CreateUserResult(ctx context.Context, userResult *entity.UserResult) error {
	return q.quizRepo.CreateUserResult(ctx, userResult)
}

func (q *quizService) GetAllUserResults(ctx context.Context) ([]entity.UserResult, error) {
	return q.quizRepo.GetAllUserResults(ctx)
}

func (q *quizService) ResetAllUserResult(ctx context.Context) error {
	return q.quizRepo.ResetAllUserResult(ctx)
}

func (q *quizService) CreateQuestion(ctx context.Context, tx pgx.Tx, question *entity.Question) (int, error) {
	return q.quizRepo.CreateQuestion(ctx, tx, question)
}

func (q *quizService) GetQuestionMarkup(ctx context.Context, method string) (*tgbotapi.InlineKeyboardMarkup, error) {
	questions, err := q.quizRepo.GetAllQuestions(ctx)
	if err != nil {
		q.log.Error("failed to get question markup: %v", err)
		return nil, err
	}
	return q.createQuestionMarkup(questions, method)
}

func (q *quizService) UpdateUserResult(ctx context.Context, answerID int, userID int64) (int, error) {
	costOfResponse, err := q.quizRepo.GetAnswerByID(ctx, answerID)
	if err != nil {
		q.log.Error("failed to get answer: %v", err)
		return 0, err
	}

	if err := q.quizRepo.CreateUserResult(ctx, &entity.UserResult{
		UserID:      userID,
		TotalPoints: costOfResponse,
	}); err != nil {
		q.log.Error("failed to create user result: %v", err)
		return 0, err
	}

	return costOfResponse, nil
}

func (q *quizService) CreateBooleanUserAnswer(ctx context.Context, answer *entity.IsUserAnswer) error {
	return q.quizRepo.CreateBooleanUserAnswer(ctx, answer)
}

func (q *quizService) IsUserAnswerExists(ctx context.Context, userAnswer *entity.IsUserAnswer) (bool, error) {
	return q.quizRepo.IsUserAnswerExists(ctx, userAnswer)
}

func (q *quizService) QuizUpdateAnswer(ctx context.Context, text string, questionID int) error {
	args, err := serialize.ParseJSON[entity.Args](text)
	if err != nil {
		q.log.Error("ParseJSON: %v", err)
		return err
	}

	if _, err := q.quizRepo.CreateAnswers(ctx, nil, updateArgsToModel(args), questionID); err != nil {
		q.log.Error("isStoreExist::store.QuizCreate:CreateAnswers: %v", err)
		return err
	}

	return nil
}

func (q *quizService) QuizUpdateOldAnswer(ctx context.Context, text string, questionID int) error {
	args, err := serialize.ParseJSON[entity.Args](text)
	if err != nil {
		q.log.Error("ParseJSON: %v", err)
		return err
	}

	if err := q.quizRepo.DeleteAndInsertNewAnswers(ctx, updateArgsToModel(args), questionID); err != nil {
		q.log.Error("isStoreExist::store.QuizCreate:CreateAnswers: %v", err)
		return err
	}

	return nil
}

func (q *quizService) createQuestionMarkup(questions []entity.Question, method string) (*tgbotapi.InlineKeyboardMarkup, error) {
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	buttonsPerRow := 1

	var isSendStr string
	for i, el := range questions {
		if el.IsSend == true {
			isSendStr = "Отправлено"
		} else {
			isSendStr = "Не отправлено"
		}

		name := el.QuestionName
		if utf8.RuneCountInString(el.QuestionName) > 10 {
			name = el.QuestionName[:10]
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s - [%s]", name, isSendStr),
			fmt.Sprintf("question_%s_%d", method, el.ID))

		row = append(row, btn)

		if (i+1)%buttonsPerRow == 0 || i == len(questions)-1 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{button.MainMenuButton})
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return &markup, nil
}
