package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// CORSMiddleware middleware для обработки CORS
type CORSMiddleware struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// NewCORSMiddleware создает новый CORS middleware
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

// Handler возвращает HTTP handler с CORS
func (c *CORSMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Проверяем разрешенные источники
		if c.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(c.AllowedOrigins) == 1 && c.AllowedOrigins[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Устанавливаем остальные заголовки
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.AllowedHeaders, ", "))

		if c.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if c.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", c.MaxAge))
		}

		// Обрабатываем preflight запросы
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// isOriginAllowed проверяет, разрешен ли origin
func (c *CORSMiddleware) isOriginAllowed(origin string) bool {
	for _, allowed := range c.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// RequestIDMiddleware добавляет request ID к запросу
type RequestIDMiddleware struct {
	HeaderName string
	Generator  func() string
}

// NewRequestIDMiddleware создает middleware для request ID
func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{
		HeaderName: "X-Request-ID",
		Generator:  generateRequestID,
	}
}

// Handler возвращает HTTP handler с request ID
func (m *RequestIDMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем существующий request ID
		requestID := r.Header.Get(m.HeaderName)
		if requestID == "" {
			requestID = m.Generator()
			r.Header.Set(m.HeaderName, requestID)
		}

		// Добавляем в ответ
		w.Header().Set(m.HeaderName, requestID)

		next(w, r)
	}
}

// generateRequestID генерирует уникальный request ID
func generateRequestID() string {
	// Простая реализация с timestamp
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
}

// CompressionMiddleware middleware для сжатия ответов
type CompressionMiddleware struct {
	Level int
}

// NewCompressionMiddleware создает middleware для сжатия
func NewCompressionMiddleware() *CompressionMiddleware {
	return &CompressionMiddleware{
		Level: 5, // Средний уровень сжатия
	}
}

// Handler возвращает HTTP handler со сжатием
func (c *CompressionMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем поддержку gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}

		// Создаем gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Устанавливаем заголовки
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")

		// Оборачиваем ResponseWriter
		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gz,
		}

		next(gzw, r)
	}
}

// gzipResponseWriter обертка для ResponseWriter с gzip
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// SecurityHeadersMiddleware добавляет заголовки безопасности
type SecurityHeadersMiddleware struct {
	FrameOptions       string
	ContentTypeOptions string
	XSSProtection      string
	ReferrerPolicy     string
}

// NewSecurityHeadersMiddleware создает middleware для заголовков безопасности
func NewSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{
		FrameOptions:       "DENY",
		ContentTypeOptions: "nosniff",
		XSSProtection:      "1; mode=block",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}
}

// Handler возвращает HTTP handler с заголовками безопасности
func (s *SecurityHeadersMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", s.FrameOptions)
		w.Header().Set("X-Content-Type-Options", s.ContentTypeOptions)
		w.Header().Set("X-XSS-Protection", s.XSSProtection)
		w.Header().Set("Referrer-Policy", s.ReferrerPolicy)

		next(w, r)
	}
}

// ChainHTTP создает цепочку HTTP middleware
func ChainHTTP(middlewares ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(final http.HandlerFunc) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
