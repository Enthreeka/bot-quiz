package entity

import "time"

type Question struct {
	ID            int        `json:"id"`
	CreatedByUser int64      `json:"created_by_user"`
	CreatedAt     time.Time  `json:"created_at"`
	QuestionName  string     `json:"question_name"`
	Deadline      *time.Time `json:"deadline"`
	FileID        *string    `json:"file_id"`
	IsSend        bool       `json:"is_send"`
}

type Answer struct {
	ID             int    `json:"id"`
	Answer         string `json:"answer"`
	CostOfResponse int    `json:"cost_of_response"`
	QuestionID     int    `json:"question_id"`
}

type QuestionsAnswers struct {
	QuestionID int `json:"questions_id"`
	AnswerID   int `json:"answers_id"`
}

type UserResult struct {
	ID          int   `json:"id"`
	UserID      int64 `json:"user_id"`
	TotalPoints int   `json:"total_points"`

	TGUsername string `json:"tg_username"`
}

type Quiz struct {
	Question Question `json:"question"`
	Answer   []Answer `json:"answer"`
}

type IsUserAnswer struct {
	UserID   int64 `json:"user_id"`
	AnswerID int   `json:"answers_id"`
}