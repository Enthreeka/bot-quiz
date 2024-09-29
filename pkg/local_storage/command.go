package store

type OperationType string

type TypeCommand string

const (
	Admin OperationType = "admin"
	Quiz  OperationType = "quiz"
)

const (
	AdminCreate TypeCommand = "create"
	AdminDelete TypeCommand = "delete"
)

const (
	QuizCreate          TypeCommand = "create_quiz"
	QuizUpdateAnswer    TypeCommand = "update_answer"
	QuizUpdateImage     TypeCommand = "update_image"
	QuizUpdateQuestion  TypeCommand = "update_question"
	QuizUpdateOldAnswer TypeCommand = "update_old_answer"
)

var MapTypes = map[TypeCommand]OperationType{
	AdminCreate:      Admin,
	AdminDelete:      Admin,
	QuizCreate:       Quiz,
	QuizUpdateAnswer: Quiz,
	QuizUpdateImage:  Quiz,
}
