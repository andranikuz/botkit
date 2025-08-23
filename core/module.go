package core

import (
	"context"
)

// Module представляет базовый интерфейс модуля системы
type Module interface {
	// Name возвращает уникальное имя модуля
	Name() string
	
	// Version возвращает версию модуля
	Version() string
	
	// Routes возвращает список маршрутов модуля
	Routes() []RoutePattern
	
	// Init инициализирует модуль с зависимостями
	Init(deps Dependencies) error
	
	// Start запускает модуль
	Start(ctx context.Context) error
	
	// Stop останавливает модуль
	Stop(ctx context.Context) error
}

// EventAwareModule модуль с поддержкой событий
type EventAwareModule interface {
	Module
	
	// Events возвращает список подписок на события
	Events() []EventSubscription
	
	// HandleEvent обрабатывает входящее событие
	HandleEvent(ctx context.Context, event Event) error
}

// WildcardModule модуль для обработки неструктурированных сообщений
type WildcardModule interface {
	Module
	
	// Priority возвращает приоритет обработки (больше = выше приоритет)
	Priority() int
	
	// ShouldHandle проверяет, должен ли модуль обработать сообщение
	ShouldHandle(ctx UniversalContext) bool
	
	// HandleWildcard обрабатывает неструктурированное сообщение
	HandleWildcard(ctx UniversalContext) Response
}

// APIModule модуль с HTTP API endpoints
type APIModule interface {
	Module
	
	// APIHandlers возвращает список HTTP обработчиков
	APIHandlers() []APIHandler
}

// Dependencies зависимости для инициализации модуля
type Dependencies interface {
	// Database возвращает подключение к БД
	Database() interface{}
	
	// EventBus возвращает шину событий
	EventBus() EventBus
	
	// Logger возвращает логгер
	Logger() Logger
	
	// Config возвращает конфигурацию
	Config() Config
	
	// Get возвращает произвольную зависимость по ключу
	Get(key string) (interface{}, bool)
	
	// Set устанавливает произвольную зависимость
	Set(key string, value interface{})
}

// HandlerFunc функция-обработчик для маршрута
type HandlerFunc func(ctx UniversalContext) Response

// APIHandlerFunc функция-обработчик для API
type APIHandlerFunc func(ctx context.Context, req APIRequest) (APIResponse, error)

// EventHandlerFunc функция-обработчик события
type EventHandlerFunc func(ctx context.Context, event Event) error

// ModuleRegistry реестр модулей
type ModuleRegistry interface {
	// Register регистрирует модуль
	Register(module Module) error
	
	// Get возвращает модуль по имени
	Get(name string) (Module, bool)
	
	// List возвращает список всех модулей
	List() []Module
	
	// Start запускает все модули
	Start(ctx context.Context) error
	
	// Stop останавливает все модули
	Stop(ctx context.Context) error
}

// ModuleMetadata метаданные модуля
type ModuleMetadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Author      string            `json:"author"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Config      map[string]string `json:"config"`
}

// Lifecycle хуки жизненного цикла модуля
type Lifecycle interface {
	// OnInit вызывается при инициализации
	OnInit(ctx context.Context) error
	
	// OnStart вызывается при запуске
	OnStart(ctx context.Context) error
	
	// OnStop вызывается при остановке
	OnStop(ctx context.Context) error
	
	// OnError вызывается при ошибке
	OnError(ctx context.Context, err error)
	
	// Health возвращает статус здоровья модуля
	Health() HealthStatus
}

// HealthStatus статус здоровья модуля
type HealthStatus struct {
	Healthy bool              `json:"healthy"`
	Message string            `json:"message"`
	Details map[string]string `json:"details"`
}