package bot

import (
	"context"
	"github.com/Enthreeka/tg-bot-quiz/internal/config"
	"github.com/Enthreeka/tg-bot-quiz/internal/handler/callback"
	"github.com/Enthreeka/tg-bot-quiz/internal/handler/middleware"
	"github.com/Enthreeka/tg-bot-quiz/internal/handler/tgbot"
	"github.com/Enthreeka/tg-bot-quiz/internal/handler/view"
	"github.com/Enthreeka/tg-bot-quiz/internal/repo"
	service "github.com/Enthreeka/tg-bot-quiz/internal/usecase"
	"github.com/Enthreeka/tg-bot-quiz/pkg/excel"
	store "github.com/Enthreeka/tg-bot-quiz/pkg/local_storage"
	"github.com/Enthreeka/tg-bot-quiz/pkg/logger"
	"github.com/Enthreeka/tg-bot-quiz/pkg/postgres"
	customMsg "github.com/Enthreeka/tg-bot-quiz/pkg/tg_bot_api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

const (
	PostgresMaxAttempts = 5
)

type Bot struct {
	bot           *tgbotapi.BotAPI
	psql          *postgres.Postgres
	store         *store.Store
	cfg           *config.Config
	log           *logger.Logger
	excel         *excel.Excel
	tgMsg         *customMsg.TelegramMsg
	callbackStore *store.CallbackStorage

	userService    service.UserService
	channelService service.ChannelService
	quizService    service.QuizService

	userRepo    repo.UserRepo
	channelRepo repo.ChannelRepo
	quizRepo    repo.QuizRepo

	callbackQuiz callback.CallbackQuiz
	callbackUser callback.CallbackUser
	viewGeneral  *view.ViewGeneral
}

func NewBot() *Bot {
	return &Bot{}
}

func (b *Bot) initHandler() {
	b.viewGeneral = view.NewViewGeneral(b.log, b.tgMsg)

	callbackUser, err := callback.NewCallbackUser(b.userService, b.log, b.store, b.tgMsg)
	if err != nil {
		log.Fatal(err)
	}
	b.callbackUser = callbackUser

	callbackQuiz, err := callback.NewCallbackQuiz(b.quizService, b.channelService, b.log, b.store, b.tgMsg, b.excel)
	if err != nil {
		log.Fatal(err)
	}
	b.callbackQuiz = callbackQuiz

	b.log.Info("Initializing handler")
}

func (b *Bot) initUsecase() {
	userService, err := service.NewUserService(b.userRepo, b.log)
	if err != nil {
		b.log.Fatal("Failed to initialize user service")
	}
	b.userService = userService

	channelService, err := service.NewChannelService(b.channelRepo, b.log)
	if err != nil {
		b.log.Fatal("NewChannelService:", err)
	}
	b.channelService = channelService

	quizService, err := service.NewQuizService(b.quizRepo, b.log)
	if err != nil {
		b.log.Fatal("NewQuizService:", err)
	}
	b.quizService = quizService

	b.log.Info("Initializing usecase")
}

func (b *Bot) initRepo() {
	userRepo, err := repo.NewUserRepo(b.psql)
	if err != nil {
		b.log.Fatal("Failed to initialize user repo")
	}
	b.userRepo = userRepo

	channelRepo, err := repo.NewChannelRepo(b.psql)
	if err != nil {
		b.log.Fatal("NewChannelRepo: ", err)
	}
	b.channelRepo = channelRepo

	quizRepo, err := repo.NewQuizRepo(b.psql)
	if err != nil {
		b.log.Fatal("NewQuizRepo: ", err)
	}
	b.quizRepo = quizRepo

	b.log.Info("Initializing repo")
}

func (b *Bot) initMessage() {
	b.tgMsg = customMsg.NewMessageSetting(b.bot, b.log)

	b.log.Info("Initializing message")
}

func (b *Bot) initPostgres(ctx context.Context) {
	psql, err := postgres.New(ctx, PostgresMaxAttempts, b.cfg.Postgres.URL)
	if err != nil {
		b.log.Fatal("failed to connect PostgreSQL: %v", err)
	}
	b.psql = psql

	b.log.Info("Initializing postgres")
}

func (b *Bot) initConfig() {
	cfg, err := config.New()
	if err != nil {
		b.log.Fatal("failed load config: %v", err)
	}
	b.cfg = cfg

	b.log.Info("Initializing config")
}

func (b *Bot) initLogger() {
	b.log = logger.New()

	b.log.Info("Initializing logger")
}

func (b *Bot) initStore() {
	b.store = store.NewStore()

	b.log.Info("Initializing store")
}

func (b *Bot) initCallbackStorage() {
	b.callbackStore = store.NewCallbackStorage()

	b.log.Info("Initializing callback storage")
}

func (b *Bot) initTelegramBot() {
	bot, err := tgbotapi.NewBotAPI(b.cfg.Telegram.Token)
	if err != nil {
		b.log.Fatal("failed to load token %v", err)
	}
	bot.Debug = false
	b.bot = bot

	b.log.Info("Initializing telegram bot")
	b.log.Info("Authorized on account %s", bot.Self.UserName)
}

func (b *Bot) initExcel() {
	b.excel = excel.NewExcel(b.log)
}

func (b *Bot) initialize(ctx context.Context) {
	b.initLogger()
	b.initExcel()
	b.initConfig()
	b.initTelegramBot()
	b.initStore()
	b.initCallbackStorage()
	b.initPostgres(ctx)
	b.initMessage()
	b.initRepo()
	b.initUsecase()
	b.initHandler()
}

func (b *Bot) Run(ctx context.Context) {
	startBot := time.Now()
	b.initialize(ctx)
	newBot, err := tgbot.NewBot(b.bot, b.log, b.store, b.tgMsg, b.userService, b.quizService, b.callbackStore, b.channelService)
	if err != nil {
		b.log.Fatal("failed go create new bot: ", err)
	}
	defer b.psql.Close()

	newBot.RegisterCommandView("start", middleware.AdminMiddleware(b.userService, b.viewGeneral.CallbackStartAdminPanel()))

	// callback user domain
	newBot.RegisterCommandCallback("main_menu", middleware.AdminMiddleware(b.userService, b.callbackUser.MainMenu()))
	newBot.RegisterCommandCallback("user_setting", middleware.AdminMiddleware(b.userService, b.callbackUser.AdminRoleSetting()))
	newBot.RegisterCommandCallback("admin_look_up", middleware.AdminMiddleware(b.userService, b.callbackUser.AdminLookUp()))
	newBot.RegisterCommandCallback("admin_delete_role", middleware.AdminMiddleware(b.userService, b.callbackUser.AdminDeleteRole()))
	newBot.RegisterCommandCallback("admin_set_role", middleware.AdminMiddleware(b.userService, b.callbackUser.AdminSetRole()))

	// callback quiz domain
	newBot.RegisterCommandCallback("quiz_setting", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackShowQuizSetting()))
	newBot.RegisterCommandCallback("create_question", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackCreateQuizQuestion()))
	newBot.RegisterCommandCallback("list_question", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackListQuestion()))
	newBot.RegisterCommandCallback("question_get", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackGetQuestion()))
	newBot.RegisterCommandCallback("delete_question", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackDeleteQuestion()))
	newBot.RegisterCommandCallback("question_delete", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackDeleteByIDQuestion()))
	newBot.RegisterCommandCallback("quiz_check", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackCheckQuiz()))
	newBot.RegisterCommandCallback("add_answers", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackCreateAnswer()))
	newBot.RegisterCommandCallback("quiz_answer", b.callbackQuiz.CallbackUserResponse()) // без middleware
	newBot.RegisterCommandCallback("show_channels", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackShowChannels()))
	// todo по хорошему вынести в другую область предметную
	newBot.RegisterCommandCallback("channel_get", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackSendQuizToChannel()))
	newBot.RegisterCommandCallback("add_image", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackAddImage()))
	newBot.RegisterCommandCallback("update_question", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackUpdateQuestion()))
	newBot.RegisterCommandCallback("cancel_update", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackCancelUpdate()))
	newBot.RegisterCommandCallback("update_answers", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackUpdateAnswers()))
	newBot.RegisterCommandCallback("downloading_rating", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackGetUserResultExcelFile()))
	newBot.RegisterCommandCallback("reset_rating", middleware.AdminMiddleware(b.userService, b.callbackQuiz.CallbackResetRating()))

	b.log.Info("Initialize bot took [%f] seconds", time.Since(startBot).Seconds())
	if err := newBot.Run(ctx); err != nil {
		b.log.Fatal("failed to run Telegram Bot: %v", err)
	}
}
