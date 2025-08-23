package routing

import (
	"errors"
	"fmt"
	"github.com/andranikuz/botkit/core"
	"sync"
	"time"
)

// SecurityRule правила безопасности для маршрута
type SecurityRule struct {
	// RequireAuth требует аутентификации
	RequireAuth bool

	// RequireRoles требуемые роли
	RequireRoles []string

	// RequirePermissions требуемые права
	RequirePermissions []string

	// RequireProfile требует загруженный профиль
	RequireProfile bool

	// AllowedSources разрешенные источники (telegram, api, websocket)
	AllowedSources []string

	// RateLimit ограничение скорости
	RateLimit *RateLimitConfig

	// ValidateFunc кастомная функция валидации
	ValidateFunc ValidatorFunc

	// OnFailure обработчик ошибки безопасности
	OnFailure SecurityFailureHandler
}

// RateLimitConfig конфигурация ограничения скорости
type RateLimitConfig struct {
	// Requests количество запросов
	Requests int

	// Window временное окно в секундах
	Window int

	// BurstSize размер всплеска
	BurstSize int

	// Strategy стратегия (sliding_window, fixed_window, token_bucket)
	Strategy string
}

// ValidatorFunc функция валидации
type ValidatorFunc func(ctx core.UniversalContext) error

// SecurityFailureHandler обработчик ошибки безопасности
type SecurityFailureHandler func(ctx core.UniversalContext, err error) core.Response

// Check проверяет правила безопасности
func (s *SecurityRule) Check(ctx core.UniversalContext) error {
	// Проверяем аутентификацию
	if s.RequireAuth && !ctx.IsAuthenticated() {
		return ErrNotAuthenticated
	}

	// Проверяем профиль
	if s.RequireProfile && ctx.GetProfile() == nil {
		return ErrProfileRequired
	}

	// Проверяем роли
	if len(s.RequireRoles) > 0 {
		userRoles := ctx.GetRoles()
		if !hasAnyRole(userRoles, s.RequireRoles) {
			return ErrInsufficientRole
		}
	}

	// Проверяем права
	if len(s.RequirePermissions) > 0 {
		for _, perm := range s.RequirePermissions {
			if !ctx.HasPermission(perm) {
				return fmt.Errorf("%w: %s", ErrInsufficientPermission, perm)
			}
		}
	}

	// Проверяем источник
	if len(s.AllowedSources) > 0 {
		source := ctx.GetSource()
		if !contains(s.AllowedSources, source) {
			return fmt.Errorf("%w: %s", ErrSourceNotAllowed, source)
		}
	}

	// Проверяем кастомную валидацию
	if s.ValidateFunc != nil {
		if err := s.ValidateFunc(ctx); err != nil {
			return err
		}
	}

	return nil
}

// HandleFailure обрабатывает ошибку безопасности
func (s *SecurityRule) HandleFailure(ctx core.UniversalContext, err error) core.Response {
	if s.OnFailure != nil {
		return s.OnFailure(ctx, err)
	}

	// Дефолтный обработчик
	return defaultSecurityFailureHandler(ctx, err)
}

// Errors
var (
	ErrNotAuthenticated       = errors.New("authentication required")
	ErrProfileRequired        = errors.New("profile required")
	ErrInsufficientRole       = errors.New("insufficient role")
	ErrInsufficientPermission = errors.New("insufficient permission")
	ErrSourceNotAllowed       = errors.New("source not allowed")
	ErrRateLimitExceeded      = errors.New("rate limit exceeded")
	ErrValidationFailed       = errors.New("validation failed")
)

// defaultSecurityFailureHandler дефолтный обработчик ошибок безопасности
func defaultSecurityFailureHandler(ctx core.UniversalContext, err error) core.Response {
	var message string

	switch {
	case errors.Is(err, ErrNotAuthenticated):
		message = "❌ Требуется авторизация"
	case errors.Is(err, ErrProfileRequired):
		message = "❌ Требуется регистрация"
	case errors.Is(err, ErrInsufficientRole):
		message = "❌ Недостаточно прав для выполнения действия"
	case errors.Is(err, ErrInsufficientPermission):
		message = "❌ Нет доступа к этой функции"
	case errors.Is(err, ErrSourceNotAllowed):
		message = "❌ Действие недоступно из этого источника"
	case errors.Is(err, ErrRateLimitExceeded):
		message = "⏱ Слишком много запросов. Попробуйте позже"
	default:
		message = fmt.Sprintf("❌ Ошибка безопасности: %v", err)
	}

	return core.NewMessage(message)
}

// RateLimiter реализация ограничителя скорости
type RateLimiter struct {
	config   *RateLimitConfig
	storage  core.Storage
	mu       sync.RWMutex
	counters map[string]*rateLimitCounter
}

// rateLimitCounter счетчик для rate limiting
type rateLimitCounter struct {
	Count     int
	ResetAt   time.Time
	UpdatedAt time.Time
}

// NewRateLimiter создает новый ограничитель
func NewRateLimiter(config *RateLimitConfig, storage core.Storage) *RateLimiter {
	if config.Strategy == "" {
		config.Strategy = "sliding_window"
	}

	return &RateLimiter{
		config:   config,
		storage:  storage,
		counters: make(map[string]*rateLimitCounter),
	}
}

// Allow проверяет, разрешен ли запрос
func (r *RateLimiter) Allow(ctx core.UniversalContext) bool {
	key := r.getKey(ctx)

	r.mu.Lock()
	defer r.mu.Unlock()

	counter, exists := r.counters[key]
	now := time.Now()

	// Создаем новый счетчик если не существует
	if !exists || now.After(counter.ResetAt) {
		r.counters[key] = &rateLimitCounter{
			Count:     1,
			ResetAt:   now.Add(time.Duration(r.config.Window) * time.Second),
			UpdatedAt: now,
		}
		return true
	}

	// Проверяем лимит
	if counter.Count >= r.config.Requests {
		return false
	}

	// Увеличиваем счетчик
	counter.Count++
	counter.UpdatedAt = now

	return true
}

// Reset сбрасывает лимит для пользователя
func (r *RateLimiter) Reset(userID int64) error {
	key := fmt.Sprintf("user:%d", userID)

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.counters, key)
	return nil
}

// GetLimit возвращает текущий лимит
func (r *RateLimiter) GetLimit(userID int64) (current, max int, resetAt int64) {
	key := fmt.Sprintf("user:%d", userID)

	r.mu.RLock()
	defer r.mu.RUnlock()

	counter, exists := r.counters[key]
	if !exists {
		return 0, r.config.Requests, 0
	}

	return counter.Count, r.config.Requests, counter.ResetAt.Unix()
}

// getKey получает ключ для rate limiting
func (r *RateLimiter) getKey(ctx core.UniversalContext) string {
	// Используем комбинацию user_id и chat_id для ключа
	return fmt.Sprintf("user:%d:chat:%d", ctx.GetUserID(), ctx.GetChatID())
}

// SecurityMiddleware middleware для проверки безопасности
type SecurityMiddleware struct {
	rule        SecurityRule
	rateLimiter *RateLimiter
}

// NewSecurityMiddleware создает новый middleware безопасности
func NewSecurityMiddleware(rule SecurityRule) *SecurityMiddleware {
	var limiter *RateLimiter
	if rule.RateLimit != nil {
		limiter = NewRateLimiter(rule.RateLimit, nil)
	}

	return &SecurityMiddleware{
		rule:        rule,
		rateLimiter: limiter,
	}
}

// Name возвращает имя middleware
func (m *SecurityMiddleware) Name() string {
	return "security"
}

// Priority возвращает приоритет
func (m *SecurityMiddleware) Priority() int {
	return 100 // Высокий приоритет
}

// Process обрабатывает запрос
func (m *SecurityMiddleware) Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response {
	// Проверяем rate limit
	if m.rateLimiter != nil && !m.rateLimiter.Allow(ctx) {
		return m.rule.HandleFailure(ctx, ErrRateLimitExceeded)
	}

	// Проверяем правила безопасности
	if err := m.rule.Check(ctx); err != nil {
		return m.rule.HandleFailure(ctx, err)
	}

	// Продолжаем цепочку
	return next(ctx)
}

// Helper functions

func hasAnyRole(userRoles, requiredRoles []string) bool {
	for _, required := range requiredRoles {
		for _, userRole := range userRoles {
			if userRole == required {
				return true
			}
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
