package http

import (
	"encoding/json"
	"fmt"
	"github.com/andranikuz/botkit/core"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Adapter HTTP адаптер для REST API
type Adapter struct {
	router      core.Router
	httpRouter  *mux.Router
	logger      core.Logger
	config      core.Config
	middlewares []mux.MiddlewareFunc
}

// NewAdapter создает новый HTTP адаптер
func NewAdapter(logger core.Logger, config core.Config) *Adapter {
	return &Adapter{
		httpRouter:  mux.NewRouter(),
		logger:      logger,
		config:      config,
		middlewares: make([]mux.MiddlewareFunc, 0),
	}
}

// UseRouter устанавливает модульный роутер
func (a *Adapter) UseRouter(router core.Router) {
	a.router = router
	a.registerRoutes()
}

// Use добавляет HTTP middleware
func (a *Adapter) Use(middleware mux.MiddlewareFunc) {
	a.middlewares = append(a.middlewares, middleware)
	a.httpRouter.Use(middleware)
}

// ServeHTTP реализует http.Handler
func (a *Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.httpRouter.ServeHTTP(w, r)
}

// ListenAndServe запускает HTTP сервер
func (a *Adapter) ListenAndServe(addr string) error {
	// Добавляем стандартные middleware
	a.Use(a.loggingMiddleware)
	a.Use(a.corsMiddleware)
	a.Use(a.recoveryMiddleware)

	a.logger.Info("Starting HTTP server", "addr", addr)
	return http.ListenAndServe(addr, a)
}

// registerRoutes регистрирует HTTP маршруты
func (a *Adapter) registerRoutes() {
	// API endpoints для модулей
	a.httpRouter.HandleFunc("/api/v1/modules", a.handleListModules).Methods("GET")
	a.httpRouter.HandleFunc("/api/v1/modules/{module}/execute", a.handleExecute).Methods("POST")

	// Регистрируем специфичные API endpoints модулей
	for _, module := range a.router.ListModules() {
		if apiModule, ok := module.(core.APIModule); ok {
			a.registerModuleAPI(apiModule)
		}
	}

	// Health check
	a.httpRouter.HandleFunc("/health", a.handleHealth).Methods("GET")

	// WebSocket endpoint
	a.httpRouter.HandleFunc("/ws", a.handleWebSocket)
}

// registerModuleAPI регистрирует API endpoints модуля
func (a *Adapter) registerModuleAPI(module core.APIModule) {
	for _, handler := range module.APIHandlers() {
		path := fmt.Sprintf("/api/v1/%s%s", module.Name(), handler.Path)

		a.httpRouter.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			// Создаем контекст
			ctx := a.requestToContext(r)

			// Создаем API request
			apiReq := core.APIRequest{
				Method:  r.Method,
				Path:    r.URL.Path,
				Headers: a.headersToMap(r.Header),
				Query:   a.queryToMap(r.URL.Query()),
				UserID:  ctx.GetUserID(),
				Context: r.Context(),
			}

			// Парсим body если есть
			if r.Body != nil {
				defer r.Body.Close()
				if err := json.NewDecoder(r.Body).Decode(&apiReq.Body); err != nil {
					a.sendError(w, err, http.StatusBadRequest)
					return
				}
			}

			// Вызываем handler
			resp, err := handler.Handler(r.Context(), apiReq)
			if err != nil {
				a.sendError(w, err, http.StatusInternalServerError)
				return
			}

			// Отправляем ответ
			a.sendResponse(w, resp)
		}).Methods(handler.Method)

		a.logger.Info("Registered API endpoint",
			"module", module.Name(),
			"method", handler.Method,
			"path", path,
		)
	}
}

// handleListModules обрабатывает запрос списка модулей
func (a *Adapter) handleListModules(w http.ResponseWriter, r *http.Request) {
	modules := a.router.ListModules()

	result := make([]map[string]interface{}, 0, len(modules))
	for _, module := range modules {
		info := map[string]interface{}{
			"name":    module.Name(),
			"version": module.Version(),
		}

		// Добавляем информацию о маршрутах
		routes := module.Routes()
		info["routes_count"] = len(routes)

		// Проверяем специальные типы
		if _, ok := module.(core.WildcardModule); ok {
			info["type"] = "wildcard"
		} else if _, ok := module.(core.APIModule); ok {
			info["type"] = "api"
		} else if _, ok := module.(core.EventAwareModule); ok {
			info["type"] = "event_aware"
		} else {
			info["type"] = "standard"
		}

		result = append(result, info)
	}

	a.sendJSON(w, result)
}

// handleExecute обрабатывает выполнение команды модуля
func (a *Adapter) handleExecute(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// moduleName := vars["module"] // TODO: использовать для фильтрации модулей

	// Парсим тело запроса
	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.sendError(w, err, http.StatusBadRequest)
		return
	}

	// Создаем контекст из запроса
	ctx := a.createContext(r, req)

	// Роутим через основной роутер
	response := a.router.Route(ctx)

	// Конвертируем ответ в HTTP response
	a.sendModuleResponse(w, response)
}

// handleHealth обрабатывает health check
func (a *Adapter) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":  "healthy",
		"time":    time.Now().Unix(),
		"modules": len(a.router.ListModules()),
	}

	a.sendJSON(w, health)
}

// handleWebSocket обрабатывает WebSocket соединения
func (a *Adapter) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket handling
	http.Error(w, "WebSocket not implemented", http.StatusNotImplemented)
}

// requestToContext конвертирует HTTP запрос в UniversalContext
func (a *Adapter) requestToContext(r *http.Request) core.UniversalContext {
	ctx := core.NewBaseContext(r.Context())

	// Устанавливаем источник
	ctx.SetSource("http")

	// Извлекаем user ID из заголовков или JWT
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		if id, err := strconv.ParseInt(userID, 10, 64); err == nil {
			ctx.SetUserID(id)
		}
	}

	// Извлекаем chat ID
	if chatID := r.Header.Get("X-Chat-ID"); chatID != "" {
		if id, err := strconv.ParseInt(chatID, 10, 64); err == nil {
			ctx.SetChatID(id)
		}
	}

	// Устанавливаем locale
	if locale := r.Header.Get("Accept-Language"); locale != "" {
		ctx.SetLocale(strings.Split(locale, ",")[0])
	}

	// Сохраняем оригинальный запрос
	ctx.SetOriginal(r)

	return ctx
}

// createContext создает контекст из ExecuteRequest
func (a *Adapter) createContext(r *http.Request, req ExecuteRequest) core.UniversalContext {
	ctx := a.requestToContext(r)

	// Устанавливаем данные из запроса
	if baseCtx, ok := ctx.(*core.BaseContext); ok {
		baseCtx.SetUserID(req.UserID)
		baseCtx.SetChatID(req.ChatID)
		baseCtx.SetText(req.Text)
		baseCtx.SetMessageID(req.MessageID)

		// Устанавливаем тип
		if req.IsCallback {
			baseCtx.SetIsCallback(true)
		} else {
			baseCtx.SetIsCommand(strings.HasPrefix(req.Text, "/"))
		}
	}

	// Устанавливаем параметры
	for k, v := range req.Params {
		ctx.SetParam(k, v)
	}

	// Устанавливаем дополнительные данные
	for k, v := range req.Data {
		ctx.Set(k, v)
	}

	return ctx
}

// sendModuleResponse отправляет ответ модуля
func (a *Adapter) sendModuleResponse(w http.ResponseWriter, response core.Response) {
	result := ModuleResponseDTO{
		Type: string(response.Type()),
		Content: ContentDTO{
			Text:      response.Content().Text,
			ParseMode: string(response.Content().ParseMode),
		},
		Options: response.Options(),
	}

	// Конвертируем медиа
	for _, media := range response.Content().Media {
		result.Content.Media = append(result.Content.Media, MediaDTO{
			Type:   string(media.Type),
			FileID: media.FileID,
			URL:    media.URL,
		})
	}

	// Конвертируем клавиатуру
	if kb := response.Content().Keyboard; kb != nil {
		result.Content.Keyboard = a.convertKeyboard(kb)
	}

	// Для множественных ответов
	if response.Type() == core.ResponseTypeMultiple {
		for _, action := range response.Actions() {
			// Рекурсивно конвертируем действия
			// Упрощенная версия - отправляем только первое действие
			a.sendModuleResponse(w, action)
			return
		}
	}

	a.sendJSON(w, result)
}

// convertKeyboard конвертирует клавиатуру в DTO
func (a *Adapter) convertKeyboard(kb core.Keyboard) interface{} {
	buttons := kb.Buttons()
	result := make([][]ButtonDTO, 0, len(buttons))

	for _, row := range buttons {
		dtoRow := make([]ButtonDTO, 0, len(row))
		for _, btn := range row {
			dtoRow = append(dtoRow, ButtonDTO{
				Text: btn.Text,
				Type: string(btn.Type),
				Data: btn.Data,
			})
		}
		result = append(result, dtoRow)
	}

	return result
}

// Middleware

func (a *Adapter) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		a.logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", time.Since(start).Milliseconds(),
			"ip", r.RemoteAddr,
		)
	})
}

func (a *Adapter) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Chat-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Adapter) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				a.logger.Error("Panic recovered", "error", err, "path", r.URL.Path)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Helper methods

func (a *Adapter) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (a *Adapter) sendError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}

func (a *Adapter) sendResponse(w http.ResponseWriter, resp core.APIResponse) {
	// Set headers
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}

	// Set status
	if resp.Status == 0 {
		resp.Status = http.StatusOK
	}
	w.WriteHeader(resp.Status)

	// Send body
	if resp.Body != nil {
		json.NewEncoder(w).Encode(resp.Body)
	}
}

func (a *Adapter) headersToMap(headers http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

func (a *Adapter) queryToMap(values map[string][]string) map[string]string {
	result := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
