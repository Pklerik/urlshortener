// Package internalmiddleware provide decorators for handlers.
package internalmiddleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/pkg/jwtgenerator"
	"github.com/Pklerik/urlshortener/pkg/random"
)

var (
	SECRET_KEY, _ = random.RandBytes(32)
)

type (
	// берём структуру для хранения сведений об ответе.
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter.
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("(*loggingResponseWriter) Write: %w", err)
	}

	r.responseData.size += size // захватываем размер

	return size, nil
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// WithLogging добавляет дополнительный код для регистрации сведений о запросе
// и возвращает новый http.Handler.
func WithLogging(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		// функция Now() возвращает текущее время
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		// Since возвращает разницу во времени между start
		// и моментом вызова Since. Таким образом можно посчитать
		// время выполнения запроса.
		duration := time.Since(start)

		// отправляем сведения о запросе в zap
		logger.Sugar.Infoln(
			r.Method,
			r.RequestURI,
			r.Proto,
			responseData.status,
			responseData.size,
			duration,
		)
	}
	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}

// GZIPMiddleware provide gzip compression/decompression for request and response.
func GZIPMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")

		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")

		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		next.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(fn)
}

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p) // nolint
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.Header().Set("Content-Encoding", "gzip")
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close() //nolint
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err // nolint
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p) // nolint
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err //nolint
	}
	return c.zr.Close() // nolint
}

func AuthUser(next http.Handler) http.Handler {
	authFn := func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("AUTH")
		if err != nil && err != http.ErrNoCookie {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err == http.ErrNoCookie {
			validJWT, err := jwtgenerator.BuildJWTString(rand.Int()%100_000, SECRET_KEY)
			if err != nil {
				http.Error(w, "Unable to generete JWT", http.StatusInternalServerError)
				logger.Sugar.Errorf("Unable to generete JWT: %w", err)
				return
			}

			cookieAuth := http.Cookie{
				Name:  "AUTH",
				Value: validJWT,
			}
			http.SetCookie(w, &cookieAuth)
			r.AddCookie(&cookieAuth)
		}
		// передаём управление хендлеру
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(authFn)
}
