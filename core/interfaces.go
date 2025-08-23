package core

import (
	"context"
)

// Router основной интерфейс роутера
type Router interface {
	// RegisterModule регистрирует модуль
	RegisterModule(module Module) error
	
	// RegisterWildcard регистрирует wildcard обработчик
	RegisterWildcard(module WildcardModule) error
	
	// Route маршрутизирует сообщение
	Route(ctx UniversalContext) Response
	
	// GetModule получает модуль по имени
	GetModule(name string) (Module, bool)
	
	// ListModules возвращает список модулей
	ListModules() []Module
	
	// Start запускает роутер
	Start(ctx context.Context) error
	
	// Stop останавливает роутер
	Stop(ctx context.Context) error
}

// EventBus шина событий
type EventBus interface {
	// Subscribe подписывается на событие
	Subscribe(eventType string, handler EventHandlerFunc) error
	
	// Unsubscribe отписывается от события
	Unsubscribe(eventType string, handler EventHandlerFunc) error
	
	// Publish публикует событие
	Publish(ctx context.Context, event Event) error
	
	// PublishAsync публикует событие асинхронно
	PublishAsync(ctx context.Context, event Event)
	
	// Start запускает шину событий
	Start(ctx context.Context) error
	
	// Stop останавливает шину событий
	Stop(ctx context.Context) error
}

// Event базовый интерфейс события
type Event interface {
	// Type возвращает тип события
	Type() string
	
	// Timestamp возвращает время события
	Timestamp() int64
	
	// Source возвращает источник события
	Source() string
	
	// Data возвращает данные события
	Data() map[string]interface{}
	
	// UserID возвращает ID пользователя (если применимо)
	UserID() int64
	
	// ChatID возвращает ID чата (если применимо)
	ChatID() int64
}

// EventSubscription подписка на событие
type EventSubscription struct {
	EventType string
	Handler   EventHandlerFunc
	Filter    EventFilter
	Priority  int
}

// EventFilter фильтр событий
type EventFilter func(event Event) bool

// Logger интерфейс логгера
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

// Config интерфейс конфигурации
type Config interface {
	// Get получает значение по ключу
	Get(key string) interface{}
	
	// GetString получает строковое значение
	GetString(key string) string
	
	// GetInt получает целочисленное значение
	GetInt(key string) int
	
	// GetBool получает булево значение
	GetBool(key string) bool
	
	// GetStringSlice получает массив строк
	GetStringSlice(key string) []string
	
	// GetStringMap получает мапу строк
	GetStringMap(key string) map[string]string
	
	// Set устанавливает значение
	Set(key string, value interface{})
	
	// IsSet проверяет наличие ключа
	IsSet(key string) bool
}

// Middleware промежуточный обработчик
type Middleware interface {
	// Name возвращает имя middleware
	Name() string
	
	// Priority возвращает приоритет (больше = выше)
	Priority() int
	
	// Process обрабатывает запрос
	Process(ctx UniversalContext, next HandlerFunc) Response
}

// SecurityMiddleware middleware для безопасности
type SecurityMiddleware interface {
	Middleware
	
	// Authenticate аутентифицирует пользователя
	Authenticate(ctx UniversalContext) error
	
	// Authorize авторизует пользователя
	Authorize(ctx UniversalContext, permissions ...string) error
	
	// ValidateRequest валидирует запрос
	ValidateRequest(ctx UniversalContext) error
}

// RateLimiter ограничитель скорости
type RateLimiter interface {
	// Allow проверяет, разрешен ли запрос
	Allow(ctx UniversalContext) bool
	
	// Reset сбрасывает лимит для пользователя
	Reset(userID int64) error
	
	// GetLimit возвращает текущий лимит
	GetLimit(userID int64) (current, max int, resetAt int64)
}

// Cache интерфейс кеша
type Cache interface {
	// Get получает значение из кеша
	Get(ctx context.Context, key string) (interface{}, error)
	
	// Set устанавливает значение в кеш
	Set(ctx context.Context, key string, value interface{}, ttl int) error
	
	// Delete удаляет значение из кеша
	Delete(ctx context.Context, key string) error
	
	// Exists проверяет наличие ключа
	Exists(ctx context.Context, key string) (bool, error)
	
	// Clear очищает весь кеш
	Clear(ctx context.Context) error
}

// Storage интерфейс хранилища
type Storage interface {
	// Save сохраняет данные
	Save(ctx context.Context, key string, data interface{}) error
	
	// Load загружает данные
	Load(ctx context.Context, key string, dest interface{}) error
	
	// Delete удаляет данные
	Delete(ctx context.Context, key string) error
	
	// List возвращает список ключей
	List(ctx context.Context, prefix string) ([]string, error)
}

// Metrics интерфейс метрик
type Metrics interface {
	// Counter увеличивает счетчик
	Counter(name string, value int64, tags ...string)
	
	// Gauge устанавливает значение
	Gauge(name string, value float64, tags ...string)
	
	// Histogram записывает распределение
	Histogram(name string, value float64, tags ...string)
	
	// Timing записывает время выполнения
	Timing(name string, duration int64, tags ...string)
}

// APIHandler HTTP обработчик для API
type APIHandler struct {
	Method      string
	Path        string
	Handler     APIHandlerFunc
	Middlewares []Middleware
	Description string
}

// APIRequest запрос к API
type APIRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Query   map[string]string
	Body    interface{}
	UserID  int64
	Context context.Context
}

// APIResponse ответ API
type APIResponse struct {
	Status  int
	Headers map[string]string
	Body    interface{}
	Error   error
}

// Route типизированный роут для callback данных
type Route struct {
	Module string
	Action string
	Params map[string]interface{}
}