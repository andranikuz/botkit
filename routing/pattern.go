package routing

import (
	"github.com/andranikuz/botkit/core"
	"regexp"
	"strings"
)

// RoutePattern описывает паттерн маршрута
// Реализует core.RoutePattern интерфейс
type RoutePattern struct {
	// Patterns список текстовых паттернов для матчинга
	Patterns []string

	// Handler функция-обработчик
	Handler core.HandlerFunc

	// Priority приоритет (больше = выше)
	Priority int

	// Type тип маршрута
	Type RouteType

	// Security правила безопасности
	Security SecurityRule

	// Meta метаданные маршрута
	Meta RouteMeta

	// Module имя модуля-владельца
	Module string

	// compiled скомпилированные регулярные выражения
	compiled []*regexp.Regexp
}

// RouteType тип маршрута
type RouteType string

const (
	RouteTypeCommand  RouteType = "command"
	RouteTypeCallback RouteType = "callback"
	RouteTypeMessage  RouteType = "message"
	RouteTypeRegex    RouteType = "regex"
	RouteTypeWildcard RouteType = "wildcard"
)

// RouteMeta метаданные маршрута
type RouteMeta struct {
	// Name имя маршрута
	Name string

	// Description описание
	Description string

	// Category категория
	Category string

	// Tags теги
	Tags []string

	// Hidden скрытый маршрут (не показывать в help)
	Hidden bool

	// Deprecated устаревший маршрут
	Deprecated bool

	// Version версия маршрута
	Version string

	// Examples примеры использования
	Examples []string
}

// Методы для реализации core.RoutePattern интерфейса

// GetPattern возвращает паттерны
func (r RoutePattern) GetPattern() []string {
	return r.Patterns
}

// GetHandler возвращает обработчик
func (r RoutePattern) GetHandler() core.HandlerFunc {
	return r.Handler
}

// GetPriority возвращает приоритет
func (r RoutePattern) GetPriority() int {
	return r.Priority
}

// GetType возвращает тип маршрута
func (r RoutePattern) GetType() string {
	return string(r.Type)
}

// Compile компилирует паттерны в регулярные выражения
func (r *RoutePattern) Compile() error {
	r.compiled = make([]*regexp.Regexp, 0, len(r.Patterns))

	for _, pattern := range r.Patterns {
		// Wildcard паттерн
		if pattern == "*" {
			// Wildcard матчит все
			r.compiled = append(r.compiled, regexp.MustCompile(".*"))
			continue
		}

		// Преобразуем паттерн в regex
		regexPattern := r.patternToRegex(pattern)

		compiled, err := regexp.Compile(regexPattern)
		if err != nil {
			return err
		}

		r.compiled = append(r.compiled, compiled)
	}

	return nil
}

// Match проверяет соответствие текста паттерну
func (r *RoutePattern) Match(text string) (bool, map[string]string) {
	text = strings.TrimSpace(strings.ToLower(text))

	// Если нет скомпилированных паттернов, компилируем
	if len(r.compiled) == 0 {
		if err := r.Compile(); err != nil {
			return false, nil
		}
	}

	// Проверяем каждый паттерн
	for i, re := range r.compiled {
		if matches := re.FindStringSubmatch(text); matches != nil {
			// Извлекаем именованные группы
			params := make(map[string]string)

			// Если это простой паттерн без групп
			if len(matches) == 1 {
				return true, params
			}

			// Извлекаем параметры из именованных групп
			for i, name := range re.SubexpNames() {
				if i > 0 && i < len(matches) && name != "" {
					params[name] = matches[i]
				}
			}

			// Также добавляем позиционные параметры
			for i := 1; i < len(matches); i++ {
				params[string(rune('0'+i))] = matches[i]
			}

			// Сохраняем оригинальный паттерн
			params["_pattern"] = r.Patterns[i]

			return true, params
		}
	}

	return false, nil
}

// patternToRegex преобразует паттерн в регулярное выражение
func (r *RoutePattern) patternToRegex(pattern string) string {
	// Escape специальные символы regex
	pattern = regexp.QuoteMeta(pattern)

	// Заменяем плейсхолдеры на группы захвата
	// {id} -> (?P<id>\d+)
	// {name} -> (?P<name>\w+)
	// {text} -> (?P<text>.+)

	replacements := map[string]string{
		`\{id\}`:     `(?P<id>\d+)`,
		`\{user\}`:   `(?P<user>\d+)`,
		`\{name\}`:   `(?P<name>\w+)`,
		`\{text\}`:   `(?P<text>.+)`,
		`\{amount\}`: `(?P<amount>\d+)`,
		`\{any\}`:    `(?P<any>.+)`,
	}

	for old, new := range replacements {
		pattern = regexp.MustCompile(old).ReplaceAllString(pattern, new)
	}

	// Добавляем якоря начала и конца
	return "^" + pattern + "$"
}

// MatchType проверяет соответствие типа маршрута контексту
func (r *RoutePattern) MatchType(ctx core.UniversalContext) bool {
	switch r.Type {
	case RouteTypeCommand:
		return ctx.IsCommand()
	case RouteTypeCallback:
		return ctx.IsCallback()
	case RouteTypeMessage:
		return ctx.IsMessage()
	case RouteTypeRegex, RouteTypeWildcard:
		return true
	default:
		return false
	}
}

// CheckSecurity проверяет правила безопасности
func (r *RoutePattern) CheckSecurity(ctx core.UniversalContext) error {
	return r.Security.Check(ctx)
}

// Execute выполняет обработчик маршрута
func (r *RoutePattern) Execute(ctx core.UniversalContext) core.Response {
	// Проверяем безопасность
	if err := r.CheckSecurity(ctx); err != nil {
		return core.NewMessage("🚫 " + err.Error())
	}

	// Выполняем обработчик
	return r.Handler(ctx)
}

// RouteBuilder построитель маршрутов
type RouteBuilder struct {
	pattern *RoutePattern
}

// NewRoute создает новый построитель маршрута
func NewRoute(patterns ...string) *RouteBuilder {
	return &RouteBuilder{
		pattern: &RoutePattern{
			Patterns: patterns,
			Priority: 50,
			Type:     RouteTypeCommand,
			Security: SecurityRule{},
			Meta:     RouteMeta{},
		},
	}
}

// Handler устанавливает обработчик
func (b *RouteBuilder) Handler(h core.HandlerFunc) *RouteBuilder {
	b.pattern.Handler = h
	return b
}

// Priority устанавливает приоритет
func (b *RouteBuilder) Priority(p int) *RouteBuilder {
	b.pattern.Priority = p
	return b
}

// Type устанавливает тип маршрута
func (b *RouteBuilder) Type(t RouteType) *RouteBuilder {
	b.pattern.Type = t
	return b
}

// RequireAuth требует аутентификации
func (b *RouteBuilder) RequireAuth() *RouteBuilder {
	b.pattern.Security.RequireAuth = true
	return b
}

// RequireRoles требует роли
func (b *RouteBuilder) RequireRoles(roles ...string) *RouteBuilder {
	b.pattern.Security.RequireRoles = roles
	return b
}

// RequirePermissions требует права
func (b *RouteBuilder) RequirePermissions(perms ...string) *RouteBuilder {
	b.pattern.Security.RequirePermissions = perms
	return b
}

// RateLimit устанавливает ограничение скорости
func (b *RouteBuilder) RateLimit(requests, window int) *RouteBuilder {
	b.pattern.Security.RateLimit = &RateLimitConfig{
		Requests: requests,
		Window:   window,
	}
	return b
}

// Meta устанавливает метаданные
func (b *RouteBuilder) Meta(name, description string) *RouteBuilder {
	b.pattern.Meta.Name = name
	b.pattern.Meta.Description = description
	return b
}

// Tags устанавливает теги
func (b *RouteBuilder) Tags(tags ...string) *RouteBuilder {
	b.pattern.Meta.Tags = tags
	return b
}

// Hidden скрывает маршрут
func (b *RouteBuilder) Hidden() *RouteBuilder {
	b.pattern.Meta.Hidden = true
	return b
}

// Build возвращает готовый паттерн
func (b *RouteBuilder) Build() RoutePattern {
	return *b.pattern
}
