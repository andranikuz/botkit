package core

// Response универсальный ответ модуля
type Response interface {
	// Type возвращает тип ответа
	Type() ResponseType
	
	// Content возвращает контент сообщения
	Content() MessageContent
	
	// Options возвращает опции ответа
	Options() ResponseOptions
	
	// Actions возвращает дополнительные действия (для множественных ответов)
	Actions() []Response
	
	// IsEmpty проверяет, пустой ли ответ
	IsEmpty() bool
	
	// IsSilent проверяет, тихий ли ответ
	IsSilent() bool
}

// ResponseType тип ответа
type ResponseType string

const (
	// ResponseTypeMessage отправить новое сообщение
	ResponseTypeMessage ResponseType = "message"
	
	// ResponseTypeEdit редактировать существующее сообщение
	ResponseTypeEdit ResponseType = "edit"
	
	// ResponseTypeDelete удалить сообщение
	ResponseTypeDelete ResponseType = "delete"
	
	// ResponseTypeCallback ответить на callback query
	ResponseTypeCallback ResponseType = "callback"
	
	// ResponseTypeSilent не отправлять ответ
	ResponseTypeSilent ResponseType = "silent"
	
	// ResponseTypeMultiple несколько действий
	ResponseTypeMultiple ResponseType = "multiple"
	
	// ResponseTypeRedirect перенаправить на другой обработчик
	ResponseTypeRedirect ResponseType = "redirect"
	
	// ResponseTypeStream потоковый ответ (для больших данных)
	ResponseTypeStream ResponseType = "stream"
)

// MessageContent содержимое сообщения
type MessageContent struct {
	// Text текст сообщения
	Text string `json:"text"`
	
	// ParseMode режим парсинга (HTML, Markdown, Plain)
	ParseMode ParseMode `json:"parse_mode"`
	
	// Media медиа файлы
	Media []Media `json:"media,omitempty"`
	
	// Keyboard клавиатура
	Keyboard Keyboard `json:"keyboard,omitempty"`
	
	// Embeds встраиваемые элементы (для rich messages)
	Embeds []Embed `json:"embeds,omitempty"`
	
	// Metadata дополнительные данные
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ParseMode режим парсинга текста
type ParseMode string

const (
	ParseModeHTML     ParseMode = "HTML"
	ParseModeMarkdown ParseMode = "Markdown"
	ParseModePlain    ParseMode = "Plain"
)

// ResponseOptions опции ответа
type ResponseOptions struct {
	// ReplyToMessageID ID сообщения для ответа
	ReplyToMessageID string `json:"reply_to_message_id,omitempty"`
	
	// MessageToEditID ID сообщения для редактирования
	MessageToEditID string `json:"message_to_edit_id,omitempty"`
	
	// MessageToDeleteID ID сообщения для удаления
	MessageToDeleteID string `json:"message_to_delete_id,omitempty"`
	
	// DeleteUserMessage удалить сообщение пользователя
	DeleteUserMessage bool `json:"delete_user_message,omitempty"`
	
	// DisableNotification отключить уведомление
	DisableNotification bool `json:"disable_notification,omitempty"`
	
	// DisableWebPreview отключить превью ссылок
	DisableWebPreview bool `json:"disable_web_preview,omitempty"`
	
	// CallbackQueryID ID callback query для ответа
	CallbackQueryID string `json:"callback_query_id,omitempty"`
	
	// CallbackText текст для callback ответа
	CallbackText string `json:"callback_text,omitempty"`
	
	// ShowAlert показать alert для callback
	ShowAlert bool `json:"show_alert,omitempty"`
	
	// CacheTime время кеширования callback ответа
	CacheTime int `json:"cache_time,omitempty"`
	
	// TargetChatID ID чата для отправки (если отличается)
	TargetChatID int64 `json:"target_chat_id,omitempty"`
	
	// TargetUserID ID пользователя для отправки
	TargetUserID int64 `json:"target_user_id,omitempty"`
	
	// TTL время жизни сообщения
	TTL int `json:"ttl,omitempty"`
	
	// Priority приоритет отправки
	Priority int `json:"priority,omitempty"`
}

// Keyboard интерфейс клавиатуры
type Keyboard interface {
	// Type возвращает тип клавиатуры
	Type() KeyboardType
	
	// Buttons возвращает кнопки
	Buttons() [][]Button
	
	// Options возвращает опции клавиатуры
	Options() KeyboardOptions
}

// KeyboardType тип клавиатуры
type KeyboardType string

const (
	KeyboardTypeInline KeyboardType = "inline"
	KeyboardTypeReply  KeyboardType = "reply"
	KeyboardTypeRemove KeyboardType = "remove"
)

// Button кнопка клавиатуры
type Button struct {
	// Text текст кнопки
	Text string `json:"text"`
	
	// Type тип кнопки
	Type ButtonType `json:"type"`
	
	// Data данные кнопки (callback data, url, etc)
	Data string `json:"data,omitempty"`
	
	// Route типизированный роут для callback
	Route *Route `json:"route,omitempty"`
	
	// Icon иконка кнопки
	Icon string `json:"icon,omitempty"`
	
	// Color цвет кнопки (для поддерживающих транспортов)
	Color string `json:"color,omitempty"`
}

// ButtonType тип кнопки
type ButtonType string

const (
	ButtonTypeCallback   ButtonType = "callback"
	ButtonTypeURL        ButtonType = "url"
	ButtonTypeSwitch     ButtonType = "switch"
	ButtonTypeWebApp     ButtonType = "webapp"
	ButtonTypeContact    ButtonType = "contact"
	ButtonTypeLocation   ButtonType = "location"
	ButtonTypePoll       ButtonType = "poll"
)

// KeyboardOptions опции клавиатуры
type KeyboardOptions struct {
	// OneTime одноразовая клавиатура
	OneTime bool `json:"one_time,omitempty"`
	
	// Resize подстроить размер
	Resize bool `json:"resize,omitempty"`
	
	// Selective выборочная отправка
	Selective bool `json:"selective,omitempty"`
	
	// Placeholder текст-заполнитель
	Placeholder string `json:"placeholder,omitempty"`
}

// Embed встраиваемый элемент (для rich messages)
type Embed struct {
	// Type тип встраиваемого элемента
	Type EmbedType `json:"type"`
	
	// Title заголовок
	Title string `json:"title,omitempty"`
	
	// Description описание
	Description string `json:"description,omitempty"`
	
	// URL ссылка
	URL string `json:"url,omitempty"`
	
	// Color цвет
	Color string `json:"color,omitempty"`
	
	// Fields поля
	Fields []EmbedField `json:"fields,omitempty"`
	
	// Footer подвал
	Footer string `json:"footer,omitempty"`
	
	// Thumbnail миниатюра
	Thumbnail string `json:"thumbnail,omitempty"`
	
	// Image изображение
	Image string `json:"image,omitempty"`
}

// EmbedType тип встраиваемого элемента
type EmbedType string

const (
	EmbedTypeRich    EmbedType = "rich"
	EmbedTypeImage   EmbedType = "image"
	EmbedTypeVideo   EmbedType = "video"
	EmbedTypeArticle EmbedType = "article"
	EmbedTypeCard    EmbedType = "card"
)

// EmbedField поле встраиваемого элемента
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// BaseResponse базовая реализация Response
type BaseResponse struct {
	responseType ResponseType
	content      MessageContent
	options      ResponseOptions
	actions      []Response
}

// NewBaseResponse создает новый базовый ответ
func NewBaseResponse(t ResponseType) *BaseResponse {
	return &BaseResponse{
		responseType: t,
		content:      MessageContent{ParseMode: ParseModeHTML},
		options:      ResponseOptions{},
		actions:      make([]Response, 0),
	}
}

// Implementation
func (r *BaseResponse) Type() ResponseType         { return r.responseType }
func (r *BaseResponse) Content() MessageContent    { return r.content }
func (r *BaseResponse) Options() ResponseOptions   { return r.options }
func (r *BaseResponse) Actions() []Response        { return r.actions }
func (r *BaseResponse) IsEmpty() bool              { return r.content.Text == "" && len(r.content.Media) == 0 }
func (r *BaseResponse) IsSilent() bool             { return r.responseType == ResponseTypeSilent }

// Builder methods
func (r *BaseResponse) WithText(text string) *BaseResponse {
	r.content.Text = text
	return r
}

func (r *BaseResponse) WithParseMode(mode ParseMode) *BaseResponse {
	r.content.ParseMode = mode
	return r
}

func (r *BaseResponse) WithKeyboard(keyboard Keyboard) *BaseResponse {
	r.content.Keyboard = keyboard
	return r
}

func (r *BaseResponse) WithMedia(media ...Media) *BaseResponse {
	r.content.Media = append(r.content.Media, media...)
	return r
}

func (r *BaseResponse) WithReplyTo(messageID string) *BaseResponse {
	r.options.ReplyToMessageID = messageID
	return r
}

func (r *BaseResponse) WithDeleteUserMessage() *BaseResponse {
	r.options.DeleteUserMessage = true
	return r
}

// === Конструкторы для удобства ===

// NewMessage создает новое сообщение
func NewMessage(text string) *BaseResponse {
	return NewBaseResponse(ResponseTypeMessage).WithText(text)
}

// NewEditMessage создает ответ с редактированием
func NewEditMessage(messageID, text string) *BaseResponse {
	resp := NewBaseResponse(ResponseTypeEdit).WithText(text)
	resp.options.MessageToEditID = messageID
	return resp
}

// NewDeleteMessage создает ответ с удалением
func NewDeleteMessage(messageID string) *BaseResponse {
	resp := NewBaseResponse(ResponseTypeDelete)
	resp.options.MessageToDeleteID = messageID
	return resp
}

// NewSilentResponse создает тихий ответ
func NewSilentResponse() *BaseResponse {
	return NewBaseResponse(ResponseTypeSilent)
}

// NewMultipleResponse создает множественный ответ
func NewMultipleResponse(actions ...Response) *BaseResponse {
	resp := NewBaseResponse(ResponseTypeMultiple)
	resp.actions = actions
	return resp
}