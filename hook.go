package sentryhook

import (
	"reflect"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

var (
	levelMap = map[logrus.Level]sentry.Level{
		logrus.TraceLevel: sentry.LevelDebug,
		logrus.DebugLevel: sentry.LevelDebug,
		logrus.InfoLevel:  sentry.LevelInfo,
		logrus.WarnLevel:  sentry.LevelWarning,
		logrus.ErrorLevel: sentry.LevelError,
		logrus.FatalLevel: sentry.LevelFatal,
		logrus.PanicLevel: sentry.LevelFatal,
	}
)

type Converter func(entry *logrus.Entry, hub *sentry.Hub) *sentry.Event

type Option func(h *Hook)

type Hook struct {
	hub       *sentry.Hub
	levels    []logrus.Level
	converter Converter
}

func New(levels []logrus.Level, options ...Option) Hook {
	h := Hook{
		levels:    levels,
		hub:       sentry.CurrentHub(),
		converter: DefaultConverter,
	}

	for _, option := range options {
		option(&h)
	}

	return h
}

func WithConverter(c Converter) Option {
	return func(h *Hook) {
		h.converter = c
	}
}

func WithHub(hub *sentry.Hub) Option {
	return func(h *Hook) {
		h.hub = hub
	}
}

func (hook Hook) Levels() []logrus.Level {
	return hook.levels
}

func (hook Hook) Fire(entry *logrus.Entry) error {
	hook.hub.CaptureEvent(
		hook.converter(entry, hook.hub),
	)

	return nil
}

func DefaultConverter(entry *logrus.Entry, hub *sentry.Hub) *sentry.Event {
	event := sentry.NewEvent()
	event.Level = levelMap[entry.Level]
	event.Message = entry.Message

	for k, v := range entry.Data {
		event.Extra[k] = v
	}

	if err, ok := entry.Data[logrus.ErrorKey].(error); ok {
		exception := sentry.Exception{
			Type:  reflect.TypeOf(err).String(),
			Value: err.Error(),
		}

		if hub.Client().Options().AttachStacktrace {
			exception.Stacktrace = sentry.ExtractStacktrace(err)
		}

		event.Exception = []sentry.Exception{exception}
	}

	return event
}
