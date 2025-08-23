package http

import (
	"github.com/andranikuz/botkit/core"
)

// ExecuteRequest запрос на выполнение команды
type ExecuteRequest struct {
	UserID     int64                  `json:"user_id"`
	ChatID     int64                  `json:"chat_id"`
	MessageID  string                 `json:"message_id,omitempty"`
	Text       string                 `json:"text"`
	IsCallback bool                   `json:"is_callback,omitempty"`
	Params     map[string]interface{} `json:"params,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// ModuleResponseDTO ответ модуля для HTTP
type ModuleResponseDTO struct {
	Type    string               `json:"type"`
	Content ContentDTO           `json:"content"`
	Options core.ResponseOptions `json:"options,omitempty"`
	Actions []ModuleResponseDTO  `json:"actions,omitempty"`
}

// ContentDTO содержимое сообщения для HTTP
type ContentDTO struct {
	Text      string      `json:"text"`
	ParseMode string      `json:"parse_mode,omitempty"`
	Media     []MediaDTO  `json:"media,omitempty"`
	Keyboard  interface{} `json:"keyboard,omitempty"`
}

// MediaDTO медиа для HTTP
type MediaDTO struct {
	Type     string `json:"type"`
	FileID   string `json:"file_id,omitempty"`
	URL      string `json:"url,omitempty"`
	Caption  string `json:"caption,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

// ButtonDTO кнопка для HTTP
type ButtonDTO struct {
	Text  string                 `json:"text"`
	Type  string                 `json:"type"`
	Data  string                 `json:"data,omitempty"`
	Route map[string]interface{} `json:"route,omitempty"`
}

// ErrorResponse ответ с ошибкой
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse успешный ответ
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ModuleInfoDTO информация о модуле
type ModuleInfoDTO struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Type        string   `json:"type"`
	RoutesCount int      `json:"routes_count"`
	Routes      []string `json:"routes,omitempty"`
}

// HealthDTO health check response
type HealthDTO struct {
	Status  string `json:"status"`
	Time    int64  `json:"time"`
	Modules int    `json:"modules"`
	Version string `json:"version,omitempty"`
}
