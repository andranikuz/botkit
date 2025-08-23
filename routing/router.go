package routing

import (
	"context"
	"fmt"
	"github.com/andranikuz/botkit/core"
	"sort"
	"sync"
)

// Router основная реализация роутера
type Router struct {
	// modules зарегистрированные модули
	modules map[string]core.Module

	// routes скомпилированные маршруты
	routes []compiledRoute

	// wildcards wildcard обработчики
	wildcards []wildcardHandler

	// middlewares промежуточные обработчики
	middlewares []Middleware

	// eventBus шина событий
	eventBus core.EventBus

	// logger логгер
	logger core.Logger

	// config конфигурация
	config core.Config

	// dependencies зависимости для модулей
	dependencies core.Dependencies

	// started флаг запуска
	started bool

	// mu мьютекс для потокобезопасности
	mu sync.RWMutex
}

// compiledRoute скомпилированный маршрут
type compiledRoute struct {
	pattern  RoutePattern
	module   string
	compiled bool
}

// wildcardHandler обработчик wildcard
type wildcardHandler struct {
	module   core.WildcardModule
	priority int
}

// NewRouter создает новый роутер
func NewRouter(eventBus core.EventBus, logger core.Logger, config core.Config) *Router {
	return &Router{
		modules:     make(map[string]core.Module),
		routes:      make([]compiledRoute, 0),
		wildcards:   make([]wildcardHandler, 0),
		middlewares: make([]Middleware, 0),
		eventBus:    eventBus,
		logger:      logger,
		config:      config,
	}
}

// SetDependencies устанавливает зависимости для модулей
func (r *Router) SetDependencies(deps core.Dependencies) {
	r.dependencies = deps
}

// RegisterModule регистрирует модуль
func (r *Router) RegisterModule(module core.Module) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.started {
		return fmt.Errorf("cannot register module after router started")
	}

	name := module.Name()
	if _, exists := r.modules[name]; exists {
		return fmt.Errorf("module %s already registered", name)
	}

	// Инициализируем модуль
	if r.dependencies != nil {
		if err := module.Init(r.dependencies); err != nil {
			return fmt.Errorf("failed to init module %s: %w", name, err)
		}
	}

	// Сохраняем модуль
	r.modules[name] = module

	// Регистрируем маршруты
	for _, iPattern := range module.Routes() {
		// Приводим к нашему типу RoutePattern
		if pattern, ok := iPattern.(RoutePattern); ok {
			route := compiledRoute{
				pattern: pattern,
				module:  name,
			}
			// pattern.Module = name // TODO: добавить поле Module в RoutePattern если нужно
			r.routes = append(r.routes, route)
		}
	}

	// Регистрируем события если модуль поддерживает
	if eventAware, ok := module.(core.EventAwareModule); ok {
		for _, sub := range eventAware.Events() {
			if err := r.eventBus.Subscribe(sub.EventType, sub.Handler); err != nil {
				return fmt.Errorf("failed to subscribe to event %s: %w", sub.EventType, err)
			}
		}
	}

	r.logger.Info("Module registered", "name", name, "routes", len(module.Routes()))

	return nil
}

// RegisterWildcard регистрирует wildcard обработчик
func (r *Router) RegisterWildcard(module core.WildcardModule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.started {
		return fmt.Errorf("cannot register wildcard after router started")
	}

	// Регистрируем как обычный модуль
	if err := r.RegisterModule(module); err != nil {
		return err
	}

	// Добавляем в wildcard обработчики
	r.wildcards = append(r.wildcards, wildcardHandler{
		module:   module,
		priority: module.Priority(),
	})

	// Сортируем по приоритету (больше = выше)
	sort.Slice(r.wildcards, func(i, j int) bool {
		return r.wildcards[i].priority > r.wildcards[j].priority
	})

	r.logger.Info("Wildcard module registered", "name", module.Name(), "priority", module.Priority())

	return nil
}

// RegisterMiddleware регистрирует middleware
func (r *Router) RegisterMiddleware(mw Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.middlewares = append(r.middlewares, mw)

	// Сортируем по приоритету
	sort.Slice(r.middlewares, func(i, j int) bool {
		return r.middlewares[i].Priority() > r.middlewares[j].Priority()
	})

	r.logger.Info("Middleware registered", "name", mw.Name(), "priority", mw.Priority())
}

// Route маршрутизирует сообщение
func (r *Router) Route(ctx core.UniversalContext) core.Response {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Применяем middleware
	handler := r.routeInternal
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		mw := r.middlewares[i]
		nextHandler := handler
		handler = func(c core.UniversalContext) core.Response {
			return mw.Process(c, nextHandler)
		}
	}

	return handler(ctx)
}

// routeInternal внутренняя маршрутизация
func (r *Router) routeInternal(ctx core.UniversalContext) core.Response {
	// Получаем текст для матчинга
	text := ctx.GetText()
	if ctx.IsCallback() {
		if data, ok := ctx.GetData()["callback_data"]; ok {
			text = data.(string)
		}
	}

	// Компилируем маршруты при первом использовании
	r.compileRoutesOnce()

	// Ищем подходящий маршрут
	var matchedRoute *compiledRoute
	var matchedParams map[string]string

	// Сортируем маршруты по приоритету
	routes := make([]compiledRoute, len(r.routes))
	copy(routes, r.routes)
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].pattern.Priority > routes[j].pattern.Priority
	})

	// Проверяем маршруты
	for _, route := range routes {
		// Проверяем тип маршрута
		if !route.pattern.MatchType(ctx) {
			continue
		}

		// Проверяем паттерн
		if matched, params := route.pattern.Match(text); matched {
			matchedRoute = &route
			matchedParams = params
			break
		}
	}

	// Если маршрут найден
	if matchedRoute != nil {
		// Устанавливаем параметры в контекст
		for key, value := range matchedParams {
			ctx.SetParam(key, value)
		}

		// Логируем
		r.logger.Debug("Route matched",
			"module", matchedRoute.module,
			"pattern", matchedParams["_pattern"],
			"user", ctx.GetUserID(),
		)

		// Выполняем обработчик
		return matchedRoute.pattern.Execute(ctx)
	}

	// Проверяем wildcard обработчики
	for _, wc := range r.wildcards {
		if wc.module.ShouldHandle(ctx) {
			r.logger.Debug("Wildcard matched",
				"module", wc.module.Name(),
				"user", ctx.GetUserID(),
			)
			return wc.module.HandleWildcard(ctx)
		}
	}

	// Маршрут не найден
	r.logger.Debug("No route matched",
		"text", text,
		"user", ctx.GetUserID(),
	)

	return core.NewSilentResponse()
}

// compileRoutesOnce компилирует маршруты один раз
func (r *Router) compileRoutesOnce() {
	for i := range r.routes {
		if !r.routes[i].compiled {
			if err := r.routes[i].pattern.Compile(); err != nil {
				r.logger.Error("Failed to compile route", "error", err, "module", r.routes[i].module)
			}
			r.routes[i].compiled = true
		}
	}
}

// GetModule получает модуль по имени
func (r *Router) GetModule(name string) (core.Module, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, ok := r.modules[name]
	return module, ok
}

// ListModules возвращает список модулей
func (r *Router) ListModules() []core.Module {
	r.mu.RLock()
	defer r.mu.RUnlock()

	modules := make([]core.Module, 0, len(r.modules))
	for _, m := range r.modules {
		modules = append(modules, m)
	}

	return modules
}

// Start запускает роутер и все модули
func (r *Router) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.started {
		return fmt.Errorf("router already started")
	}

	// Запускаем все модули
	for name, module := range r.modules {
		if err := module.Start(ctx); err != nil {
			return fmt.Errorf("failed to start module %s: %w", name, err)
		}
		r.logger.Info("Module started", "name", name)
	}

	// Запускаем шину событий
	if r.eventBus != nil {
		if err := r.eventBus.Start(ctx); err != nil {
			return fmt.Errorf("failed to start event bus: %w", err)
		}
	}

	r.started = true
	r.logger.Info("Router started", "modules", len(r.modules))

	return nil
}

// Stop останавливает роутер и все модули
func (r *Router) Stop(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.started {
		return nil
	}

	// Останавливаем все модули
	for name, module := range r.modules {
		if err := module.Stop(ctx); err != nil {
			r.logger.Error("Failed to stop module", "name", name, "error", err)
		} else {
			r.logger.Info("Module stopped", "name", name)
		}
	}

	// Останавливаем шину событий
	if r.eventBus != nil {
		if err := r.eventBus.Stop(ctx); err != nil {
			r.logger.Error("Failed to stop event bus", "error", err)
		}
	}

	r.started = false
	r.logger.Info("Router stopped")

	return nil
}

// GetRoutes возвращает все зарегистрированные маршруты (для отладки)
func (r *Router) GetRoutes() []RouteInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]RouteInfo, 0, len(r.routes))
	for _, route := range r.routes {
		routes = append(routes, RouteInfo{
			Module:      route.module,
			Patterns:    route.pattern.Patterns,
			Type:        string(route.pattern.Type),
			Priority:    route.pattern.Priority,
			Description: route.pattern.Meta.Description,
		})
	}

	return routes
}

// RouteInfo информация о маршруте
type RouteInfo struct {
	Module      string   `json:"module"`
	Patterns    []string `json:"patterns"`
	Type        string   `json:"type"`
	Priority    int      `json:"priority"`
	Description string   `json:"description"`
}
