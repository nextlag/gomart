package gzip

import (
	"net/http"
	"strings"
)

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

			// Вызываем оригинальный обработчик (h) с модифицированным rest.ResponseWriter (ow) и исходным запросом (r).
			next.ServeHTTP(ow, r)
		}
		return http.HandlerFunc(fn)
	}
}
