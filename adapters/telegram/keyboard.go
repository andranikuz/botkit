package telegram

import (
	"fmt"
	"github.com/andranikuz/botkit/core"
)

// InlineKeyboard реализация inline клавиатуры
type InlineKeyboard struct {
	buttons [][]core.Button
	options core.KeyboardOptions
}

// NewInlineKeyboard создает новую inline клавиатуру
func NewInlineKeyboard() *InlineKeyboard {
	return &InlineKeyboard{
		buttons: make([][]core.Button, 0),
	}
}

// Type возвращает тип клавиатуры
func (k *InlineKeyboard) Type() core.KeyboardType {
	return core.KeyboardTypeInline
}

// Buttons возвращает кнопки
func (k *InlineKeyboard) Buttons() [][]core.Button {
	return k.buttons
}

// Options возвращает опции
func (k *InlineKeyboard) Options() core.KeyboardOptions {
	return k.options
}

// Row добавляет новую строку кнопок
func (k *InlineKeyboard) Row(buttons ...core.Button) *InlineKeyboard {
	k.buttons = append(k.buttons, buttons)
	return k
}

// Button добавляет кнопку в последнюю строку
func (k *InlineKeyboard) Button(text, data string) *InlineKeyboard {
	btn := core.Button{
		Text: text,
		Type: core.ButtonTypeCallback,
		Data: data,
	}

	if len(k.buttons) == 0 {
		k.buttons = append(k.buttons, []core.Button{btn})
	} else {
		lastRow := len(k.buttons) - 1
		k.buttons[lastRow] = append(k.buttons[lastRow], btn)
	}

	return k
}

// CallbackButton добавляет callback кнопку
func (k *InlineKeyboard) CallbackButton(text, data string) *InlineKeyboard {
	return k.Button(text, data)
}

// URLButton добавляет URL кнопку
func (k *InlineKeyboard) URLButton(text, url string) *InlineKeyboard {
	btn := core.Button{
		Text: text,
		Type: core.ButtonTypeURL,
		Data: url,
	}

	if len(k.buttons) == 0 {
		k.buttons = append(k.buttons, []core.Button{btn})
	} else {
		lastRow := len(k.buttons) - 1
		k.buttons[lastRow] = append(k.buttons[lastRow], btn)
	}

	return k
}

// RouteButton добавляет кнопку с типизированным роутом
func (k *InlineKeyboard) RouteButton(text string, route *core.Route) *InlineKeyboard {
	btn := core.Button{
		Text:  text,
		Type:  core.ButtonTypeCallback,
		Route: route,
	}

	if len(k.buttons) == 0 {
		k.buttons = append(k.buttons, []core.Button{btn})
	} else {
		lastRow := len(k.buttons) - 1
		k.buttons[lastRow] = append(k.buttons[lastRow], btn)
	}

	return k
}

// ReplyKeyboard реализация reply клавиатуры
type ReplyKeyboard struct {
	buttons [][]core.Button
	options core.KeyboardOptions
}

// NewReplyKeyboard создает новую reply клавиатуру
func NewReplyKeyboard() *ReplyKeyboard {
	return &ReplyKeyboard{
		buttons: make([][]core.Button, 0),
		options: core.KeyboardOptions{
			Resize: true,
		},
	}
}

// Type возвращает тип клавиатуры
func (k *ReplyKeyboard) Type() core.KeyboardType {
	return core.KeyboardTypeReply
}

// Buttons возвращает кнопки
func (k *ReplyKeyboard) Buttons() [][]core.Button {
	return k.buttons
}

// Options возвращает опции
func (k *ReplyKeyboard) Options() core.KeyboardOptions {
	return k.options
}

// Row добавляет новую строку кнопок
func (k *ReplyKeyboard) Row(texts ...string) *ReplyKeyboard {
	buttons := make([]core.Button, 0, len(texts))
	for _, text := range texts {
		buttons = append(buttons, core.Button{
			Text: text,
		})
	}
	k.buttons = append(k.buttons, buttons)
	return k
}

// Button добавляет кнопку в последнюю строку
func (k *ReplyKeyboard) Button(text string) *ReplyKeyboard {
	btn := core.Button{
		Text: text,
	}

	if len(k.buttons) == 0 {
		k.buttons = append(k.buttons, []core.Button{btn})
	} else {
		lastRow := len(k.buttons) - 1
		k.buttons[lastRow] = append(k.buttons[lastRow], btn)
	}

	return k
}

// ContactButton добавляет кнопку запроса контакта
func (k *ReplyKeyboard) ContactButton(text string) *ReplyKeyboard {
	btn := core.Button{
		Text: text,
		Type: core.ButtonTypeContact,
	}

	if len(k.buttons) == 0 {
		k.buttons = append(k.buttons, []core.Button{btn})
	} else {
		lastRow := len(k.buttons) - 1
		k.buttons[lastRow] = append(k.buttons[lastRow], btn)
	}

	return k
}

// LocationButton добавляет кнопку запроса локации
func (k *ReplyKeyboard) LocationButton(text string) *ReplyKeyboard {
	btn := core.Button{
		Text: text,
		Type: core.ButtonTypeLocation,
	}

	if len(k.buttons) == 0 {
		k.buttons = append(k.buttons, []core.Button{btn})
	} else {
		lastRow := len(k.buttons) - 1
		k.buttons[lastRow] = append(k.buttons[lastRow], btn)
	}

	return k
}

// OneTime устанавливает одноразовую клавиатуру
func (k *ReplyKeyboard) OneTime() *ReplyKeyboard {
	k.options.OneTime = true
	return k
}

// Resize устанавливает подстройку размера
func (k *ReplyKeyboard) Resize() *ReplyKeyboard {
	k.options.Resize = true
	return k
}

// Selective устанавливает выборочную отправку
func (k *ReplyKeyboard) Selective() *ReplyKeyboard {
	k.options.Selective = true
	return k
}

// Placeholder устанавливает текст-заполнитель
func (k *ReplyKeyboard) Placeholder(text string) *ReplyKeyboard {
	k.options.Placeholder = text
	return k
}

// RemoveKeyboard клавиатура для удаления
type RemoveKeyboard struct{}

// Type возвращает тип клавиатуры
func (k *RemoveKeyboard) Type() core.KeyboardType {
	return core.KeyboardTypeRemove
}

// Buttons возвращает пустой массив
func (k *RemoveKeyboard) Buttons() [][]core.Button {
	return nil
}

// Options возвращает пустые опции
func (k *RemoveKeyboard) Options() core.KeyboardOptions {
	return core.KeyboardOptions{}
}

// NewRemoveKeyboard создает клавиатуру для удаления
func NewRemoveKeyboard() *RemoveKeyboard {
	return &RemoveKeyboard{}
}

// Helper функции для быстрого создания клавиатур

// QuickInlineKeyboard создает inline клавиатуру с одной строкой кнопок
func QuickInlineKeyboard(buttons ...core.Button) *InlineKeyboard {
	kb := NewInlineKeyboard()
	kb.buttons = [][]core.Button{buttons}
	return kb
}

// QuickReplyKeyboard создает reply клавиатуру с одной строкой кнопок
func QuickReplyKeyboard(texts ...string) *ReplyKeyboard {
	kb := NewReplyKeyboard()
	return kb.Row(texts...)
}

// YesNoKeyboard создает клавиатуру Да/Нет
func YesNoKeyboard(yesData, noData string) *InlineKeyboard {
	return NewInlineKeyboard().
		Row(
			core.Button{Text: "✅ Да", Type: core.ButtonTypeCallback, Data: yesData},
			core.Button{Text: "❌ Нет", Type: core.ButtonTypeCallback, Data: noData},
		)
}

// BackKeyboard создает клавиатуру с кнопкой "Назад"
func BackKeyboard(backData string) *InlineKeyboard {
	return NewInlineKeyboard().
		Button("⬅️ Назад", backData)
}

// PaginationKeyboard создает клавиатуру пагинации
func PaginationKeyboard(currentPage, totalPages int, baseRoute string) *InlineKeyboard {
	kb := NewInlineKeyboard()

	buttons := make([]core.Button, 0, 3)

	// Кнопка "Назад"
	if currentPage > 1 {
		buttons = append(buttons, core.Button{
			Text: "⬅️",
			Type: core.ButtonTypeCallback,
			Route: &core.Route{
				Module: baseRoute,
				Action: "page",
				Params: map[string]interface{}{"page": currentPage - 1},
			},
		})
	}

	// Текущая страница
	buttons = append(buttons, core.Button{
		Text: fmt.Sprintf("%d/%d", currentPage, totalPages),
		Type: core.ButtonTypeCallback,
		Data: "noop",
	})

	// Кнопка "Вперед"
	if currentPage < totalPages {
		buttons = append(buttons, core.Button{
			Text: "➡️",
			Type: core.ButtonTypeCallback,
			Route: &core.Route{
				Module: baseRoute,
				Action: "page",
				Params: map[string]interface{}{"page": currentPage + 1},
			},
		})
	}

	return kb.Row(buttons...)
}
