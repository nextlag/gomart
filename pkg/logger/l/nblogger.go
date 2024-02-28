package l

import (
	"context"
	"encoding/json"
	"io"
	stdLog "log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"

	"github.com/nextlag/gomart/internal/config"
)

func LoggerNew(projectRoot string) *Logger {
	opts := LoggerOptions{
		SlogOpts: &HandlerOptions{
			Level:     config.Cfg.LogLevel,
			AddSource: true,
		},
		ProjectRoot: projectRoot,
	}
	handler := opts.NewNbHandler(os.Stdout, projectRoot)
	return slog.New(handler)
}

type LoggerOptions struct {
	SlogOpts    *HandlerOptions
	ProjectRoot string
}

type NbHandler struct {
	opts LoggerOptions
	Handler
	l     *stdLog.Logger
	attrs []Attr
	File  string
	Line  int
}

func (opts LoggerOptions) NewNbHandler(out io.Writer, projectRoot string) *NbHandler {
	h := &NbHandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts),
		l:       stdLog.New(out, "", 0),
		opts:    LoggerOptions{ProjectRoot: projectRoot},
	}

	return h
}

func (h *NbHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case LevelDebug:
		level = color.YellowString(level)
	case LevelInfo:
		level = color.GreenString(level)
	case LevelWarn:
		level = color.MagentaString(level)
	case LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())

	r.Attrs(func(a Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	var b []byte
	var err error

	if len(fields) > 0 {
		b, err = json.Marshal(fields)
		if err != nil {
			return err
		}
	}
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???" // Если информация не доступна
		line = 0
	}
	relPath, err := filepath.Rel(h.opts.ProjectRoot, file)
	if err != nil {
		// Если произошла ошибка, использовать полный путь
		relPath = file
	}

	// Добавляем информацию о вызове в лог
	h.File = relPath
	h.Line = line

	timeStr := r.Time.Format("[02.01.2006 15:04:05.000]")
	msg := color.CyanString(r.Message)

	h.l.Println(
		timeStr,
		level,
		msg,
		color.WhiteString("%s:%d", h.File, h.Line),
		color.WhiteString(string(b)),
	)

	return nil
}

func (h *NbHandler) WithAttrs(attrs []Attr) Handler {
	return &NbHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}

func (h *NbHandler) WithGroup(name string) Handler {
	// TODO: implement
	return &NbHandler{
		Handler: h.Handler.WithGroup(name),
		l:       h.l,
	}
}

func L(ctx context.Context) *Logger {
	return loggerFromContext(ctx)
}
