package core

// RoutePattern интерфейс для паттерна маршрута
// Реальная реализация в пакете routing
type RoutePattern interface {
	// GetPattern возвращает паттерн
	GetPattern() []string
	
	// GetHandler возвращает обработчик
	GetHandler() HandlerFunc
	
	// GetPriority возвращает приоритет
	GetPriority() int
	
	// GetType возвращает тип маршрута
	GetType() string
}