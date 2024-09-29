package markup

import (
	"fmt"
	"github.com/Enthreeka/tg-bot-quiz/pkg/tg_bot_api/button"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	StartMenu = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Управление ботом", "quiz_setting")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Управление пользователями", "user_setting")),
	)

	UserSetting = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назначить роль администратора", "admin_set_role"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отозвать роль администратора", "admin_delete_role"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Посмотреть список администраторов", "admin_look_up"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Скачать рейтинг", "downloading_rating")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Обнулить рейтинг", "reset_rating")),
		tgbotapi.NewInlineKeyboardRow(button.MainMenuButton),
	)

	QuizSetting = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Создать вопрос", "create_question")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Открыть список вопросов", "list_question")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалить вопрос", "delete_question")),
	)

	MainMenu = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button.MainMenuButton))
)

func QuestionSetting(questionID int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Обновить вопрос", fmt.Sprintf("update_question_%d", questionID))),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить ответы", fmt.Sprintf("add_answers_%d", questionID))),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Обновить ответы", fmt.Sprintf("update_answers_%d", questionID))),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить изображение", fmt.Sprintf("add_image_%d", questionID))),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Предварительный просмотр", fmt.Sprintf("quiz_check_%d", questionID))),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Выбрать канал и отправить вопрос", fmt.Sprintf("show_channels_%d", questionID))),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вернуться назад", "quiz_setting")),
	)
}

func CancelCommandQuestion(questionID int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Отмена выполнения", fmt.Sprintf("cancel_update_%d", questionID))))
}
