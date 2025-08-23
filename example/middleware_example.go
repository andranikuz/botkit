package example

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andranikuz/botkit/core"
	"github.com/andranikuz/botkit/events"
	"github.com/andranikuz/botkit/middleware"
	"github.com/andranikuz/botkit/routing"
)

// MiddlewareExample демонстрирует использование middleware
func MiddlewareExample() {
	// Создаем зависимости
	logger := NewSimpleLogger()
	eventBus := events.NewEventBus(logger, nil)
	config := NewSimpleConfig()
	deps := NewSimpleDependencies(eventBus, logger, config)

	// Создаем роутер
	router := routing.NewRouter(eventBus, logger, config)
	router.SetDependencies(deps)

	// Создаем и регистрируем middleware

	// 1. Recovery middleware - должен быть первым (высший приоритет)
	recoveryMW := middleware.NewRecoveryMiddleware(logger, 100)
	router.RegisterMiddleware(recoveryMW)

	// 2. Logging middleware
	loggingMW := middleware.NewLoggingMiddleware(logger, 90)
	router.RegisterMiddleware(loggingMW)

	// 3. Rate limiting middleware
	// Используем RateLimiter из security.go
	rateLimitConfig := &routing.RateLimitConfig{
		Requests: 10,
		Window:   60,
		Strategy: "sliding_window",
	}
	rateLimiter := routing.NewRateLimiter(rateLimitConfig, nil)
	rateLimitMW := middleware.NewRateLimitMiddleware(rateLimiter, 80)
	router.RegisterMiddleware(rateLimitMW)

	// 4. Custom validation middleware
	validationMW := middleware.NewValidationMiddleware(func(ctx core.UniversalContext) error {
		text := ctx.GetText()
		if len(text) > 1000 {
			return fmt.Errorf("сообщение слишком длинное (максимум 1000 символов)")
		}
		if ctx.GetUserID() == 0 {
			return fmt.Errorf("неизвестный пользователь")
		}
		return nil
	}, 70)
	router.RegisterMiddleware(validationMW)

	// 5. Custom context middleware с таймаутом
	contextMW := middleware.NewContextMiddleware(func(ctx core.UniversalContext) context.Context {
		// Создаем контекст с таймаутом 30 секунд
		reqCtx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		return reqCtx
	}, 60)
	router.RegisterMiddleware(contextMW)

	// 6. Custom middleware с использованием функции
	customMW := middleware.NewMiddleware("custom", 50, func(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
		// Добавляем кастомный заголовок или данные
		ctx.SetParam("processed_at", time.Now().Format(time.RFC3339))

		// Проверяем специальные условия
		if ctx.GetText() == "secret" {
			return core.NewMessage("🤫 Это секретная команда!")
		}

		// Продолжаем выполнение
		return next(ctx)
	})
	router.RegisterMiddleware(customMW)

	// Регистрируем модули
	router.RegisterModule(NewMiddlewareTestModule())

	// Запускаем роутер
	if err := router.Start(context.Background()); err != nil {
		log.Fatal("Failed to start router:", err)
	}

	// Демонстрация работы
	log.Println("Middleware chain registered:")
	log.Println("1. Recovery (priority: 100) - обработка паник")
	log.Println("2. Logging (priority: 90) - логирование запросов")
	log.Println("3. RateLimit (priority: 80) - ограничение частоты")
	log.Println("4. Validation (priority: 70) - валидация данных")
	log.Println("5. Context (priority: 60) - добавление контекста")
	log.Println("6. Custom (priority: 50) - кастомная логика")
}

// MiddlewareTestModule модуль для тестирования middleware
type MiddlewareTestModule struct {
	name    string
	version string
	logger  core.Logger
}

func NewMiddlewareTestModule() *MiddlewareTestModule {
	return &MiddlewareTestModule{
		name:    "middleware_test",
		version: "1.0.0",
	}
}

func (m *MiddlewareTestModule) Name() string    { return m.name }
func (m *MiddlewareTestModule) Version() string { return m.version }

func (m *MiddlewareTestModule) Init(deps core.Dependencies) error {
	m.logger = deps.Logger()
	return nil
}

func (m *MiddlewareTestModule) Start(ctx context.Context) error {
	m.logger.Info("MiddlewareTest module started")
	return nil
}

func (m *MiddlewareTestModule) Stop(ctx context.Context) error {
	return nil
}

func (m *MiddlewareTestModule) Routes() []core.RoutePattern {
	return []core.RoutePattern{
		routing.RoutePattern{
			Patterns: []string{"/test", "test"},
			Handler:  m.handleTest,
			Priority: 100,
			Type:     routing.RouteTypeCommand,
			Meta: routing.RouteMeta{
				Name:        "test",
				Description: "Test middleware chain",
			},
		},
		routing.RoutePattern{
			Patterns: []string{"/panic", "panic"},
			Handler:  m.handlePanic,
			Priority: 100,
			Type:     routing.RouteTypeCommand,
			Meta: routing.RouteMeta{
				Name:        "panic",
				Description: "Test panic recovery",
			},
		},
		routing.RoutePattern{
			Patterns: []string{"/slow", "slow"},
			Handler:  m.handleSlow,
			Priority: 100,
			Type:     routing.RouteTypeCommand,
			Meta: routing.RouteMeta{
				Name:        "slow",
				Description: "Test slow operation",
			},
		},
		routing.RoutePattern{
			Patterns: []string{"/auth", "auth"},
			Handler:  m.handleAuth,
			Priority: 100,
			Type:     routing.RouteTypeCommand,
			Security: routing.SecurityRule{
				RequireAuth:  true,
				RequireRoles: []string{"admin"},
			},
			Meta: routing.RouteMeta{
				Name:        "auth",
				Description: "Test authentication",
			},
		},
	}
}

func (m *MiddlewareTestModule) handleTest(ctx core.UniversalContext) core.Response {
	// Получаем данные, добавленные middleware
	processedAt, _ := ctx.GetParam("processed_at")

	response := fmt.Sprintf(
		"✅ Middleware chain test successful!\n\n"+
			"User ID: %d\n"+
			"Text: %s\n"+
			"Processed at: %s\n"+
			"All middleware executed successfully!",
		ctx.GetUserID(),
		ctx.GetText(),
		processedAt,
	)

	return core.NewMessage(response)
}

func (m *MiddlewareTestModule) handlePanic(ctx core.UniversalContext) core.Response {
	// Симулируем панику для проверки recovery middleware
	m.logger.Info("About to panic!")
	panic("Test panic! This should be caught by recovery middleware")
}

func (m *MiddlewareTestModule) handleSlow(ctx core.UniversalContext) core.Response {
	// Симулируем долгую операцию
	m.logger.Info("Starting slow operation...")

	// Получаем контекст, добавленный middleware
	if reqCtx, ok := ctx.GetParam("_context"); ok {
		if ctx, ok := reqCtx.(context.Context); ok {
			select {
			case <-time.After(5 * time.Second):
				return core.NewMessage("✅ Slow operation completed!")
			case <-ctx.Done():
				return core.NewMessage("⏱️ Operation cancelled by timeout")
			}
		}
	}

	time.Sleep(5 * time.Second)
	return core.NewMessage("✅ Slow operation completed (no context)")
}

func (m *MiddlewareTestModule) handleAuth(ctx core.UniversalContext) core.Response {
	// Этот обработчик требует аутентификации
	// SecurityMiddleware должен проверить права доступа
	return core.NewMessage("🔓 Authenticated! You have admin access.")
}

// DemoMiddlewareChain демонстрирует создание цепочки middleware
func DemoMiddlewareChain() {
	logger := NewSimpleLogger()

	// Создаем цепочку middleware
	middlewares := []routing.Middleware{
		middleware.NewRecoveryMiddleware(logger, 100),
		middleware.NewLoggingMiddleware(logger, 90),
		middleware.NewValidationMiddleware(func(ctx core.UniversalContext) error {
			if ctx.GetText() == "" {
				return fmt.Errorf("пустое сообщение")
			}
			return nil
		}, 80),
	}

	// Создаем обработчик с цепочкой middleware
	chain := middleware.Chain(middlewares...)

	// Финальный обработчик
	handler := func(ctx core.UniversalContext) core.Response {
		return core.NewMessage("Hello from handler!")
	}

	// Применяем цепочку
	wrappedHandler := chain(handler)

	// Создаем тестовый контекст
	testCtx := &core.BaseContext{}
	testCtx.SetUserID(123)
	testCtx.SetText("test message")

	// Выполняем
	response := wrappedHandler(testCtx)

	log.Printf("Response: %+v", response)
}

// CustomAuthMiddleware пример кастомного middleware для аутентификации
type CustomAuthMiddleware struct {
	token    string
	priority int
}

func NewCustomAuthMiddleware(token string) *CustomAuthMiddleware {
	return &CustomAuthMiddleware{
		token:    token,
		priority: 85,
	}
}

func (m *CustomAuthMiddleware) Name() string {
	return "custom_auth"
}

func (m *CustomAuthMiddleware) Priority() int {
	return m.priority
}

func (m *CustomAuthMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
	// Проверяем токен в сообщении
	if ctx.GetText() == m.token {
		// Устанавливаем флаг аутентификации
		ctx.SetParam("authenticated", true)
		ctx.SetParam("auth_time", time.Now())
	}

	// Проверяем, требуется ли аутентификация
	if needsAuth, ok := ctx.GetParam("require_auth"); ok && needsAuth.(bool) {
		if authenticated, ok := ctx.GetParam("authenticated"); !ok || !authenticated.(bool) {
			return core.NewMessage("🔒 Authentication required. Please provide token.")
		}
	}

	return next(ctx)
}

// CachingMiddleware пример middleware для кеширования
type CachingMiddleware struct {
	cache    map[string]core.Response
	ttl      time.Duration
	priority int
}

func NewCachingMiddleware(ttl time.Duration) *CachingMiddleware {
	return &CachingMiddleware{
		cache:    make(map[string]core.Response),
		ttl:      ttl,
		priority: 75,
	}
}

func (m *CachingMiddleware) Name() string {
	return "caching"
}

func (m *CachingMiddleware) Priority() int {
	return m.priority
}

func (m *CachingMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
	// Создаем ключ кеша
	key := fmt.Sprintf("%d:%s", ctx.GetUserID(), ctx.GetText())

	// Проверяем кеш
	if cached, ok := m.cache[key]; ok {
		// Возвращаем закешированный ответ
		return cached
	}

	// Выполняем обработчик
	response := next(ctx)

	// Кешируем ответ
	m.cache[key] = response

	// Очищаем кеш через TTL
	go func() {
		time.Sleep(m.ttl)
		delete(m.cache, key)
	}()

	return response
}
