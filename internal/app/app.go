package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"

	"go.uber.org/zap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/wanomir/e"
	"github.com/wanomir/keychron-tg-parser/pkg/logger"
)

const (
	exitStatusOk     = 0
	exitStatusFailed = 1
)

type App struct {
	config *Config
	logger *zap.Logger
	bot    *tgbotapi.BotAPI
}

func NewApp() (*App, error) {
	a := new(App)

	if err := a.init(); err != nil {
		return nil, e.Wrap("failed to init app", err)
	}

	return a, nil
}

func (a *App) Run() (exitCode int) {
	defer a.recoverFromPanic(&exitCode)
	var err error

	errChan := make(chan error)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	// for now, just mirror incoming messages to make sure that the bot works
	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		// TODO: move updates chanel creation outside running goroutine
		//		 so it will be placed next to `StopReceivingUpdates()`
		updates := a.bot.GetUpdatesChan(u)

		for {
			select {
			case update := <-updates:
				if update.Message != nil { // If we got a message
					a.logger.Info(fmt.Sprintf("[%s] %s", update.Message.From.UserName, update.Message.Text))

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
					msg.ReplyToMessageID = update.Message.MessageID

					_, _ = a.bot.Send(msg)
				}
			case <-ctx.Done():
				a.logger.Info("telegram shutdown")
				return
			}
		}
	}()
	defer a.bot.StopReceivingUpdates()

	select {
	case err = <-errChan:
		a.logger.Error(e.Wrap("fatal error, service shutdown", err).Error())
		exitCode = exitStatusFailed
	case <-ctx.Done():
		a.logger.Info("service shutdown")
	}

	return exitStatusOk
}

func (a *App) init() (err error) {
	// config
	if err = a.readConfig(); err != nil {
		return e.Wrap("failed to read config", err)
	}

	// TODO: log events in the usecase layer
	a.logger = logger.NewLogger(a.config.Log.Level)

	// telegram service
	if a.bot, err = tgbotapi.NewBotAPI(a.config.TG.Token); err != nil {
		return e.Wrap("failed to init telegram", err)
	}
	a.logger.Info("authorized telegram service", zap.String("account", a.bot.Self.UserName))

	return nil
}

func (a *App) readConfig() (err error) {
	a.config = new(Config)
	if err = cleanenv.ReadEnv(a.config); err != nil {
		return err
	}

	return nil
}

func (a *App) recoverFromPanic(exitCode *int) {
	if panicErr := recover(); panicErr != nil {
		a.logger.Error(fmt.Sprintf("recover from panic: %v, stacktrace: %s", panicErr, string(debug.Stack())))
		*exitCode = exitStatusFailed
	}
}
