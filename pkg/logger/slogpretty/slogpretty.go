// Package slogpretty предоставляет структурированный логгер с цветным форматированием вывода.

package slogpretty

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

// SetupLogger инициализирует и настраивает логгер с предоставленными параметрами.
// Возвращает объект slog.Logger для логирования.
func SetupLogger(projectRoot string) *slog.Logger {
	opts := PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level:     config.Cfg.LogLevel,
			AddSource: true,
		},
		ProjectRoot: projectRoot,
	}
	handler := opts.NewPrettyHandler(os.Stdout, projectRoot)
	return slog.New(handler)
}

// PrettyHandlerOptions содержит опции для создания PrettyHandler.
type PrettyHandlerOptions struct {
	SlogOpts    *slog.HandlerOptions
	ProjectRoot string
}

// PrettyHandler представляет собой обработчик структурированного логгера с цветным форматированием.
type PrettyHandler struct {
	opts PrettyHandlerOptions
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
	File  string
	Line  int
}

// NewPrettyHandler создает новый экземпляр PrettyHandler с заданным выводом и корневой директорией проекта.
func (opts PrettyHandlerOptions) NewPrettyHandler(out io.Writer, projectRoot string) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts),
		l:       stdLog.New(out, "", 0),
		opts:    PrettyHandlerOptions{ProjectRoot: projectRoot},
	}
	return h
}

// Handle реализует интерфейс slog.Handler. Обрабатывает записи логов, форматирует их и записывает в вывод.
func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.YellowString(level)
	case slog.LevelInfo:
		level = color.GreenString(level)
	case slog.LevelWarn:
		level = color.MagentaString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
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

// WithAttrs добавляет атрибуты к обработчику логгера.
func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}

// WithGroup создает новый обработчик логгера с заданным именем группы.
func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	// TODO: implement
	return &PrettyHandler{
		Handler: h.Handler.WithGroup(name),
		l:       h.l,
	}
}
