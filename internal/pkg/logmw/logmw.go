package logmw

import (
	"github.com/go-logr/logr"
	tele "gopkg.in/telebot.v3"
)

const LoggerKey = "logger"

func NewLogMW(logger logr.Logger) func(next tele.HandlerFunc) tele.HandlerFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			c.Set(LoggerKey, logger.WithValues("chatID", c.Chat().ID, "reqID", c.Message().ID))
			return next(c)
		}
	}
}

func GetLogger(c tele.Context) logr.Logger {
	return c.Get(LoggerKey).(logr.Logger)
}
