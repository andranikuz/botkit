# BotKit - Universal Bot Framework for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/andranikuz/botkit.svg)](https://pkg.go.dev/github.com/andranikuz/botkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/andranikuz/botkit)](https://goreportcard.com/report/github.com/andranikuz/botkit)

Universal bot framework for Go. Write your bot logic once, deploy everywhere. Create transport-agnostic bot modules that work seamlessly with Telegram, HTTP APIs, WebSockets and more.

## 📦 Структура

```
botkit/
├── core/              # Ядро системы
│   ├── module.go      # Интерфейсы модулей
│   ├── context.go     # Универсальный контекст
│   ├── response.go    # Универсальные ответы
│   └── interfaces.go  # Основные интерфейсы
│
├── routing/           # Роутинг и безопасность
│   ├── pattern.go     # Паттерны маршрутов
│   ├── security.go    # Правила безопасности
│   └── router.go      # Основной роутер
│
├── events/            # Событийная система
│   ├── event.go       # События
│   └── bus.go         # Шина событий
│
├── adapters/          # Адаптеры транспортов
│   ├── telegram/      # Telegram Bot API
│   └── http/          # REST API
│
└── example/           # Пример модуля
    └── arena_module.go
```

## 🚀 Быстрый старт

### 1. Установка

```bash
go get github.com/andranikuz/botkit@latest
```

### 2. Создание модуля

```go
package mymodule

import (
    "github.com/andranikuz/botkit/core"
    "github.com/andranikuz/botkit/routing"
)

type MyModule struct {
    name    string
    version string
    logger  core.Logger
}

func NewMyModule() *MyModule {
    return &MyModule{
        name:    "mymodule",
        version: "1.0.0",
    }
}

func (m *MyModule) Name() string    { return m.name }
func (m *MyModule) Version() string { return m.version }

func (m *MyModule) Init(deps core.Dependencies) error {
    m.logger = deps.Logger()
    return nil
}

func (m *MyModule) Start(ctx context.Context) error {
    m.logger.Info("Module started")
    return nil
}

func (m *MyModule) Stop(ctx context.Context) error {
    m.logger.Info("Module stopped")
    return nil
}

func (m *MyModule) Routes() []routing.RoutePattern {
    return []routing.RoutePattern{
        routing.NewRoute("/start", "начать").
            Handler(m.handleStart).
            RequireAuth().
            Meta("start", "Начало работы").
            Build(),
    }
}

func (m *MyModule) handleStart(ctx core.UniversalContext) core.Response {
    return core.NewMessage("Привет! Это мой модуль.")
}
```

### 3. Регистрация модуля

```go
import (
    "github.com/andranikuz/botkit/events"
    "github.com/andranikuz/botkit/routing"
)

// Создаем зависимости
eventBus := events.NewEventBus(logger, metrics)
router := routing.NewRouter(eventBus, logger, config)

// Регистрируем модули
router.RegisterModule(mymodule.NewMyModule())
router.RegisterModule(arena.NewArenaModule())

// Для wildcard модулей
router.RegisterWildcard(ai.NewAIModule())

// Запускаем
router.Start(ctx)
```

### 4. Использование с Telegram

```go
import (
    "github.com/andranikuz/botkit/adapters/telegram"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Создаем Telegram адаптер
bot, _ := tgbotapi.NewBotAPI(token)
adapter := telegram.NewAdapter(bot, logger, config)
adapter.UseRouter(router)

// Обрабатываем updates
updates := bot.GetUpdatesChan(u)
for update := range updates {
    adapter.HandleUpdate(update)
}
```

### 5. Использование с HTTP API

```go
import "github.com/andranikuz/botkit/adapters/http"

// Создаем HTTP адаптер
adapter := http.NewAdapter(logger, config)
adapter.UseRouter(router)

// Запускаем сервер
adapter.ListenAndServe(":8080")
```

## 🛡️ Безопасность

### Настройка прав доступа

```go
routing.NewRoute("/admin").
    Handler(handleAdmin).
    RequireAuth().                      // Требует аутентификации
    RequireRoles("admin", "moderator"). // Требует роль
    RequirePermissions("manage_users"). // Требует право
    RateLimit(10, 60).                  // 10 запросов в минуту
    Build()
```

### Middleware

```go
// Создаем security middleware
security := routing.NewSecurityMiddleware(routing.SecurityRule{
    RequireAuth: true,
    RateLimit: &routing.RateLimitConfig{
        Requests: 100,
        Window:   60,
    },
})

router.RegisterMiddleware(security)
```

## 🎯 Wildcard модули

Для обработки неструктурированных сообщений (AI, поиск):

```go
type AIModule struct {
    // ...
}

func (m *AIModule) Priority() int { 
    return 10 // Низкий приоритет
}

func (m *AIModule) ShouldHandle(ctx core.UniversalContext) bool {
    // Проверяем, должен ли модуль обработать сообщение
    return strings.Contains(ctx.GetText(), "?")
}

func (m *AIModule) HandleWildcard(ctx core.UniversalContext) core.Response {
    // Обрабатываем сообщение через AI
    answer := m.ai.Process(ctx.GetText())
    return core.NewMessage(answer)
}
```

## 📡 События

### Публикация событий

```go
event := events.NewEvent("user.action", "mymodule")
event.SetUserID(userID).
    SetData("action", "purchase").
    SetData("amount", 100)

eventBus.PublishAsync(ctx, event)
```

### Подписка на события

```go
func (m *MyModule) Events() []core.EventSubscription {
    return []core.EventSubscription{
        {
            EventType: "user.action",
            Handler:   m.handleUserAction,
            Priority:  50,
        },
    }
}

func (m *MyModule) handleUserAction(ctx context.Context, event core.Event) error {
    action := event.Data()["action"]
    // Обрабатываем событие
    return nil
}
```

## 🔧 HTTP API

Модули могут предоставлять HTTP endpoints:

```go
func (m *MyModule) APIHandlers() []core.APIHandler {
    return []core.APIHandler{
        {
            Method:  "GET",
            Path:    "/items",
            Handler: m.apiGetItems,
        },
        {
            Method:  "POST",
            Path:    "/items",
            Handler: m.apiCreateItem,
        },
    }
}

func (m *MyModule) apiGetItems(ctx context.Context, req core.APIRequest) (core.APIResponse, error) {
    items := m.service.GetItems(req.UserID)
    
    return core.APIResponse{
        Status: 200,
        Body:   items,
    }, nil
}
```

## 📋 Клавиатуры

### Inline клавиатура

```go
import "github.com/andranikuz/botkit/adapters/telegram"

keyboard := telegram.NewInlineKeyboard().
    Row(
        core.Button{Text: "Да", Type: core.ButtonTypeCallback, Data: "yes"},
        core.Button{Text: "Нет", Type: core.ButtonTypeCallback, Data: "no"},
    ).
    Button("Отмена", "cancel")

response := core.NewMessage("Выберите действие").
    WithKeyboard(keyboard)
```

### Reply клавиатура

```go
keyboard := telegram.NewReplyKeyboard().
    Row("Кнопка 1", "Кнопка 2").
    Row("Кнопка 3").
    OneTime().
    Resize()
```

## 🎨 Ответы

### Простое сообщение

```go
core.NewMessage("Привет!")
```

### Редактирование сообщения

```go
core.NewEditMessage(messageID, "Новый текст")
```

### Удаление сообщения

```go
core.NewDeleteMessage(messageID)
```

### Множественный ответ

```go
core.NewMultipleResponse(
    core.NewDeleteMessage(oldMessageID),
    core.NewMessage("Новое сообщение"),
)
```

### С медиа

```go
response := core.NewMessage("Фото").
    WithMedia(core.Media{
        Type:   core.MediaTypePhoto,
        FileID: "AgACAgIAAxkBAAIF...",
    })
```

## 🔍 Паттерны маршрутов

```go
// Простые паттерны
"/start"
"помощь"
"статистика"

// С параметрами
"бой {id}"          // Извлекает ID как параметр
"товар {name}"      // Извлекает название
"страница {page}"   // Извлекает номер страницы

// Wildcard
"*"                 // Матчит все сообщения

// Callback паттерны
"module:action"
"arena:fight"
"shop:buy:{id}"
```

## 📊 Метрики и логирование

```go
// Логирование
logger.Info("Module action", "user", userID, "action", "purchase")

// Метрики
metrics.Counter("module.actions", 1, "type", "purchase")
metrics.Timing("module.response_time", duration, "handler", "purchase")
```

## 🏗️ Архитектурные принципы

1. **Транспортная независимость** - модули не знают о способе доставки сообщений
2. **Единая точка входа** - все сообщения проходят через роутер
3. **Безопасность по умолчанию** - проверки прав на уровне роутинга
4. **Event-driven** - слабая связанность через события
5. **Расширяемость** - легко добавлять новые транспорты и модули

## 📚 Примеры

См. папку `example/` для полного примера модуля арены с:
- Командами и callback'ами
- Событиями
- HTTP API
- Клавиатурами
- Безопасностью

## 🤝 Миграция существующих модулей

1. Замените `tgbotapi.Update` на `core.UniversalContext`
2. Замените `tgbotapi.Message` на `core.Response`
3. Используйте `routing.RoutePattern` вместо ручного роутинга
4. Добавьте `Init`, `Start`, `Stop` методы
5. Опционально добавьте HTTP API через `APIHandlers()`

## 📄 Лицензия

MIT License - см. файл LICENSE

## 🤔 Поддержка

- Создайте issue на GitHub
- Документация: https://pkg.go.dev/github.com/andranikuz/botkit
- Примеры: см. папку `example/`