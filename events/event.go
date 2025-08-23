package events

import (
	"time"
)

// Event базовая реализация события
type Event struct {
	eventType string
	timestamp int64
	source    string
	data      map[string]interface{}
	userID    int64
	chatID    int64
}

// NewEvent создает новое событие
func NewEvent(eventType, source string) *Event {
	return &Event{
		eventType: eventType,
		timestamp: time.Now().Unix(),
		source:    source,
		data:      make(map[string]interface{}),
	}
}

// Type возвращает тип события
func (e *Event) Type() string {
	return e.eventType
}

// Timestamp возвращает время события
func (e *Event) Timestamp() int64 {
	return e.timestamp
}

// Source возвращает источник события
func (e *Event) Source() string {
	return e.source
}

// Data возвращает данные события
func (e *Event) Data() map[string]interface{} {
	return e.data
}

// UserID возвращает ID пользователя
func (e *Event) UserID() int64 {
	return e.userID
}

// ChatID возвращает ID чата
func (e *Event) ChatID() int64 {
	return e.chatID
}

// SetUserID устанавливает ID пользователя
func (e *Event) SetUserID(id int64) *Event {
	e.userID = id
	return e
}

// SetChatID устанавливает ID чата
func (e *Event) SetChatID(id int64) *Event {
	e.chatID = id
	return e
}

// SetData устанавливает данные события
func (e *Event) SetData(key string, value interface{}) *Event {
	e.data[key] = value
	return e
}

// GetData получает данные по ключу
func (e *Event) GetData(key string) (interface{}, bool) {
	val, ok := e.data[key]
	return val, ok
}

// === Специфичные события ===

// MessageReceivedEvent событие получения сообщения
type MessageReceivedEvent struct {
	*Event
	MessageID string
	Text      string
	Media     []string
}

// NewMessageReceivedEvent создает событие получения сообщения
func NewMessageReceivedEvent(userID, chatID int64, messageID, text string) *MessageReceivedEvent {
	event := NewEvent("message.received", "router")
	event.SetUserID(userID).SetChatID(chatID)
	
	return &MessageReceivedEvent{
		Event:     event,
		MessageID: messageID,
		Text:      text,
	}
}

// CommandExecutedEvent событие выполнения команды
type CommandExecutedEvent struct {
	*Event
	Command string
	Module  string
	Success bool
	Error   error
}

// NewCommandExecutedEvent создает событие выполнения команды
func NewCommandExecutedEvent(userID int64, command, module string, success bool) *CommandExecutedEvent {
	event := NewEvent("command.executed", module)
	event.SetUserID(userID)
	
	return &CommandExecutedEvent{
		Event:   event,
		Command: command,
		Module:  module,
		Success: success,
	}
}

// UserActionEvent событие действия пользователя
type UserActionEvent struct {
	*Event
	Action   string
	Resource string
	Result   string
}

// NewUserActionEvent создает событие действия пользователя
func NewUserActionEvent(userID int64, action, resource, result string) *UserActionEvent {
	event := NewEvent("user.action", "user")
	event.SetUserID(userID)
	
	return &UserActionEvent{
		Event:    event,
		Action:   action,
		Resource: resource,
		Result:   result,
	}
}

// StateChangedEvent событие изменения состояния
type StateChangedEvent struct {
	*Event
	Entity   string
	EntityID string
	OldState string
	NewState string
}

// NewStateChangedEvent создает событие изменения состояния
func NewStateChangedEvent(entity, entityID, oldState, newState string) *StateChangedEvent {
	event := NewEvent("state.changed", "system")
	
	return &StateChangedEvent{
		Event:    event,
		Entity:   entity,
		EntityID: entityID,
		OldState: oldState,
		NewState: newState,
	}
}

// ErrorOccurredEvent событие ошибки
type ErrorOccurredEvent struct {
	*Event
	ErrorCode    string
	ErrorMessage string
	Module       string
	Severity     string
}

// NewErrorOccurredEvent создает событие ошибки
func NewErrorOccurredEvent(module, code, message, severity string) *ErrorOccurredEvent {
	event := NewEvent("error.occurred", module)
	
	return &ErrorOccurredEvent{
		Event:        event,
		ErrorCode:    code,
		ErrorMessage: message,
		Module:       module,
		Severity:     severity,
	}
}

// WildcardMessageEvent событие wildcard сообщения
type WildcardMessageEvent struct {
	*Event
	Text     string
	Purpose  string
	Metadata map[string]interface{}
}

// NewWildcardMessageEvent создает событие wildcard сообщения
func NewWildcardMessageEvent(userID, chatID int64, text, purpose string) *WildcardMessageEvent {
	event := NewEvent("wildcard.message", "wildcard")
	event.SetUserID(userID).SetChatID(chatID)
	
	return &WildcardMessageEvent{
		Event:    event,
		Text:     text,
		Purpose:  purpose,
		Metadata: make(map[string]interface{}),
	}
}

// ModuleLifecycleEvent событие жизненного цикла модуля
type ModuleLifecycleEvent struct {
	*Event
	Module    string
	Lifecycle string // started, stopped, error, health_check
	Status    string
	Details   map[string]string
}

// NewModuleLifecycleEvent создает событие жизненного цикла модуля
func NewModuleLifecycleEvent(module, lifecycle, status string) *ModuleLifecycleEvent {
	event := NewEvent("module.lifecycle", module)
	
	return &ModuleLifecycleEvent{
		Event:     event,
		Module:    module,
		Lifecycle: lifecycle,
		Status:    status,
		Details:   make(map[string]string),
	}
}