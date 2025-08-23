package example

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/andranikuz/botkit/adapters/websocket"
	"github.com/andranikuz/botkit/core"
	"github.com/andranikuz/botkit/events"
	"github.com/andranikuz/botkit/routing"
	"github.com/gorilla/mux"
)

// WebSocketExample пример использования WebSocket адаптера
func WebSocketExample() {
	// Создаем зависимости
	logger := NewSimpleLogger()
	eventBus := events.NewEventBus(logger, nil)
	config := NewSimpleConfig()

	// Создаем роутер
	router := routing.NewRouter(eventBus, logger, config)

	// Создаем зависимости для модулей
	deps := NewSimpleDependencies(eventBus, logger, config)
	router.SetDependencies(deps)

	// Регистрируем модули
	// router.RegisterModule(NewArenaModule()) // Arena module нужно адаптировать
	router.RegisterModule(NewChatModule())

	// Запускаем роутер
	if err := router.Start(context.Background()); err != nil {
		log.Fatal("Failed to start router:", err)
	}

	// Создаем WebSocket адаптер
	wsAdapter := websocket.NewAdapter(logger, config)
	wsAdapter.UseRouter(router)

	// Создаем HTTP роутер
	httpRouter := mux.NewRouter()

	// Регистрируем WebSocket endpoint
	httpRouter.HandleFunc("/ws", wsAdapter.WebSocketHandler())

	// Статическая страница для теста
	httpRouter.HandleFunc("/", serveTestPage)

	// Запускаем сервер
	log.Println("WebSocket server starting on :8080")
	log.Println("Open http://localhost:8080 to test")
	log.Fatal(http.ListenAndServe(":8080", httpRouter))
}

// serveTestPage отдает тестовую HTML страницу
func serveTestPage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>BotKit WebSocket Test</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        #messages { 
            border: 1px solid #ccc; 
            height: 400px; 
            overflow-y: auto; 
            padding: 10px;
            margin-bottom: 10px;
            background: #f5f5f5;
        }
        .message { 
            margin: 5px 0; 
            padding: 5px 10px;
            border-radius: 5px;
        }
        .sent { 
            background: #e3f2fd; 
            text-align: right;
        }
        .received { 
            background: #f5f5f5; 
            text-align: left;
        }
        .error { 
            background: #ffebee; 
            color: #c62828;
        }
        .system { 
            background: #fff3e0; 
            color: #e65100;
            font-style: italic;
        }
        #input { 
            width: 70%; 
            padding: 10px;
            font-size: 16px;
        }
        button { 
            padding: 10px 20px;
            font-size: 16px;
            margin-left: 10px;
        }
        #status {
            display: inline-block;
            padding: 5px 10px;
            border-radius: 5px;
            margin-bottom: 10px;
        }
        .connected { background: #c8e6c9; color: #2e7d32; }
        .disconnected { background: #ffcdd2; color: #c62828; }
    </style>
</head>
<body>
    <h1>BotKit WebSocket Test</h1>
    
    <div id="status" class="disconnected">Disconnected</div>
    
    <div id="messages"></div>
    
    <div>
        <input type="text" id="input" placeholder="Type a message..." />
        <button onclick="sendMessage()">Send</button>
        <button onclick="sendCommand()">Send /start</button>
        <button onclick="connect()">Connect</button>
        <button onclick="disconnect()">Disconnect</button>
    </div>

    <script>
        let ws = null;
        let messageId = 1;
        const userId = Math.floor(Math.random() * 1000000);

        function connect() {
            if (ws) {
                ws.close();
            }

            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/ws?user_id=' + userId;
            
            ws = new WebSocket(wsUrl);

            ws.onopen = function() {
                updateStatus(true);
                addMessage('Connected to server', 'system');
            };

            ws.onmessage = function(event) {
                const msg = JSON.parse(event.data);
                handleMessage(msg);
            };

            ws.onerror = function(error) {
                addMessage('WebSocket error: ' + error, 'error');
            };

            ws.onclose = function() {
                updateStatus(false);
                addMessage('Disconnected from server', 'system');
            };
        }

        function disconnect() {
            if (ws) {
                ws.close();
                ws = null;
            }
        }

        function sendMessage() {
            const input = document.getElementById('input');
            const text = input.value.trim();
            
            if (!text || !ws || ws.readyState !== WebSocket.OPEN) {
                return;
            }

            const msg = {
                type: 'message',
                id: 'msg_' + messageId++,
                user_id: userId,
                chat_id: userId,
                text: text,
                data: {}
            };

            ws.send(JSON.stringify(msg));
            addMessage('You: ' + text, 'sent');
            input.value = '';
        }

        function sendCommand() {
            const msg = {
                type: 'command',
                id: 'cmd_' + messageId++,
                user_id: userId,
                chat_id: userId,
                text: '/start',
                data: {}
            };

            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify(msg));
                addMessage('You: /start', 'sent');
            }
        }

        function handleMessage(msg) {
            if (msg.type === 'connected') {
                addMessage('Server version: ' + msg.data.version, 'system');
            } else if (msg.type === 'response') {
                if (msg.data && msg.data.text) {
                    addMessage('Bot: ' + msg.data.text, 'received');
                }
            } else if (msg.type === 'error') {
                addMessage('Error: ' + msg.error, 'error');
            } else {
                addMessage('Unknown message: ' + JSON.stringify(msg), 'system');
            }
        }

        function addMessage(text, className) {
            const messages = document.getElementById('messages');
            const div = document.createElement('div');
            div.className = 'message ' + className;
            div.textContent = text;
            messages.appendChild(div);
            messages.scrollTop = messages.scrollHeight;
        }

        function updateStatus(connected) {
            const status = document.getElementById('status');
            if (connected) {
                status.className = 'connected';
                status.textContent = 'Connected (User ID: ' + userId + ')';
            } else {
                status.className = 'disconnected';
                status.textContent = 'Disconnected';
            }
        }

        // Enter key to send
        document.getElementById('input').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });

        // Auto-connect on load
        window.onload = function() {
            connect();
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// ChatModule простой модуль для демонстрации
type ChatModule struct {
	name    string
	version string
}

func NewChatModule() *ChatModule {
	return &ChatModule{
		name:    "chat",
		version: "1.0.0",
	}
}

func (m *ChatModule) Name() string    { return m.name }
func (m *ChatModule) Version() string { return m.version }

func (m *ChatModule) Init(deps core.Dependencies) error {
	return nil
}

func (m *ChatModule) Start(ctx context.Context) error {
	return nil
}

func (m *ChatModule) Stop(ctx context.Context) error {
	return nil
}

func (m *ChatModule) Routes() []core.RoutePattern {
	return []core.RoutePattern{
		routing.RoutePattern{
			Patterns: []string{"/start", "начать", "start"},
			Handler:  m.handleStart,
			Priority: 100,
			Type:     routing.RouteTypeCommand,
		},
		routing.RoutePattern{
			Patterns: []string{"привет", "hello", "hi"},
			Handler:  m.handleHello,
			Priority: 50,
			Type:     routing.RouteTypeMessage,
		},
	}
}

func (m *ChatModule) handleStart(ctx core.UniversalContext) core.Response {
	return core.NewMessage("👋 Добро пожаловать в BotKit!\n\nЯ универсальный бот, работающий через WebSocket.\n\nПопробуйте написать 'привет' или любое сообщение.")
}

func (m *ChatModule) handleHello(ctx core.UniversalContext) core.Response {
	username := ctx.GetUsername()
	if username == "" {
		username = "друг"
	}
	return core.NewMessage(fmt.Sprintf("Привет, %s! 👋", username))
}

// Helper implementations

type SimpleLogger struct{}

func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{}
}

func (l *SimpleLogger) Debug(msg string, fields ...interface{}) {
	log.Printf("[DEBUG] %s %v", msg, fields)
}
func (l *SimpleLogger) Info(msg string, fields ...interface{}) {
	log.Printf("[INFO] %s %v", msg, fields)
}
func (l *SimpleLogger) Warn(msg string, fields ...interface{}) {
	log.Printf("[WARN] %s %v", msg, fields)
}
func (l *SimpleLogger) Error(msg string, fields ...interface{}) {
	log.Printf("[ERROR] %s %v", msg, fields)
}
func (l *SimpleLogger) Fatal(msg string, fields ...interface{}) {
	log.Fatalf("[FATAL] %s %v", msg, fields)
}

func (l *SimpleLogger) WithField(key string, value interface{}) core.Logger  { return l }
func (l *SimpleLogger) WithFields(fields map[string]interface{}) core.Logger { return l }
func (l *SimpleLogger) WithError(err error) core.Logger                      { return l }

type SimpleConfig struct{}

func NewSimpleConfig() *SimpleConfig {
	return &SimpleConfig{}
}

func (c *SimpleConfig) Get(key string) interface{}                { return nil }
func (c *SimpleConfig) GetString(key string) string               { return "" }
func (c *SimpleConfig) GetInt(key string) int                     { return 0 }
func (c *SimpleConfig) GetBool(key string) bool                   { return false }
func (c *SimpleConfig) GetStringSlice(key string) []string        { return nil }
func (c *SimpleConfig) GetStringMap(key string) map[string]string { return nil }
func (c *SimpleConfig) Set(key string, value interface{})         {}
func (c *SimpleConfig) IsSet(key string) bool                     { return false }

type SimpleDependencies struct {
	eventBus core.EventBus
	logger   core.Logger
	config   core.Config
}

func NewSimpleDependencies(eventBus core.EventBus, logger core.Logger, config core.Config) *SimpleDependencies {
	return &SimpleDependencies{
		eventBus: eventBus,
		logger:   logger,
		config:   config,
	}
}

func (d *SimpleDependencies) Database() interface{}              { return nil }
func (d *SimpleDependencies) EventBus() core.EventBus            { return d.eventBus }
func (d *SimpleDependencies) Logger() core.Logger                { return d.logger }
func (d *SimpleDependencies) Config() core.Config                { return d.config }
func (d *SimpleDependencies) Get(key string) (interface{}, bool) { return nil, false }
func (d *SimpleDependencies) Set(key string, value interface{})  {}
