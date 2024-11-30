package repo

import (
	"context"
	"errors"
	"github.com/Enthreeka/tg-bot-quiz/internal/entity"
	"github.com/Enthreeka/tg-bot-quiz/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type QuizRepo interface {
	CreateQuestion(ctx context.Context, tx pgx.Tx, question *entity.Question) (int, error)
	GetAllQuestionsByChannelID(ctx context.Context, channelID int64) ([]entity.Question, error)
	GetQuestionByID(ctx context.Context, id int) (*entity.Question, error)
	UpdateQuestion(ctx context.Context, questionID int, question string) error
	DeleteQuestion(ctx context.Context, id int) error
	UpdateImage(ctx context.Context, questionID int, image string) error
	SetSendStatus(ctx context.Context, id int) error
	GetChannelTgIDByQuestionID(ctx context.Context, questionID int) (int, error)

	CreateAnswers(ctx context.Context, tx pgx.Tx, answers []entity.Answer, questionID int) ([]int, error)
	GetAnswerByID(ctx context.Context, id int) (int, int, error)
	GetQuizByQuestionID(ctx context.Context, id int) (*entity.Quiz, error)
	UpdateAnswer(ctx context.Context, answer *entity.Answer) error
	DeleteAnswer(ctx context.Context, tx pgx.Tx, id int) error
	IsAnswerExists(ctx context.Context, questionID int) (bool, error)
	DeleteAndInsertNewAnswers(ctx context.Context, answers []entity.Answer, questionID int) error

	CreateUserResult(ctx context.Context, userResult *entity.UserResult) error
	GetAllUserResultsByChannelID(ctx context.Context, channelID int) ([]entity.UserResult, error)
	ResetAllUserResult(ctx context.Context, channelTgID int) error

	CreateBooleanUserAnswer(ctx context.Context, answer *entity.IsUserAnswer) error
	IsUserAnswerExists(ctx context.Context, userAnswer *entity.IsUserAnswer) (bool, error)
}

type quizRepo struct {
	*postgres.Postgres
}

func NewQuizRepo(pg *postgres.Postgres) (QuizRepo, error) {
	if pg == nil {
		return nil, errors.New("nil postgres")
	}
	return &quizRepo{
		Postgres: pg,
	}, nil
}

// Question domain

func (q *quizRepo) GetChannelTgIDByQuestionID(ctx context.Context, questionID int) (int, error) {
	query := `SELECT channel_tg_id FROM questions WHERE id = $1`
	var (
		channelTgId int
	)

	err := q.Pool.QueryRow(ctx, query, questionID).Scan(&channelTgId)
	return channelTgId, err
}

func (q *quizRepo) SetSendStatus(ctx context.Context, id int) error {
	query := `UPDATE questions SET is_send = true WHERE id = $1;`

	_, err := q.Pool.Exec(ctx, query, id)
	return err
}

func (q *quizRepo) CreateQuestion(ctx context.Context, tx pgx.Tx, question *entity.Question) (int, error) {
	query := `INSERT INTO questions (created_by_user, question_name, file_id, channel_tg_id) VALUES ($1, $2, $3, $4) RETURNING id`

	var err error
	var id int
	if tx == nil {
		err = q.Pool.QueryRow(ctx, query, question.CreatedByUser, question.QuestionName, question.FileID, question.ChannelID).Scan(&id)
	} else {
		err = tx.QueryRow(ctx, query, question.CreatedByUser, question.QuestionName, question.FileID, question.ChannelID).Scan(&id)
	}

	return id, err
}

func (q *quizRepo) GetAllQuestionsByChannelID(ctx context.Context, channelID int64) ([]entity.Question, error) {
	query := `SELECT 
    id,
    created_by_user,
    created_at,
    question_name,
    file_id,
    deadline,
    is_send
	FROM questions
	WHERE channel_tg_id = $1`

	rows, err := q.Pool.Query(ctx, query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []entity.Question
	for rows.Next() {
		var question entity.Question
		err := rows.Scan(&question.ID,
			&question.CreatedByUser,
			&question.CreatedAt,
			&question.QuestionName,
			&question.FileID,
			&question.Deadline,
			&question.IsSend,
		)
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}

func (q *quizRepo) UpdateImage(ctx context.Context, questionID int, image string) error {
	query := `UPDATE questions SET file_id = $1 WHERE id= $2`

	_, err := q.Pool.Exec(ctx, query, image, questionID)
	return err
}

func (q *quizRepo) GetQuestionByID(ctx context.Context, id int) (*entity.Question, error) {
	query := `SELECT 
    id,
    created_by_user,
    created_at,
    question_name,
    file_id,
    deadline,
    is_send,
    channel_tg_id
	FROM questions
	WHERE id = $1`
	question := new(entity.Question)

	err := q.Pool.QueryRow(ctx, query, id).Scan(&question.ID,
		&question.CreatedByUser,
		&question.CreatedAt,
		&question.QuestionName,
		&question.FileID,
		&question.Deadline,
		&question.IsSend,
		&question.ChannelID,
	)
	return question, err
}

func (q *quizRepo) UpdateQuestion(ctx context.Context, questionID int, question string) error {
	query := `UPDATE questions SET question_name = $1 WHERE id = $2`

	_, err := q.Pool.Exec(ctx, query, question, questionID)
	return err
}

func (q *quizRepo) DeleteQuestion(ctx context.Context, id int) error {
	query := `DELETE FROM questions WHERE id = $1`

	_, err := q.Pool.Exec(ctx, query, id)
	return err
}

// Answer domain

func (q *quizRepo) IsAnswerExists(ctx context.Context, questionID int) (bool, error) {
	query := `select exists (select id from answers where question_id = $1)`
	var isExist bool

	err := q.Pool.QueryRow(ctx, query, questionID).Scan(&isExist)
	if checkErr := ErrorHandler(err); checkErr != nil {
		return isExist, checkErr
	}

	return isExist, err
}

func (q *quizRepo) DeleteAndInsertNewAnswers(ctx context.Context, answers []entity.Answer, questionID int) error {
	tx, err := q.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	if err = q.DeleteAnswer(ctx, tx, questionID); err != nil {
		return err
	}

	if _, err = q.CreateAnswers(ctx, tx, answers, questionID); err != nil {
		return err
	}

	return err
}

func (q *quizRepo) CreateAnswers(ctx context.Context, tx pgx.Tx, answers []entity.Answer, questionID int) ([]int, error) {
	query := `INSERT INTO answers (answer, cost_of_response,question_id) VALUES ($1, $2,$3) RETURNING id`
	var newID []int

	for _, value := range answers {
		var id int
		var err error

		if tx != nil {
			err = tx.QueryRow(ctx, query, value.Answer, value.CostOfResponse, questionID).Scan(&id)
		} else {
			err = q.Pool.QueryRow(ctx, query, value.Answer, value.CostOfResponse, questionID).Scan(&id)

		}

		if err != nil {
			return nil, err
		}
		newID = append(newID, id)
	}

	return newID, nil
}

func (q *quizRepo) GetAnswerByID(ctx context.Context, id int) (int, int, error) {
	query := `SELECT cost_of_response,question_id from answers WHERE id = $1`
	var (
		costOfResponse int
		questionID     int
	)

	err := q.Pool.QueryRow(ctx, query, id).Scan(&costOfResponse, &questionID)
	return costOfResponse, questionID, err
}

func (q *quizRepo) GetQuizByQuestionID(ctx context.Context, id int) (*entity.Quiz, error) {
	queryQuestion := `SELECT question_name, file_id, channel_tg_id FROM questions WHERE id = $1`

	queryAnswer := `SELECT a.id, a.answer, a.cost_of_response FROM answers a
					JOIN questions q ON q.id = a.question_id
								WHERE a.question_id = $1`

	tx, err := q.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()
	qu := new(entity.Quiz)

	if err = tx.QueryRow(ctx, queryQuestion, id).Scan(
		&qu.Question.QuestionName,
		&qu.Question.FileID,
		&qu.Question.ChannelID,
	); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, queryAnswer, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.Answer
	for rows.Next() {
		var result entity.Answer
		err := rows.Scan(&result.ID, &result.Answer, &result.CostOfResponse)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	qu.Answer = results
	qu.Question.ID = id

	return qu, nil
}

func (q *quizRepo) UpdateAnswer(ctx context.Context, answer *entity.Answer) error {
	query := `UPDATE answers SET answer = $1, cost_of_response = $2 WHERE id = $3`

	_, err := q.Pool.Exec(ctx, query, answer.Answer, answer.CostOfResponse, answer.ID)
	return err
}

func (q *quizRepo) DeleteAnswer(ctx context.Context, tx pgx.Tx, questionID int) error {
	query := `DELETE FROM answers WHERE question_id = $1`

	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, questionID)
	} else {
		_, err = q.Pool.Exec(ctx, query, questionID)
	}

	return err
}

// User result entity

func (q *quizRepo) CreateUserResult(ctx context.Context, userResult *entity.UserResult) error {
	//query := `INSERT INTO user_results (user_id,total_points) VALUES ($1, $2)
	//		ON CONFLICT (user_id) DO UPDATE SET
	//			total_points = user_results.total_points + $2`

	query := `INSERT INTO user_results (user_id,points,questions_id) VALUES ($1, $2, $3)`

	_, err := q.Pool.Exec(ctx, query, userResult.UserID, userResult.Points, userResult.QuestionID)
	return err
}

func (q *quizRepo) GetAllUserResultsByChannelID(ctx context.Context, channelID int) ([]entity.UserResult, error) {
	query := `SELECT
    u.tg_username,
    user_results.user_id,
    user_results.id,
    user_results.points,
	q.question_name
	FROM user_results
			 JOIN "user" u
				  ON u.id = user_results.user_id
			 JOIN questions q on user_results.questions_id = q.id
			 JOIN channel c on q.channel_tg_id = c.tg_id
			WHERE c.tg_id = $1;`

	rows, err := q.Pool.Query(ctx, query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.UserResult
	for rows.Next() {
		var result entity.UserResult
		err := rows.Scan(&result.TGUsername, &result.UserID, &result.ID, &result.Points, &result.QuestionName)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (q *quizRepo) ResetAllUserResult(ctx context.Context, channelTgID int) error {
	query := `UPDATE user_results
			SET points = 0
			FROM questions q
					 JOIN channel c ON q.channel_tg_id = c.tg_id
			WHERE user_results.questions_id = q.id AND c.tg_id = $1;`

	_, err := q.Pool.Exec(ctx, query, channelTgID)
	return err
}

// Is user answer domain

func (q *quizRepo) CreateBooleanUserAnswer(ctx context.Context, answer *entity.IsUserAnswer) error {
	query := `INSERT INTO is_user_answer (user_id,is_answer,question_id) VALUES ($1, true, (select question_id from answers where id = $2));`

	_, err := q.Pool.Exec(ctx, query, answer.UserID, answer.AnswerID)
	return err
}

func (q *quizRepo) IsUserAnswerExists(ctx context.Context, userAnswer *entity.IsUserAnswer) (bool, error) {
	query := `SELECT EXISTS (SELECT user_id from is_user_answer WHERE user_id = $1 AND
         question_id = (select question_id from answers where id = $2) AND is_answer = true)`
	var isExist bool

	err := q.Pool.QueryRow(ctx, query, userAnswer.UserID, userAnswer.AnswerID).Scan(&isExist)
	if checkErr := ErrorHandler(err); checkErr != nil {
		return isExist, checkErr
	}

	return isExist, err
}
