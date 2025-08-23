package routing

import "github.com/andranikuz/botkit/core"

// Middleware интерфейс для промежуточных обработчиков
type Middleware interface {
	// Name возвращает имя middleware
	Name() string

	// Priority возвращает приоритет (больше = выше)
	Priority() int

	// Process обрабатывает запрос
	Process(ctx core.UniversalContext, next core.HandlerFunc) core.Response
}
