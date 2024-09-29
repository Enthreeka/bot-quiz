package bot_error

import (
	"fmt"
)

const (
	InvalidRequest      = "Invalid Request"
	ServerError         = "Internal Server Error"
	NotFound            = "Not Found"
	NoRows              = "No Rows"
	ForeignKeyViolation = "Foreign Key Violation"
	UniqueViolation     = "Violation Must Be Unique"
	AdminPermission     = "Permission Denied"
)

var (
	ErrInvalidRequest      = NewError(InvalidRequest)
	ErrServerError         = NewError(ServerError)
	ErrNotFound            = NewError(NotFound)
	ErrNoRows              = NewError(NoRows)
	ErrForeignKeyViolation = NewError(ForeignKeyViolation)
	ErrUniqueViolation     = NewError(UniqueViolation)
	ErrIsNotAdmin          = NewError(AdminPermission)
)

type ErrorCode string

type BotError struct {
	Err ErrorCode `json:"error"`
	Msg string    `json:"message"`
}

func (a *BotError) Error() string {
	return fmt.Sprintf("%s", a.Msg)
}

func NewError(err ErrorCode) *BotError {
	return &BotError{
		Err: err,
		Msg: parseErrToText(err),
	}
}

func parseErrToText(err ErrorCode) string {
	switch err {
	case InvalidRequest:
		return "Некорректный запрос"
	case NotFound:
		return "Поисковая сущность отсутствует"
	case AdminPermission:
		return "Недостаточно прав доступа"
	case NoRows, ForeignKeyViolation, UniqueViolation:
		return "Ошибка связанная с базой данных"
	default:
		return "Произошла внутрення ошибка на сервере"
	}

}