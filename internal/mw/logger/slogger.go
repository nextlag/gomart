// Package logger - middleware logger
package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/nextlag/gomart/internal/usecase"
)

// RequestFields содержит поля запроса для логгера.
type RequestFields struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	RemoteAddr  string `json:"remote_addr"`
	UserAgent   string `json:"user_agent"`
	RequestID   string `json:"request_id"`
	ContentType string `json:"content_type,omitempty"`
	Status      int    `json:"status"`
	Bytes       int    `json:"bytes,omitempty"`
	Duration    string `json:"duration"`
	Compress    string `json:"compress"`
}

// New - creates and returns a new middleware for logging HTTP requests.
func New(log usecase.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Создаем логгер запроса
			requestFields := RequestFields{
				Method:      r.Method,
				Path:        r.URL.Path,
				RemoteAddr:  r.RemoteAddr,
				UserAgent:   r.UserAgent(),
				RequestID:   middleware.GetReqID(r.Context()),
				ContentType: r.Header.Get("Content-Type"),
			}

			// Создаем WrapResponseWriter для перехвата статуса ответа и количества байтов.
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now() // Запускаем таймер для измерения продолжительности запроса.

			defer func() {
				// После выполнения запроса, логируем информацию о запросе, включая статус ответа, количество байтов и продолжительность.
				requestFields.Status = ww.Status()
				requestFields.Bytes = ww.BytesWritten()
				requestFields.Duration = time.Since(t1).String()
				requestFields.Compress = ww.Header().Get("Content-Encoding")

				// Добавляем логирование, только если статус запроса - ошибка
				if requestFields.Status >= http.StatusInternalServerError {
					log.Error("request completed with error", "error logger", requestFields)
				} else {
					log.Info("request", "request fields", requestFields)
				}
			}()

			// Передаем запрос следующему обработчику.
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
