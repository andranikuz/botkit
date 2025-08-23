package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/andranikuz/botkit/core"
	"github.com/gorilla/websocket"
)

// Adapter WebSocket адаптер
type Adapter struct {
	router      core.Router
	logger      core.Logger
	config      core.Config
	upgrader    websocket.Upgrader
	connections map[string]*Connection
	mu          sync.RWMutex
}

// Connection представляет WebSocket соединение
type Connection struct {
	ID      string
	UserID  int64
	Conn    *websocket.Conn
	Send    chan []byte
	Hub     *Adapter
	Context context.Context
	Cancel  context.CancelFunc
}

// Message сообщение WebSocket
type Message struct {
	Type   string                 `json:"type"`
	ID     string                 `json:"id,omitempty"`
	UserID int64                  `json:"user_id,omitempty"`
	ChatID int64                  `json:"chat_id,omitempty"`
	Text   string                 `json:"text,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

// NewAdapter создает новый WebSocket адаптер
func NewAdapter(logger core.Logger, config core.Config) *Adapter {
	return &Adapter{
		logger: logger,
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// В продакшене нужно проверять origin
				return true
			},
		},
		connections: make(map[string]*Connection),
	}
}

// UseRouter устанавливает роутер
func (a *Adapter) UseRouter(router core.Router) {
	a.router = router
}

// ServeHTTP обрабатывает WebSocket соединения
func (a *Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Извлекаем user ID из заголовков или query params
	userID := a.extractUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Апгрейд соединения
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		a.logger.Error("Failed to upgrade connection", "error", err)
		return
	}

	// Создаем соединение
	ctx, cancel := context.WithCancel(context.Background())
	connection := &Connection{
		ID:      generateConnectionID(),
		UserID:  userID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Hub:     a,
		Context: ctx,
		Cancel:  cancel,
	}

	// Регистрируем соединение
	a.register(connection)

	// Запускаем горутины для чтения и записи
	go connection.readPump()
	go connection.writePump()

	// Отправляем приветственное сообщение
	welcome := Message{
		Type: "connected",
		ID:   connection.ID,
		Data: map[string]interface{}{
			"version": "1.0.0",
			"time":    time.Now().Unix(),
		},
	}
	connection.sendMessage(welcome)
}

// register регистрирует соединение
func (a *Adapter) register(conn *Connection) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.connections[conn.ID] = conn
	a.logger.Info("WebSocket connection registered",
		"id", conn.ID,
		"user", conn.UserID,
	)
}

// unregister удаляет соединение
func (a *Adapter) unregister(conn *Connection) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.connections[conn.ID]; ok {
		delete(a.connections, conn.ID)
		close(conn.Send)
		conn.Cancel()

		a.logger.Info("WebSocket connection unregistered",
			"id", conn.ID,
			"user", conn.UserID,
		)
	}
}

// readPump читает сообщения от клиента
func (c *Connection) readPump() {
	defer func() {
		c.Hub.unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error("WebSocket error", "error", err)
			}
			break
		}

		// Парсим сообщение
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			c.sendError("Invalid message format")
			continue
		}

		// Обрабатываем сообщение
		c.handleMessage(msg)
	}
}

// writePump отправляет сообщения клиенту
func (c *Connection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.Context.Done():
			return
		}
	}
}

// handleMessage обрабатывает входящее сообщение
func (c *Connection) handleMessage(msg Message) {
	// Создаем универсальный контекст
	ctx := c.messageToContext(msg)

	// Роутим через основной роутер
	if c.Hub.router != nil {
		response := c.Hub.router.Route(ctx)
		c.sendResponse(msg.ID, response)
	} else {
		c.sendError("Router not configured")
	}
}

// messageToContext конвертирует WebSocket сообщение в UniversalContext
func (c *Connection) messageToContext(msg Message) core.UniversalContext {
	ctx := core.NewBaseContext(c.Context)

	ctx.SetSource("websocket")
	ctx.SetUserID(c.UserID)
	ctx.SetChatID(msg.ChatID)
	ctx.SetMessageID(msg.ID)
	ctx.SetText(msg.Text)

	// Устанавливаем тип сообщения
	switch msg.Type {
	case "command":
		ctx.SetIsCommand(true)
	case "callback":
		ctx.SetIsCallback(true)
	}

	// Добавляем данные
	for k, v := range msg.Data {
		ctx.Set(k, v)
	}

	// Сохраняем оригинальное сообщение
	ctx.SetOriginal(msg)

	return ctx
}

// sendResponse отправляет ответ клиенту
func (c *Connection) sendResponse(requestID string, response core.Response) {
	msg := Message{
		Type: "response",
		ID:   requestID,
	}

	switch response.Type() {
	case core.ResponseTypeMessage:
		msg.Data = map[string]interface{}{
			"text":       response.Content().Text,
			"parse_mode": response.Content().ParseMode,
		}

	case core.ResponseTypeEdit:
		msg.Data = map[string]interface{}{
			"action":     "edit",
			"message_id": response.Options().MessageToEditID,
			"text":       response.Content().Text,
		}

	case core.ResponseTypeDelete:
		msg.Data = map[string]interface{}{
			"action":     "delete",
			"message_id": response.Options().MessageToDeleteID,
		}

	case core.ResponseTypeSilent:
		// Не отправляем ничего
		return

	case core.ResponseTypeMultiple:
		// Отправляем каждое действие отдельно
		for _, action := range response.Actions() {
			c.sendResponse(requestID, action)
		}
		return
	}

	c.sendMessage(msg)
}

// sendMessage отправляет сообщение клиенту
func (c *Connection) sendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		c.Hub.logger.Error("Failed to marshal message", "error", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		// Канал переполнен, закрываем соединение
		c.Hub.unregister(c)
		c.Conn.Close()
	}
}

// sendError отправляет сообщение об ошибке
func (c *Connection) sendError(errMsg string) {
	msg := Message{
		Type:  "error",
		Error: errMsg,
	}
	c.sendMessage(msg)
}

// Broadcast отправляет сообщение всем подключенным клиентам
func (a *Adapter) Broadcast(msg Message) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, conn := range a.connections {
		conn.sendMessage(msg)
	}
}

// SendToUser отправляет сообщение конкретному пользователю
func (a *Adapter) SendToUser(userID int64, msg Message) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, conn := range a.connections {
		if conn.UserID == userID {
			conn.sendMessage(msg)
		}
	}
}

// GetConnections возвращает количество активных соединений
func (a *Adapter) GetConnections() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.connections)
}

// extractUserID извлекает user ID из запроса
func (a *Adapter) extractUserID(r *http.Request) int64 {
	// Пробуем из заголовка
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		var id int64
		fmt.Sscanf(userID, "%d", &id)
		return id
	}

	// Пробуем из query параметров
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		var id int64
		fmt.Sscanf(userID, "%d", &id)
		return id
	}

	// TODO: Извлечь из JWT токена

	return 0
}

// generateConnectionID генерирует уникальный ID соединения
func generateConnectionID() string {
	return fmt.Sprintf("ws_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
}

// WebSocketHandler HTTP handler для WebSocket endpoint
func (a *Adapter) WebSocketHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.ServeHTTP(w, r)
	}
}
