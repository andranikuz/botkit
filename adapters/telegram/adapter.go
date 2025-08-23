package telegram

import (
	"context"
	"fmt"
	"github.com/andranikuz/botkit/core"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Adapter адаптер для Telegram
type Adapter struct {
	bot    *tgbotapi.BotAPI
	router core.Router
	logger core.Logger
	config core.Config
}

// NewAdapter создает новый Telegram адаптер
func NewAdapter(bot *tgbotapi.BotAPI, logger core.Logger, config core.Config) *Adapter {
	return &Adapter{
		bot:    bot,
		logger: logger,
		config: config,
	}
}

// UseRouter устанавливает роутер
func (a *Adapter) UseRouter(router core.Router) {
	a.router = router
}

// HandleUpdate обрабатывает Telegram update
func (a *Adapter) HandleUpdate(update tgbotapi.Update) {
	if a.router == nil {
		a.logger.Error("Router not set")
		return
	}

	// Конвертируем update в UniversalContext
	ctx := a.updateToContext(&update)

	// Роутим через основной роутер
	response := a.router.Route(ctx)

	// Отправляем ответ
	if err := a.sendResponse(ctx, response); err != nil {
		a.logger.Error("Failed to send response", "error", err)
	}
}

// updateToContext конвертирует Telegram Update в UniversalContext
func (a *Adapter) updateToContext(update *tgbotapi.Update) core.UniversalContext {
	ctx := core.NewBaseContext(context.Background())

	// Устанавливаем источник
	ctx.SetSource("telegram")

	// Сохраняем оригинальный update
	ctx.SetOriginal(update)

	// Обрабатываем разные типы update
	if update.Message != nil {
		a.fillFromMessage(ctx, update.Message)
		ctx.SetIsCommand(update.Message.IsCommand())
	} else if update.CallbackQuery != nil {
		a.fillFromCallbackQuery(ctx, update.CallbackQuery)
		ctx.SetIsCallback(true)
	} else if update.EditedMessage != nil {
		a.fillFromMessage(ctx, update.EditedMessage)
		ctx.Set("edited", true)
	}

	return ctx
}

// fillFromMessage заполняет контекст из сообщения
func (a *Adapter) fillFromMessage(ctx *core.BaseContext, msg *tgbotapi.Message) {
	ctx.SetUserID(msg.From.ID)
	ctx.SetChatID(msg.Chat.ID)
	ctx.SetMessageID(strconv.Itoa(msg.MessageID))
	// MessageThreadID доступен только в новых версиях API
	// ctx.SetThreadID(strconv.Itoa(msg.MessageThreadID))

	ctx.SetUsername(msg.From.UserName)
	ctx.SetFirstName(msg.From.FirstName)
	ctx.SetLastName(msg.From.LastName)

	ctx.SetText(msg.Text)

	// Обрабатываем медиа
	if msg.Photo != nil && len(msg.Photo) > 0 {
		media := make([]core.Media, 0, len(msg.Photo))
		for _, photo := range msg.Photo {
			media = append(media, core.Media{
				Type:   core.MediaTypePhoto,
				FileID: photo.FileID,
				Size:   int64(photo.FileSize),
			})
		}
		ctx.SetMedia(media)
	}

	if msg.Video != nil {
		ctx.SetMedia([]core.Media{{
			Type:     core.MediaTypeVideo,
			FileID:   msg.Video.FileID,
			Size:     int64(msg.Video.FileSize),
			MimeType: msg.Video.MimeType,
		}})
	}

	if msg.Document != nil {
		ctx.SetMedia([]core.Media{{
			Type:     core.MediaTypeDocument,
			FileID:   msg.Document.FileID,
			Size:     int64(msg.Document.FileSize),
			MimeType: msg.Document.MimeType,
		}})
	}
}

// fillFromCallbackQuery заполняет контекст из callback query
func (a *Adapter) fillFromCallbackQuery(ctx *core.BaseContext, query *tgbotapi.CallbackQuery) {
	ctx.SetUserID(query.From.ID)

	if query.Message != nil {
		ctx.SetChatID(query.Message.Chat.ID)
		ctx.SetMessageID(strconv.Itoa(query.Message.MessageID))
		// MessageThreadID доступен только в новых версиях API
		// ctx.SetThreadID(strconv.Itoa(query.Message.MessageThreadID))
	}

	ctx.SetUsername(query.From.UserName)
	ctx.SetFirstName(query.From.FirstName)
	ctx.SetLastName(query.From.LastName)

	ctx.SetText(query.Data)
	ctx.Set("callback_query_id", query.ID)

	// Парсим callback data как route
	if route, params := parseCallbackData(query.Data); route != "" {
		ctx.Set("route", route)
		for k, v := range params {
			ctx.SetParam(k, v)
		}
	}
}

// sendResponse отправляет ответ в Telegram
func (a *Adapter) sendResponse(ctx core.UniversalContext, response core.Response) error {
	if response.IsSilent() {
		return nil
	}

	switch response.Type() {
	case core.ResponseTypeMessage:
		return a.sendMessage(ctx, response)

	case core.ResponseTypeEdit:
		return a.editMessage(ctx, response)

	case core.ResponseTypeDelete:
		return a.deleteMessage(ctx, response)

	case core.ResponseTypeCallback:
		return a.answerCallback(ctx, response)

	case core.ResponseTypeMultiple:
		for _, action := range response.Actions() {
			if err := a.sendResponse(ctx, action); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("unsupported response type: %s", response.Type())
	}
}

// sendMessage отправляет новое сообщение
func (a *Adapter) sendMessage(ctx core.UniversalContext, response core.Response) error {
	content := response.Content()
	options := response.Options()

	msg := tgbotapi.NewMessage(ctx.GetChatID(), content.Text)

	// Устанавливаем parse mode
	switch content.ParseMode {
	case core.ParseModeHTML:
		msg.ParseMode = tgbotapi.ModeHTML
	case core.ParseModeMarkdown:
		msg.ParseMode = tgbotapi.ModeMarkdownV2
	}

	// Reply to message
	if options.ReplyToMessageID != "" {
		if id, err := strconv.Atoi(options.ReplyToMessageID); err == nil {
			msg.ReplyToMessageID = id
		}
	}

	// Disable notification
	msg.DisableNotification = options.DisableNotification

	// Disable web preview
	msg.DisableWebPagePreview = options.DisableWebPreview

	// Добавляем клавиатуру
	if content.Keyboard != nil {
		msg.ReplyMarkup = a.buildKeyboard(content.Keyboard)
	}

	// Отправляем
	_, err := a.bot.Send(msg)

	// Удаляем сообщение пользователя если нужно
	if err == nil && options.DeleteUserMessage {
		if msgID := ctx.GetMessageID(); msgID != "" {
			if id, err := strconv.Atoi(msgID); err == nil {
				deleteMsg := tgbotapi.NewDeleteMessage(ctx.GetChatID(), id)
				a.bot.Send(deleteMsg)
			}
		}
	}

	return err
}

// editMessage редактирует сообщение
func (a *Adapter) editMessage(ctx core.UniversalContext, response core.Response) error {
	content := response.Content()
	options := response.Options()

	msgID, err := strconv.Atoi(options.MessageToEditID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	edit := tgbotapi.NewEditMessageText(ctx.GetChatID(), msgID, content.Text)

	// Устанавливаем parse mode
	switch content.ParseMode {
	case core.ParseModeHTML:
		edit.ParseMode = tgbotapi.ModeHTML
	case core.ParseModeMarkdown:
		edit.ParseMode = tgbotapi.ModeMarkdownV2
	}

	// Добавляем клавиатуру
	if content.Keyboard != nil {
		markup := a.buildKeyboard(content.Keyboard)
		if inlineMarkup, ok := markup.(tgbotapi.InlineKeyboardMarkup); ok {
			edit.ReplyMarkup = &inlineMarkup
		}
	}

	_, err = a.bot.Send(edit)
	return err
}

// deleteMessage удаляет сообщение
func (a *Adapter) deleteMessage(ctx core.UniversalContext, response core.Response) error {
	options := response.Options()

	msgID, err := strconv.Atoi(options.MessageToDeleteID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	deleteMsg := tgbotapi.NewDeleteMessage(ctx.GetChatID(), msgID)
	_, err = a.bot.Send(deleteMsg)
	return err
}

// answerCallback отвечает на callback query
func (a *Adapter) answerCallback(ctx core.UniversalContext, response core.Response) error {
	options := response.Options()

	callbackID, ok := ctx.Get("callback_query_id")
	if !ok {
		return fmt.Errorf("callback query ID not found")
	}

	callback := tgbotapi.NewCallback(callbackID.(string), options.CallbackText)
	callback.ShowAlert = options.ShowAlert
	callback.CacheTime = options.CacheTime

	_, err := a.bot.Request(callback)
	return err
}

// buildKeyboard строит клавиатуру
func (a *Adapter) buildKeyboard(keyboard core.Keyboard) interface{} {
	switch keyboard.Type() {
	case core.KeyboardTypeInline:
		return a.buildInlineKeyboard(keyboard)

	case core.KeyboardTypeReply:
		return a.buildReplyKeyboard(keyboard)

	case core.KeyboardTypeRemove:
		return tgbotapi.NewRemoveKeyboard(true)

	default:
		return nil
	}
}

// buildInlineKeyboard строит inline клавиатуру
func (a *Adapter) buildInlineKeyboard(keyboard core.Keyboard) tgbotapi.InlineKeyboardMarkup {
	buttons := keyboard.Buttons()
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(buttons))

	for _, row := range buttons {
		tgRow := make([]tgbotapi.InlineKeyboardButton, 0, len(row))

		for _, btn := range row {
			tgBtn := tgbotapi.InlineKeyboardButton{
				Text: btn.Text,
			}

			switch btn.Type {
			case core.ButtonTypeCallback:
				data := btn.Data
				if btn.Route != nil {
					data = formatRoute(btn.Route)
				}
				tgBtn.CallbackData = &data

			case core.ButtonTypeURL:
				tgBtn.URL = &btn.Data

			case core.ButtonTypeSwitch:
				tgBtn.SwitchInlineQuery = &btn.Data
			}

			tgRow = append(tgRow, tgBtn)
		}

		rows = append(rows, tgRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// buildReplyKeyboard строит reply клавиатуру
func (a *Adapter) buildReplyKeyboard(keyboard core.Keyboard) tgbotapi.ReplyKeyboardMarkup {
	buttons := keyboard.Buttons()
	options := keyboard.Options()

	rows := make([][]tgbotapi.KeyboardButton, 0, len(buttons))

	for _, row := range buttons {
		tgRow := make([]tgbotapi.KeyboardButton, 0, len(row))

		for _, btn := range row {
			tgBtn := tgbotapi.NewKeyboardButton(btn.Text)

			switch btn.Type {
			case core.ButtonTypeContact:
				tgBtn.RequestContact = true
			case core.ButtonTypeLocation:
				tgBtn.RequestLocation = true
			}

			tgRow = append(tgRow, tgBtn)
		}

		rows = append(rows, tgRow)
	}

	markup := tgbotapi.NewReplyKeyboard(rows...)
	markup.OneTimeKeyboard = options.OneTime
	markup.ResizeKeyboard = options.Resize
	markup.Selective = options.Selective

	return markup
}

// Helper functions

// parseCallbackData парсит callback data
func parseCallbackData(data string) (route string, params map[string]string) {
	params = make(map[string]string)

	// Формат: module:action:param1=value1:param2=value2
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return data, params
	}

	route = parts[0] + ":" + parts[1]

	for i := 2; i < len(parts); i++ {
		kv := strings.SplitN(parts[i], "=", 2)
		if len(kv) == 2 {
			params[kv[0]] = kv[1]
		}
	}

	return route, params
}

// formatRoute форматирует route в callback data
func formatRoute(route *core.Route) string {
	parts := []string{route.Module, route.Action}

	for k, v := range route.Params {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	return strings.Join(parts, ":")
}
