package sentryhook

import (
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

var (
	severityMap = map[logrus.Level]sentry.Level{
		logrus.TraceLevel: sentry.LevelDebug,
		logrus.DebugLevel: sentry.LevelDebug,
		logrus.InfoLevel:  sentry.LevelInfo,
		logrus.WarnLevel:  sentry.LevelWarning,
		logrus.ErrorLevel: sentry.LevelError,
		logrus.FatalLevel: sentry.LevelFatal,
		logrus.PanicLevel: sentry.LevelFatal,
	}
)

type WithScope func(entry *logrus.Entry) func(s *sentry.Scope)

type Hook struct {
	hub       *sentry.Hub
	levels    []logrus.Level
	withScope WithScope
}

func New(levels []logrus.Level) Hook {
	return NewWithScope(levels, defaultWithScope)
}

func NewWithScope(levels []logrus.Level, ws WithScope) Hook {
	return Hook{
		levels:    levels,
		hub:       sentry.CurrentHub(),
		withScope: ws,
	}
}

func (hook Hook) Levels() []logrus.Level {
	return hook.levels
}

func (hook Hook) Fire(entry *logrus.Entry) error {
	hub := sentry.CurrentHub().Clone()
	hub.WithScope(defaultWithScope(entry))

	if err, ok := entry.Data[logrus.ErrorKey].(error); ok {
		hub.CaptureException(err)
	} else {
		hub.CaptureMessage(entry.Message)
	}

	return nil
}

func defaultWithScope(entry *logrus.Entry) func(s *sentry.Scope) {
	return func(s *sentry.Scope) {
		s.SetLevel(severityMap[entry.Level])

		for k, v := range entry.Data {
			if k == "user_id" {
				if userID, ok := v.(string); ok {

					s.SetUser(sentry.User{
						ID: userID,
					})

					continue
				}
			}

			s.SetExtra(k, v)
		}
	}
}
