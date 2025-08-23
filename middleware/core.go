package middleware

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/andranikuz/botkit/core"
	"github.com/andranikuz/botkit/routing"
)

// Ensure all middleware types implement routing.Middleware interface
var (
	_ routing.Middleware = (*BaseMiddleware)(nil)
	_ routing.Middleware = (*LoggingMiddleware)(nil)
	_ routing.Middleware = (*RecoveryMiddleware)(nil)
	_ routing.Middleware = (*AuthMiddleware)(nil)
	_ routing.Middleware = (*RateLimitMiddleware)(nil)
	_ routing.Middleware = (*MetricsMiddleware)(nil)
	_ routing.Middleware = (*ContextMiddleware)(nil)
	_ routing.Middleware = (*ValidationMiddleware)(nil)
)

// BaseMiddleware базовая реализация middleware
type BaseMiddleware struct {
	name     string
	priority int
	handler  func(core.UniversalContext, core.HandlerFunc) core.Response
}

// NewMiddleware создает новый middleware
func NewMiddleware(name string, priority int, handler func(core.UniversalContext, core.HandlerFunc) core.Response) *BaseMiddleware {
	return &BaseMiddleware{
		name:     name,
		priority: priority,
		handler:  handler,
	}
}

// Name возвращает имя
func (m *BaseMiddleware) Name() string {
	return m.name
}

// Priority возвращает приоритет
func (m *BaseMiddleware) Priority() int {
	return m.priority
}

// Process обрабатывает запрос
func (m *BaseMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
	return m.handler(ctx, next)
}

// LoggingMiddleware middleware для логирования
type LoggingMiddleware struct {
	logger   core.Logger
	priority int
}

// NewLoggingMiddleware создает middleware для логирования
func NewLoggingMiddleware(logger core.Logger, priority int) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger:   logger,
		priority: priority,
	}
}

// Name возвращает имя
func (m *LoggingMiddleware) Name() string {
	return "logging"
}

// Priority возвращает приоритет
func (m *LoggingMiddleware) Priority() int {
	return m.priority
}

// Process логирует запрос и ответ
func (m *LoggingMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
	start := time.Now()

	// Логируем входящий запрос
	m.logger.Info("Request received",
		"user_id", ctx.GetUserID(),
		"chat_id", ctx.GetChatID(),
		"text", ctx.GetText(),
		"is_command", ctx.IsCommand(),
		"is_callback", ctx.IsCallback(),
	)

	// Выполняем обработчик
	response := next(ctx)

	// Логируем ответ
	duration := time.Since(start)
	m.logger.Info("Request processed",
		"user_id", ctx.GetUserID(),
		"duration_ms", duration.Milliseconds(),
		"response_type", fmt.Sprintf("%T", response),
	)

	return response
}

// RecoveryMiddleware middleware для обработки паник
type RecoveryMiddleware struct {
	logger   core.Logger
	priority int
}

// NewRecoveryMiddleware создает middleware для обработки паник
func NewRecoveryMiddleware(logger core.Logger, priority int) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger:   logger,
		priority: priority,
	}
}

// Name возвращает имя
func (m *RecoveryMiddleware) Name() string {
	return "recovery"
}

// Priority возвращает приоритет
func (m *RecoveryMiddleware) Priority() int {
	return m.priority
}

// Process обрабатывает панику
func (m *RecoveryMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) (response core.Response) {
	defer func() {
		if r := recover(); r != nil {
			// Логируем панику
			m.logger.Error("Panic recovered",
				"panic", r,
				"stack", string(debug.Stack()),
				"user_id", ctx.GetUserID(),
				"text", ctx.GetText(),
			)

			// Возвращаем ошибку пользователю
			response = core.NewMessage("❌ Произошла внутренняя ошибка. Попробуйте позже.")
		}
	}()

	return next(ctx)
}

// AuthMiddleware middleware для проверки аутентификации
type AuthMiddleware struct {
	checkAuth func(core.UniversalContext) bool
	priority  int
}

// NewAuthMiddleware создает middleware для проверки аутентификации
func NewAuthMiddleware(checkAuth func(core.UniversalContext) bool, priority int) *AuthMiddleware {
	return &AuthMiddleware{
		checkAuth: checkAuth,
		priority:  priority,
	}
}

// Name возвращает имя
func (m *AuthMiddleware) Name() string {
	return "auth"
}

// Priority возвращает приоритет
func (m *AuthMiddleware) Priority() int {
	return m.priority
}

// Process проверяет аутентификацию
func (m *AuthMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
	if !m.checkAuth(ctx) {
		return core.NewMessage("🔒 Доступ запрещен. Требуется аутентификация.")
	}

	return next(ctx)
}

// RateLimitMiddleware middleware для ограничения частоты запросов
type RateLimitMiddleware struct {
	limiter  *routing.RateLimiter
	priority int
}

// NewRateLimitMiddleware создает middleware для rate limiting
func NewRateLimitMiddleware(limiter *routing.RateLimiter, priority int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter:  limiter,
		priority: priority,
	}
}

// Name возвращает имя
func (m *RateLimitMiddleware) Name() string {
	return "rate_limit"
}

// Priority возвращает приоритет
func (m *RateLimitMiddleware) Priority() int {
	return m.priority
}

// Process проверяет rate limit
func (m *RateLimitMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
	if !m.limiter.Allow(ctx) {
		return core.NewMessage("⏱️ Слишком много запросов. Подождите немного.")
	}

	return next(ctx)
}

// MetricsMiddleware middleware для сбора метрик
type MetricsMiddleware struct {
	metrics  core.Metrics
	priority int
}

// NewMetricsMiddleware создает middleware для сбора метрик
func NewMetricsMiddleware(metrics core.Metrics, priority int) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics:  metrics,
		priority: priority,
	}
}

// Name возвращает имя
func (m *MetricsMiddleware) Name() string {
	return "metrics"
}

// Priority возвращает приоритет
func (m *MetricsMiddleware) Priority() int {
	return m.priority
}

// Process собирает метрики
func (m *MetricsMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
	start := time.Now()

	// Увеличиваем счетчик запросов
	if m.metrics != nil {
		m.metrics.Counter("requests.total", 1,
			"type", getRequestType(ctx),
		)
	}

	// Выполняем обработчик
	response := next(ctx)

	// Записываем время обработки
	if m.metrics != nil {
		duration := time.Since(start)
		m.metrics.Timing("requests.duration", int64(duration.Milliseconds()),
			"type", getRequestType(ctx),
		)
	}

	return response
}

func getRequestType(ctx core.UniversalContext) string {
	if ctx.IsCommand() {
		return "command"
	}
	if ctx.IsCallback() {
		return "callback"
	}
	return "message"
}

// ContextMiddleware middleware для добавления контекста
type ContextMiddleware struct {
	contextFunc func(core.UniversalContext) context.Context
	priority    int
}

// NewContextMiddleware создает middleware для добавления контекста
func NewContextMiddleware(contextFunc func(core.UniversalContext) context.Context, priority int) *ContextMiddleware {
	return &ContextMiddleware{
		contextFunc: contextFunc,
		priority:    priority,
	}
}

// Name возвращает имя
func (m *ContextMiddleware) Name() string {
	return "context"
}

// Priority возвращает приоритет
func (m *ContextMiddleware) Priority() int {
	return m.priority
}

// Process добавляет контекст
func (m *ContextMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
	// Создаем контекст с таймаутом
	reqCtx := m.contextFunc(ctx)

	// Сохраняем контекст в UniversalContext
	ctx.SetParam("_context", reqCtx)

	// Проверяем отмену контекста
	select {
	case <-reqCtx.Done():
		return core.NewMessage("⏱️ Время обработки запроса истекло.")
	default:
		return next(ctx)
	}
}

// ValidationMiddleware middleware для валидации данных
type ValidationMiddleware struct {
	validator func(core.UniversalContext) error
	priority  int
}

// NewValidationMiddleware создает middleware для валидации
func NewValidationMiddleware(validator func(core.UniversalContext) error, priority int) *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: validator,
		priority:  priority,
	}
}

// Name возвращает имя
func (m *ValidationMiddleware) Name() string {
	return "validation"
}

// Priority возвращает приоритет
func (m *ValidationMiddleware) Priority() int {
	return m.priority
}

// Process валидирует данные
func (m *ValidationMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
	if err := m.validator(ctx); err != nil {
		return core.NewMessage(fmt.Sprintf("❌ Ошибка валидации: %s", err.Error()))
	}

	return next(ctx)
}

// Chain создает цепочку middleware
func Chain(middlewares ...routing.Middleware) func(core.HandlerFunc) core.HandlerFunc {
	return func(final core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.UniversalContext) core.Response {
			// Строим цепочку от последнего к первому
			handler := final
			for i := len(middlewares) - 1; i >= 0; i-- {
				mw := middlewares[i]
				next := handler
				handler = func(c core.UniversalContext) core.Response {
					return mw.Process(c, next)
				}
			}
			return handler(ctx)
		}
	}
}

// MiddlewareFunc адаптер для функций как middleware
type MiddlewareFunc func(core.UniversalContext, core.HandlerFunc) core.Response

// Process реализует интерфейс Middleware
func (f MiddlewareFunc) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
	return f(ctx, next)
}

// Name возвращает имя
func (f MiddlewareFunc) Name() string {
	return "func"
}

// Priority возвращает приоритет
func (f MiddlewareFunc) Priority() int {
	return 50
}
