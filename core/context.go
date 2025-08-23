package core

import (
	"context"
	"time"
)

// UniversalContext универсальный контекст для обработки сообщений
type UniversalContext interface {
	// Context возвращает базовый контекст Go
	Context() context.Context
	
	// === Идентификаторы ===
	
	// GetUserID возвращает ID пользователя
	GetUserID() int64
	
	// GetChatID возвращает ID чата
	GetChatID() int64
	
	// GetMessageID возвращает ID сообщения
	GetMessageID() string
	
	// GetThreadID возвращает ID треда (для групповых чатов)
	GetThreadID() string
	
	// === Данные пользователя ===
	
	// GetUsername возвращает username пользователя
	GetUsername() string
	
	// GetFirstName возвращает имя пользователя
	GetFirstName() string
	
	// GetLastName возвращает фамилию пользователя
	GetLastName() string
	
	// GetProfile возвращает профиль пользователя (если загружен)
	GetProfile() *Profile
	
	// === Контент сообщения ===
	
	// GetText возвращает текст сообщения
	GetText() string
	
	// GetData возвращает дополнительные данные (callback data, etc)
	GetData() map[string]interface{}
	
	// === Тип сообщения ===
	
	// IsCommand проверяет, является ли сообщение командой
	IsCommand() bool
	
	// IsCallback проверяет, является ли сообщение callback'ом
	IsCallback() bool
	
	// IsMessage проверяет, является ли это обычным сообщением
	IsMessage() bool
	
	// === Медиа ===
	
	// HasMedia проверяет наличие медиа в сообщении
	HasMedia() bool
	
	// GetMedia возвращает список медиа файлов
	GetMedia() []Media
	
	// === Параметры роутинга ===
	
	// GetParam получает параметр из роута
	GetParam(key string) (interface{}, bool)
	
	// SetParam устанавливает параметр в контекст
	SetParam(key string, value interface{})
	
	// GetIntParam получает int параметр
	GetIntParam(key string) (int, bool)
	
	// GetStringParam получает string параметр
	GetStringParam(key string) (string, bool)
	
	// === Безопасность ===
	
	// HasPermission проверяет наличие прав
	HasPermission(perm string) bool
	
	// GetRoles возвращает роли пользователя
	GetRoles() []string
	
	// IsAuthenticated проверяет, аутентифицирован ли пользователь
	IsAuthenticated() bool
	
	// === Метаданные ===
	
	// GetSource возвращает источник сообщения (telegram, api, websocket)
	GetSource() string
	
	// GetLocale возвращает локаль пользователя
	GetLocale() string
	
	// GetTimestamp возвращает время получения сообщения
	GetTimestamp() time.Time
	
	// === Транспорт-специфичные данные ===
	
	// GetOriginal возвращает оригинальное сообщение от транспорта
	GetOriginal() interface{}
	
	// Set устанавливает произвольное значение в контекст
	Set(key string, value interface{})
	
	// Get получает произвольное значение из контекста
	Get(key string) (interface{}, bool)
}

// BaseContext базовая реализация UniversalContext
type BaseContext struct {
	ctx        context.Context
	userID     int64
	chatID     int64
	messageID  string
	threadID   string
	username   string
	firstName  string
	lastName   string
	profile    *Profile
	text       string
	data       map[string]interface{}
	params     map[string]interface{}
	isCommand  bool
	isCallback bool
	media      []Media
	roles      []string
	source     string
	locale     string
	timestamp  time.Time
	original   interface{}
	values     map[string]interface{}
}

// NewBaseContext создает новый базовый контекст
func NewBaseContext(ctx context.Context) *BaseContext {
	return &BaseContext{
		ctx:       ctx,
		data:      make(map[string]interface{}),
		params:    make(map[string]interface{}),
		values:    make(map[string]interface{}),
		timestamp: time.Now(),
		locale:    "ru",
		source:    "unknown",
	}
}

// Context implementation
func (c *BaseContext) Context() context.Context                 { return c.ctx }
func (c *BaseContext) GetUserID() int64                         { return c.userID }
func (c *BaseContext) GetChatID() int64                         { return c.chatID }
func (c *BaseContext) GetMessageID() string                     { return c.messageID }
func (c *BaseContext) GetThreadID() string                      { return c.threadID }
func (c *BaseContext) GetUsername() string                      { return c.username }
func (c *BaseContext) GetFirstName() string                     { return c.firstName }
func (c *BaseContext) GetLastName() string                      { return c.lastName }
func (c *BaseContext) GetProfile() *Profile                     { return c.profile }
func (c *BaseContext) GetText() string                          { return c.text }
func (c *BaseContext) GetData() map[string]interface{}          { return c.data }
func (c *BaseContext) IsCommand() bool                          { return c.isCommand }
func (c *BaseContext) IsCallback() bool                         { return c.isCallback }
func (c *BaseContext) IsMessage() bool                          { return !c.isCommand && !c.isCallback }
func (c *BaseContext) HasMedia() bool                           { return len(c.media) > 0 }
func (c *BaseContext) GetMedia() []Media                        { return c.media }
func (c *BaseContext) GetSource() string                        { return c.source }
func (c *BaseContext) GetLocale() string                        { return c.locale }
func (c *BaseContext) GetTimestamp() time.Time                  { return c.timestamp }
func (c *BaseContext) GetOriginal() interface{}                 { return c.original }
func (c *BaseContext) GetRoles() []string                       { return c.roles }
func (c *BaseContext) IsAuthenticated() bool                    { return c.userID > 0 }

func (c *BaseContext) GetParam(key string) (interface{}, bool) {
	val, ok := c.params[key]
	return val, ok
}

func (c *BaseContext) SetParam(key string, value interface{}) {
	c.params[key] = value
}

func (c *BaseContext) GetIntParam(key string) (int, bool) {
	if val, ok := c.params[key]; ok {
		if i, ok := val.(int); ok {
			return i, true
		}
	}
	return 0, false
}

func (c *BaseContext) GetStringParam(key string) (string, bool) {
	if val, ok := c.params[key]; ok {
		if s, ok := val.(string); ok {
			return s, true
		}
	}
	return "", false
}

func (c *BaseContext) HasPermission(perm string) bool {
	// TODO: implement permission check
	return true
}

func (c *BaseContext) Set(key string, value interface{}) {
	c.values[key] = value
}

func (c *BaseContext) Get(key string) (interface{}, bool) {
	val, ok := c.values[key]
	return val, ok
}

// Setters for BaseContext
func (c *BaseContext) SetUserID(id int64)           { c.userID = id }
func (c *BaseContext) SetChatID(id int64)           { c.chatID = id }
func (c *BaseContext) SetMessageID(id string)       { c.messageID = id }
func (c *BaseContext) SetThreadID(id string)        { c.threadID = id }
func (c *BaseContext) SetUsername(name string)      { c.username = name }
func (c *BaseContext) SetFirstName(name string)     { c.firstName = name }
func (c *BaseContext) SetLastName(name string)      { c.lastName = name }
func (c *BaseContext) SetProfile(p *Profile)        { c.profile = p }
func (c *BaseContext) SetText(text string)          { c.text = text }
func (c *BaseContext) SetIsCommand(v bool)          { c.isCommand = v }
func (c *BaseContext) SetIsCallback(v bool)         { c.isCallback = v }
func (c *BaseContext) SetMedia(media []Media)       { c.media = media }
func (c *BaseContext) SetRoles(roles []string)      { c.roles = roles }
func (c *BaseContext) SetSource(source string)      { c.source = source }
func (c *BaseContext) SetLocale(locale string)      { c.locale = locale }
func (c *BaseContext) SetOriginal(orig interface{}) { c.original = orig }

// Media представляет медиа файл
type Media struct {
	Type      MediaType `json:"type"`
	FileID    string    `json:"file_id"`
	URL       string    `json:"url"`
	Caption   string    `json:"caption"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	Thumbnail string    `json:"thumbnail"`
}

// MediaType тип медиа
type MediaType string

const (
	MediaTypePhoto    MediaType = "photo"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
	MediaTypeVoice    MediaType = "voice"
	MediaTypeSticker  MediaType = "sticker"
)

// Profile профиль пользователя
type Profile struct {
	ID           int64             `json:"id"`
	Username     string            `json:"username"`
	FirstName    string            `json:"first_name"`
	LastName     string            `json:"last_name"`
	Balance      int64             `json:"balance"`
	Level        int               `json:"level"`
	Experience   int               `json:"experience"`
	Rating       int               `json:"rating"`
	Roles        []string          `json:"roles"`
	Permissions  []string          `json:"permissions"`
	Metadata     map[string]string `json:"metadata"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	LastActiveAt time.Time         `json:"last_active_at"`
}