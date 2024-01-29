// Package gzip - middleware gzip
package gzip

import (
	"net/http"
	"strings"
)

// New возвращает middleware для обработки сжатия gzip для HTTP-запросов и ответов.
//
// Эта функция создает middleware, которое проверяет поддержку сжатия gzip для HTTP-запросов и ответов.
// Если поддержка gzip обнаружена в заголовке "Accept-Encoding" запроса, middleware сжимает ответ с использованием gzip.
// Если контент был отправлен с использованием сжатия gzip в заголовке "Content-Encoding", middleware распаковывает его.
// Возвращенное middleware оборачивает следующий обработчик HTTP и модифицирует rest.ResponseWriter и входящий запрос,
// чтобы поддерживать сжатие gzip.
//
// Возвращаемые значения:
//   - func(http.Handler) http.Handler: middleware для обработки сжатия gzip для HTTP-запросов и ответов.
func New() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ow := w // Создаем переменную ow и инициализируем ее с rest.ResponseWriter из входящих параметров.

			// Получаем значение заголовка "Accept-Encoding" из запроса.
			acceptEncoding := r.Header.Get("Accept-Encoding")

			// Проверяем поддержку сжатия gzip. Если заголовок "Accept-Encoding" содержит "gzip", устанавливаем флаг supportGzip.
			supportGzip := strings.Contains(acceptEncoding, "gzip")

			// Если поддержка gzip обнаружена, создаем новый gzip.Writer (cw) и устанавливаем ow на него.
			if supportGzip {
				cw := NewCompressWriter(w)
				ow = cw
				defer cw.Close() // Отложенное закрытие gzip.Writer после завершения обработки.
			}

			// Получаем значение заголовка "Content-Encoding" из запроса.
			contentEncoding := r.Header.Get("Content-Encoding")

			// Проверяем, был ли отправлен контент с использованием сжатия gzip. Если "Content-Encoding" содержит "gzip", устанавливаем флаг sendGzip.
			sendGzip := strings.Contains(contentEncoding, "gzip")

			// Если контент был отправлен с использованием gzip, создаем новый gzip.Reader (cr) и устанавливаем r.Body на него.
			if sendGzip {
				cr, err := NewCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer cr.Close() // Отложенное закрытие gzip.Reader после завершения обработки.
			}

			// Вызываем оригинальный обработчик (next) с модифицированным rest.ResponseWriter (ow) и исходным запросом (r).
			next.ServeHTTP(ow, r)
		}
		return http.HandlerFunc(fn)
	}
}
