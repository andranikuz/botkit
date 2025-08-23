# BotKit - Руководство по использованию

## 🚀 Установка как отдельного пакета

### 1. Создание репозитория

```bash
# Переместите botkit в отдельную папку
cp -r botkit ~/projects/botkit
cd ~/projects/botkit

# Инициализируйте git
git init
git add .
git commit -m "Initial release of BotKit"

# Создайте репозиторий на GitHub и запушьте
git remote add origin https://github.com/andranikuz/botkit.git
git push -u origin main

# Создайте версию
git tag v1.0.0
git push origin v1.0.0
```

### 2. Использование в проекте

```bash
# В вашем проекте
go get github.com/andranikuz/botkit@latest
```

## 📦 Основные компоненты

### Core
- **Module** - базовый интерфейс модуля
- **UniversalContext** - универсальный контекст сообщения
- **Response** - универсальный ответ
- **Router** - маршрутизатор сообщений

### Adapters
- **Telegram** - адаптер для Telegram Bot API
- **HTTP** - адаптер для REST API
- **WebSocket** - адаптер для WebSocket соединений

### Events
- **EventBus** - асинхронная шина событий
- **Event** - базовые события системы

### Routing
- **RoutePattern** - паттерны маршрутов с поддержкой параметров
- **SecurityRule** - правила безопасности и rate limiting
- **Middleware** - промежуточные обработчики

## 💡 Примеры использования

### 1. Простой Telegram бот

```go
package main

import (
    "context"
    "log"
    
    "github.com/andranikuz/botkit/adapters/telegram"
    "github.com/andranikuz/botkit/core"
    "github.com/andranikuz/botkit/events"
    "github.com/andranikuz/botkit/routing"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
    // Создаем компоненты
    logger := NewLogger()
    eventBus := events.NewEventBus(logger, nil)
    config := NewConfig()
    
    // Создаем роутер
    router := routing.NewRouter(eventBus, logger, config)
    
    // Регистрируем модули
    router.RegisterModule(NewMyModule())
    
    // Запускаем роутер
    router.Start(context.Background())
    
    // Создаем Telegram бота
    bot, _ := tgbotapi.NewBotAPI("YOUR_TOKEN")
    adapter := telegram.NewAdapter(bot, logger, config)
    adapter.UseRouter(router)
    
    // Обрабатываем сообщения
    u := tgbotapi.NewUpdate(0)
    updates := bot.GetUpdatesChan(u)
    
    for update := range updates {
        adapter.HandleUpdate(update)
    }
}
```

### 2. HTTP API сервер

```go
package main

import (
    "context"
    "log"
    
    "github.com/andranikuz/botkit/adapters/http"
    "github.com/andranikuz/botkit/events"
    "github.com/andranikuz/botkit/routing"
)

func main() {
    // Создаем компоненты
    logger := NewLogger()
    eventBus := events.NewEventBus(logger, nil)
    config := NewConfig()
    
    // Создаем роутер
    router := routing.NewRouter(eventBus, logger, config)
    router.RegisterModule(NewAPIModule())
    router.Start(context.Background())
    
    // Создаем HTTP адаптер
    adapter := http.NewAdapter(logger, config)
    adapter.UseRouter(router)
    
    // Запускаем сервер
    log.Fatal(adapter.ListenAndServe(":8080"))
}
```

### 3. WebSocket сервер

```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/andranikuz/botkit/adapters/websocket"
    "github.com/andranikuz/botkit/events"
    "github.com/andranikuz/botkit/routing"
    "github.com/gorilla/mux"
)

func main() {
    // Создаем компоненты
    logger := NewLogger()
    eventBus := events.NewEventBus(logger, nil)
    config := NewConfig()
    
    // Создаем роутер
    router := routing.NewRouter(eventBus, logger, config)
    router.RegisterModule(NewChatModule())
    router.Start(context.Background())
    
    // Создаем WebSocket адаптер
    wsAdapter := websocket.NewAdapter(logger, config)
    wsAdapter.UseRouter(router)
    
    // HTTP роутер
    httpRouter := mux.NewRouter()
    httpRouter.HandleFunc("/ws", wsAdapter.WebSocketHandler())
    
    // Запускаем сервер
    log.Fatal(http.ListenAndServe(":8080", httpRouter))
}
```

## 🔧 Создание модуля

```go
package mymodule

import (
    "context"
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

// Обязательные методы интерфейса Module

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

func (m *MyModule) Routes() []core.RoutePattern {
    return []core.RoutePattern{
        routing.RoutePattern{
            Patterns: []string{"/start", "начать"},
            Handler:  m.handleStart,
            Priority: 100,
            Type:     routing.RouteTypeCommand,
            Security: routing.SecurityRule{
                RequireAuth: true,
            },
        },
    }
}

func (m *MyModule) handleStart(ctx core.UniversalContext) core.Response {
    return core.NewMessage("Привет! Я работаю через BotKit.")
}
```

## 🛡️ Безопасность

### Rate Limiting

```go
routing.NewRoute("/api").
    Handler(handleAPI).
    RateLimit(10, 60). // 10 запросов в минуту
    Build()
```

### Проверка прав

```go
routing.SecurityRule{
    RequireAuth:        true,
    RequireRoles:       []string{"admin"},
    RequirePermissions: []string{"manage_users"},
}
```

## 📡 События

### Публикация

```go
event := events.NewEvent("user.action", "mymodule")
event.SetUserID(userID).
    SetData("action", "purchase")

eventBus.PublishAsync(ctx, event)
```

### Подписка

```go
func (m *MyModule) Events() []core.EventSubscription {
    return []core.EventSubscription{
        {
            EventType: "user.action",
            Handler:   m.handleUserAction,
        },
    }
}
```

## 🎯 Wildcard модули

```go
type AIModule struct {
    ai AIService
}

func (m *AIModule) Priority() int { return 10 }

func (m *AIModule) ShouldHandle(ctx core.UniversalContext) bool {
    return strings.Contains(ctx.GetText(), "?")
}

func (m *AIModule) HandleWildcard(ctx core.UniversalContext) core.Response {
    answer := m.ai.Process(ctx.GetText())
    return core.NewMessage(answer)
}
```

## 📝 Тестирование WebSocket

Запустите пример:

```go
go run example/websocket_example.go
```

Откройте http://localhost:8080 в браузере для теста WebSocket соединения.

## 🚢 Деплой

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bot main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bot .
CMD ["./bot"]
```

### Systemd

```ini
[Unit]
Description=BotKit Bot
After=network.target

[Service]
Type=simple
User=bot
WorkingDirectory=/opt/bot
ExecStart=/opt/bot/bot
Restart=always

[Install]
WantedBy=multi-user.target
```

## 📄 Лицензия

MIT License

## 🤝 Поддержка

- GitHub Issues: https://github.com/andranikuz/botkit/issues
- Documentation: https://pkg.go.dev/github.com/andranikuz/botkit