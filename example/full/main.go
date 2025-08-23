package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpAdapter "github.com/andranikuz/botkit/adapters/http"
	"github.com/andranikuz/botkit/adapters/telegram"
	"github.com/andranikuz/botkit/adapters/websocket"
	"github.com/andranikuz/botkit/core"
	"github.com/andranikuz/botkit/events"
	"github.com/andranikuz/botkit/routing"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/mux"
)

// Flags
var (
	mode     = flag.String("mode", "all", "Mode: telegram, http, websocket, or all")
	token    = flag.String("token", "", "Telegram bot token")
	httpPort = flag.String("http-port", ":8080", "HTTP server port")
	wsPort   = flag.String("ws-port", ":8081", "WebSocket server port")
)

func main() {
	flag.Parse()

	// Create dependencies
	logger := NewSimpleLogger()
	eventBus := events.NewEventBus(logger, nil)
	config := NewSimpleConfig()
	deps := NewSimpleDependencies(eventBus, logger, config)

	// Create router
	router := routing.NewRouter(eventBus, logger, config)
	router.SetDependencies(deps)

	// Register modules
	router.RegisterModule(NewUniversalModule())
	router.RegisterModule(NewEventModule())

	// Register wildcard module
	router.RegisterWildcard(NewSimpleAIModule())

	// Start router
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := router.Start(ctx); err != nil {
		log.Fatal("Failed to start router:", err)
	}

	// Signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start adapters based on mode
	switch *mode {
	case "telegram":
		runTelegram(router, logger, config)
	case "http":
		runHTTP(router, logger, config)
	case "websocket":
		runWebSocket(router, logger, config)
	case "all":
		go runTelegram(router, logger, config)
		go runHTTP(router, logger, config)
		go runWebSocket(router, logger, config)
	default:
		log.Fatal("Invalid mode. Use: telegram, http, websocket, or all")
	}

	// Wait for signal
	<-sigChan
	log.Println("Shutting down...")

	// Stop router
	router.Stop(ctx)
}

func runTelegram(router core.Router, logger core.Logger, config core.Config) {
	if *token == "" {
		logger.Warn("Telegram token not provided, skipping Telegram adapter")
		return
	}

	bot, err := tgbotapi.NewBotAPI(*token)
	if err != nil {
		logger.Error("Failed to create Telegram bot", "error", err)
		return
	}

	logger.Info("Telegram bot started", "username", bot.Self.UserName)

	adapter := telegram.NewAdapter(bot, logger, config)
	adapter.UseRouter(router)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		adapter.HandleUpdate(update)
	}
}

func runHTTP(router core.Router, logger core.Logger, config core.Config) {
	adapter := httpAdapter.NewAdapter(logger, config)
	adapter.UseRouter(router)

	logger.Info("HTTP server starting", "port", *httpPort)
	if err := adapter.ListenAndServe(*httpPort); err != nil {
		logger.Error("HTTP server failed", "error", err)
	}
}

func runWebSocket(router core.Router, logger core.Logger, config core.Config) {
	wsAdapter := websocket.NewAdapter(logger, config)
	wsAdapter.UseRouter(router)

	httpRouter := mux.NewRouter()
	httpRouter.HandleFunc("/ws", wsAdapter.WebSocketHandler())
	httpRouter.HandleFunc("/", serveWebSocketTestPage)

	logger.Info("WebSocket server starting", "port", *wsPort)
	if err := http.ListenAndServe(*wsPort, httpRouter); err != nil {
		logger.Error("WebSocket server failed", "error", err)
	}
}

// UniversalModule - модуль, работающий со всеми транспортами
type UniversalModule struct {
	name    string
	version string
	logger  core.Logger
}

func NewUniversalModule() *UniversalModule {
	return &UniversalModule{
		name:    "universal",
		version: "1.0.0",
	}
}

func (m *UniversalModule) Name() string    { return m.name }
func (m *UniversalModule) Version() string { return m.version }

func (m *UniversalModule) Init(deps core.Dependencies) error {
	m.logger = deps.Logger()
	return nil
}

func (m *UniversalModule) Start(ctx context.Context) error {
	m.logger.Info("Universal module started")
	return nil
}

func (m *UniversalModule) Stop(ctx context.Context) error {
	m.logger.Info("Universal module stopped")
	return nil
}

func (m *UniversalModule) Routes() []core.RoutePattern {
	return []core.RoutePattern{
		routing.RoutePattern{
			Patterns: []string{"/start", "start", "начать"},
			Handler:  m.handleStart,
			Priority: 100,
			Type:     routing.RouteTypeCommand,
			Meta: routing.RouteMeta{
				Name:        "start",
				Description: "Start the bot",
				Examples:    []string{"/start", "начать"},
			},
		},
		routing.RoutePattern{
			Patterns: []string{"/help", "help", "помощь"},
			Handler:  m.handleHelp,
			Priority: 90,
			Type:     routing.RouteTypeCommand,
		},
		routing.RoutePattern{
			Patterns: []string{"привет", "hello", "hi"},
			Handler:  m.handleGreeting,
			Priority: 50,
			Type:     routing.RouteTypeMessage,
		},
		routing.RoutePattern{
			Patterns: []string{"test_buttons"},
			Handler:  m.handleTestButtons,
			Priority: 50,
			Type:     routing.RouteTypeCallback,
		},
		routing.RoutePattern{
			Patterns: []string{"button_yes", "button_no"},
			Handler:  m.handleButtonCallback,
			Priority: 50,
			Type:     routing.RouteTypeCallback,
		},
	}
}

func (m *UniversalModule) handleStart(ctx core.UniversalContext) core.Response {
	username := ctx.GetUsername()
	if username == "" {
		username = "User"
	}

	text := fmt.Sprintf(
		"👋 Welcome %s!\n\n"+
			"I'm a universal bot working through BotKit.\n"+
			"I can work via:\n"+
			"• Telegram Bot API\n"+
			"• HTTP REST API\n"+
			"• WebSocket connections\n\n"+
			"Try /help for available commands.",
		username,
	)

	return core.NewMessage(text)
}

func (m *UniversalModule) handleHelp(ctx core.UniversalContext) core.Response {
	text := "📚 Available commands:\n\n" +
		"/start - Start the bot\n" +
		"/help - Show this help\n" +
		"hello/привет - Get a greeting\n" +
		"test_buttons - Test inline buttons\n\n" +
		"Any question? - AI will try to help"

	return core.NewMessage(text)
}

func (m *UniversalModule) handleGreeting(ctx core.UniversalContext) core.Response {
	username := ctx.GetUsername()
	if username == "" {
		username = "friend"
	}
	return core.NewMessage(fmt.Sprintf("Hello, %s! 👋", username))
}

func (m *UniversalModule) handleTestButtons(ctx core.UniversalContext) core.Response {
	// For now, just return a message without keyboard
	// Since core.Keyboard interface is not fully implemented yet
	return core.NewMessage("Test buttons feature is not yet implemented in this example.")
}

func (m *UniversalModule) handleButtonCallback(ctx core.UniversalContext) core.Response {
	var data string
	if ctx.IsCallback() {
		if cbData, ok := ctx.GetData()["callback_data"].(string); ok {
			data = cbData
		}
	}

	var response string
	switch data {
	case "button_yes":
		response = "You clicked Yes! ✅"
	case "button_no":
		response = "You clicked No! ❌"
	default:
		response = "Unknown button"
	}

	// If it's a callback, edit the message
	if ctx.IsCallback() {
		if msgID, ok := ctx.GetData()["message_id"].(int); ok {
			return core.NewEditMessage(fmt.Sprintf("%d", msgID), response)
		}
	}

	return core.NewMessage(response)
}

// EventModule - модуль для демонстрации событий
type EventModule struct {
	name     string
	version  string
	logger   core.Logger
	eventBus core.EventBus
}

func NewEventModule() *EventModule {
	return &EventModule{
		name:    "events",
		version: "1.0.0",
	}
}

func (m *EventModule) Name() string    { return m.name }
func (m *EventModule) Version() string { return m.version }

func (m *EventModule) Init(deps core.Dependencies) error {
	m.logger = deps.Logger()
	m.eventBus = deps.EventBus()
	return nil
}

func (m *EventModule) Start(ctx context.Context) error {
	// Subscribe to events
	m.eventBus.Subscribe("user.message", m.handleUserMessage)
	return nil
}

func (m *EventModule) Stop(ctx context.Context) error {
	return nil
}

func (m *EventModule) Routes() []core.RoutePattern {
	return []core.RoutePattern{
		routing.RoutePattern{
			Patterns: []string{"/event", "event"},
			Handler:  m.handleEventCommand,
			Priority: 50,
			Type:     routing.RouteTypeCommand,
		},
	}
}

func (m *EventModule) handleEventCommand(ctx core.UniversalContext) core.Response {
	// Publish event
	event := events.NewEvent("user.action", m.name)
	event.SetUserID(ctx.GetUserID())
	event.SetData("action", "test_event")
	event.SetData("chat_id", ctx.GetChatID())

	m.eventBus.PublishAsync(context.Background(), event)

	return core.NewMessage("Event published! Check the logs.")
}

func (m *EventModule) handleUserMessage(ctx context.Context, event core.Event) error {
	m.logger.Info("Received event",
		"type", event.Type(),
		"user", event.UserID(),
		"data", event.Data(),
	)
	return nil
}

// SimpleAIModule - простой AI модуль для wildcard обработки
type SimpleAIModule struct {
	name    string
	version string
	logger  core.Logger
}

func NewSimpleAIModule() *SimpleAIModule {
	return &SimpleAIModule{
		name:    "ai",
		version: "1.0.0",
	}
}

func (m *SimpleAIModule) Name() string    { return m.name }
func (m *SimpleAIModule) Version() string { return m.version }
func (m *SimpleAIModule) Priority() int   { return 10 }

func (m *SimpleAIModule) Init(deps core.Dependencies) error {
	m.logger = deps.Logger()
	return nil
}

func (m *SimpleAIModule) Start(ctx context.Context) error {
	m.logger.Info("AI module started")
	return nil
}

func (m *SimpleAIModule) Stop(ctx context.Context) error {
	return nil
}

func (m *SimpleAIModule) Routes() []core.RoutePattern {
	return nil // Wildcard module doesn't have specific routes
}

func (m *SimpleAIModule) ShouldHandle(ctx core.UniversalContext) bool {
	text := ctx.GetText()
	// Handle questions
	return len(text) > 0 && text[len(text)-1] == '?'
}

func (m *SimpleAIModule) HandleWildcard(ctx core.UniversalContext) core.Response {
	m.logger.Info("AI handling message", "text", ctx.GetText())

	// Simple mock AI response
	responses := []string{
		"That's an interesting question! 🤔",
		"Let me think about it... 💭",
		"I'm not sure, but it sounds important! 🎯",
		"Have you tried turning it off and on again? 🔄",
		"The answer is 42! 🌟",
	}

	// Simple hash to select response
	hash := 0
	for _, r := range ctx.GetText() {
		hash += int(r)
	}

	response := responses[hash%len(responses)]
	return core.NewMessage(response)
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

func (l *SimpleLogger) WithField(key string, value interface{}) core.Logger {
	return l
}
func (l *SimpleLogger) WithFields(fields map[string]interface{}) core.Logger {
	return l
}
func (l *SimpleLogger) WithError(err error) core.Logger {
	return l
}

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

// serveWebSocketTestPage serves a test HTML page
func serveWebSocketTestPage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>BotKit WebSocket Test</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f0f0f0; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 20px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #4CAF50; padding-bottom: 10px; }
        #messages { 
            border: 1px solid #ddd; 
            height: 400px; 
            overflow-y: auto; 
            padding: 15px;
            margin-bottom: 20px;
            background: #fafafa;
            border-radius: 5px;
        }
        .message { 
            margin: 10px 0; 
            padding: 10px 15px;
            border-radius: 10px;
            animation: slideIn 0.3s ease;
        }
        @keyframes slideIn {
            from { opacity: 0; transform: translateX(-20px); }
            to { opacity: 1; transform: translateX(0); }
        }
        .sent { 
            background: #E3F2FD; 
            margin-left: 20%;
            text-align: right;
            border: 1px solid #90CAF9;
        }
        .received { 
            background: #E8F5E9; 
            margin-right: 20%;
            border: 1px solid #A5D6A7;
        }
        .error { 
            background: #FFEBEE; 
            color: #C62828;
            border: 1px solid #EF9A9A;
        }
        .system { 
            background: #FFF3E0; 
            color: #E65100;
            font-style: italic;
            text-align: center;
            border: 1px solid #FFCC80;
        }
        .input-group {
            display: flex;
            gap: 10px;
        }
        #input { 
            flex: 1;
            padding: 12px;
            font-size: 16px;
            border: 2px solid #ddd;
            border-radius: 5px;
            transition: border-color 0.3s;
        }
        #input:focus {
            outline: none;
            border-color: #4CAF50;
        }
        button { 
            padding: 12px 20px;
            font-size: 16px;
            background: #4CAF50;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            transition: background 0.3s;
        }
        button:hover {
            background: #45a049;
        }
        button:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .button-group {
            margin-top: 10px;
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }
        .button-group button {
            background: #2196F3;
        }
        .button-group button:hover {
            background: #1976D2;
        }
        #status {
            display: inline-block;
            padding: 8px 15px;
            border-radius: 20px;
            margin-bottom: 20px;
            font-weight: bold;
        }
        .connected { 
            background: #C8E6C9; 
            color: #2E7D32;
        }
        .disconnected { 
            background: #FFCDD2; 
            color: #C62828;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 BotKit WebSocket Test</h1>
        
        <div id="status" class="disconnected">⭕ Disconnected</div>
        
        <div id="messages"></div>
        
        <div class="input-group">
            <input type="text" id="input" placeholder="Type a message..." disabled />
            <button id="sendBtn" onclick="sendMessage()" disabled>Send</button>
        </div>
        
        <div class="button-group">
            <button onclick="sendCommand('/start')">▶️ Start</button>
            <button onclick="sendCommand('/help')">❓ Help</button>
            <button onclick="sendCommand('hello')">👋 Hello</button>
            <button onclick="sendCommand('test_buttons')">🔘 Test Buttons</button>
            <button onclick="sendCommand('What is BotKit?')">🤔 Ask Question</button>
            <button onclick="connect()">🔌 Connect</button>
            <button onclick="disconnect()">🔴 Disconnect</button>
        </div>
    </div>

    <script>
        let ws = null;
        let messageId = 1;
        const userId = Math.floor(Math.random() * 1000000);

        function connect() {
            if (ws && ws.readyState === WebSocket.OPEN) {
                addMessage('Already connected', 'system');
                return;
            }

            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.hostname + ':8081/ws?user_id=' + userId;
            
            addMessage('Connecting to ' + wsUrl + '...', 'system');
            
            ws = new WebSocket(wsUrl);

            ws.onopen = function() {
                updateStatus(true);
                addMessage('✅ Connected to server', 'system');
                document.getElementById('input').disabled = false;
                document.getElementById('sendBtn').disabled = false;
            };

            ws.onmessage = function(event) {
                try {
                    const msg = JSON.parse(event.data);
                    handleMessage(msg);
                } catch (e) {
                    addMessage('Failed to parse message: ' + e, 'error');
                }
            };

            ws.onerror = function(error) {
                addMessage('❌ WebSocket error', 'error');
                console.error('WebSocket error:', error);
            };

            ws.onclose = function() {
                updateStatus(false);
                addMessage('🔴 Disconnected from server', 'system');
                document.getElementById('input').disabled = true;
                document.getElementById('sendBtn').disabled = true;
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
                type: text.startsWith('/') ? 'command' : 'message',
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

        function sendCommand(text) {
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                addMessage('Not connected. Click Connect first!', 'error');
                return;
            }

            const msg = {
                type: text.startsWith('/') ? 'command' : 'message',
                id: 'cmd_' + messageId++,
                user_id: userId,
                chat_id: userId,
                text: text,
                data: {}
            };

            ws.send(JSON.stringify(msg));
            addMessage('You: ' + text, 'sent');
        }

        function handleMessage(msg) {
            console.log('Received:', msg);
            
            if (msg.type === 'connected') {
                addMessage('🎉 Server version: ' + (msg.data.version || 'unknown'), 'system');
            } else if (msg.type === 'response') {
                if (msg.data && msg.data.text) {
                    addMessage('Bot: ' + msg.data.text, 'received');
                }
                
                // Handle buttons if present
                if (msg.data && msg.data.keyboard) {
                    handleKeyboard(msg.data.keyboard);
                }
            } else if (msg.type === 'error') {
                addMessage('Error: ' + msg.error, 'error');
            } else if (msg.type === 'pong') {
                // Ignore pong messages
            } else {
                addMessage('Unknown message type: ' + msg.type, 'system');
            }
        }

        function handleKeyboard(keyboard) {
            if (!keyboard || !Array.isArray(keyboard)) return;
            
            const buttonsHtml = keyboard.map(row => 
                row.map(btn => 
                    '<button onclick="sendCallback(\'' + btn.data + '\')">' + btn.text + '</button>'
                ).join(' ')
            ).join('<br>');
            
            const div = document.createElement('div');
            div.innerHTML = 'Buttons: ' + buttonsHtml;
            div.className = 'message system';
            document.getElementById('messages').appendChild(div);
            scrollToBottom();
        }

        function sendCallback(data) {
            const msg = {
                type: 'callback',
                id: 'cb_' + messageId++,
                user_id: userId,
                chat_id: userId,
                text: data,
                callback_data: data,
                data: { callback_data: data }
            };

            ws.send(JSON.stringify(msg));
            addMessage('Clicked: ' + data, 'sent');
        }

        function addMessage(text, className) {
            const messages = document.getElementById('messages');
            const div = document.createElement('div');
            div.className = 'message ' + className;
            div.textContent = text;
            messages.appendChild(div);
            scrollToBottom();
        }

        function scrollToBottom() {
            const messages = document.getElementById('messages');
            messages.scrollTop = messages.scrollHeight;
        }

        function updateStatus(connected) {
            const status = document.getElementById('status');
            if (connected) {
                status.className = 'connected';
                status.textContent = '🟢 Connected (User ID: ' + userId + ')';
            } else {
                status.className = 'disconnected';
                status.textContent = '⭕ Disconnected';
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
            setTimeout(connect, 500);
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
