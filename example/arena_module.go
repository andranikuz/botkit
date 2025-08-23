package example

import (
	"context"
	"fmt"
	"github.com/andranikuz/botkit/core"
	"github.com/andranikuz/botkit/events"
	"github.com/andranikuz/botkit/routing"
	"time"
)

// ArenaModule пример модуля арены
type ArenaModule struct {
	name         string
	version      string
	eventBus     core.EventBus
	logger       core.Logger
	arenaService ArenaService // Сервис с бизнес-логикой
}

// ArenaService интерфейс сервиса арены
type ArenaService interface {
	GetOpponents(userID int64) ([]Opponent, error)
	StartFight(userID, opponentID int64) (*FightResult, error)
	GetStats(userID int64) (*ArenaStats, error)
}

// NewArenaModule создает новый модуль арены
func NewArenaModule() *ArenaModule {
	return &ArenaModule{
		name:    "arena",
		version: "1.0.0",
	}
}

// Name возвращает имя модуля
func (m *ArenaModule) Name() string {
	return m.name
}

// Version возвращает версию модуля
func (m *ArenaModule) Version() string {
	return m.version
}

// Init инициализирует модуль
func (m *ArenaModule) Init(deps core.Dependencies) error {
	m.eventBus = deps.EventBus()
	m.logger = deps.Logger()

	// Получаем сервис арены из зависимостей
	if service, ok := deps.Get("arenaService"); ok {
		m.arenaService = service.(ArenaService)
	}

	m.logger.Info("Arena module initialized", "version", m.version)

	return nil
}

// Start запускает модуль
func (m *ArenaModule) Start(ctx context.Context) error {
	m.logger.Info("Arena module started")

	// Публикуем событие о запуске
	event := events.NewModuleLifecycleEvent(m.name, "started", "healthy")
	m.eventBus.PublishAsync(ctx, event)

	return nil
}

// Stop останавливает модуль
func (m *ArenaModule) Stop(ctx context.Context) error {
	m.logger.Info("Arena module stopped")

	// Публикуем событие об остановке
	event := events.NewModuleLifecycleEvent(m.name, "stopped", "stopped")
	m.eventBus.PublishAsync(ctx, event)

	return nil
}

// Routes возвращает маршруты модуля
func (m *ArenaModule) Routes() []routing.RoutePattern {
	return []routing.RoutePattern{
		// Команды
		routing.NewRoute("арена", "/arena").
			Handler(m.handleArenaMenu).
			RequireAuth().
			Meta("arena_menu", "Главное меню арены").
			Build(),

		routing.NewRoute("бой", "/fight").
			Handler(m.handleFightCommand).
			RequireAuth().
			RateLimit(5, 60). // 5 боев в минуту
			Meta("start_fight", "Начать бой").
			Build(),

		routing.NewRoute("статистика арены", "/arena_stats").
			Handler(m.handleStats).
			RequireAuth().
			Meta("arena_stats", "Статистика боев").
			Build(),

		// Callback'и
		routing.NewRoute("arena:fight").
			Type(routing.RouteTypeCallback).
			Handler(m.handleFightCallback).
			RequireAuth().
			Meta("fight_callback", "Выбор противника").
			Build(),

		routing.NewRoute("arena:opponent:{id}").
			Type(routing.RouteTypeCallback).
			Handler(m.handleOpponentSelect).
			RequireAuth().
			Meta("opponent_select", "Подтверждение боя").
			Build(),

		routing.NewRoute("arena:back").
			Type(routing.RouteTypeCallback).
			Handler(m.handleArenaMenu).
			RequireAuth().
			Meta("back_to_menu", "Назад в меню").
			Build(),
	}
}

// Events возвращает подписки на события
func (m *ArenaModule) Events() []core.EventSubscription {
	return []core.EventSubscription{
		{
			EventType: "user.level_up",
			Handler:   m.handleLevelUp,
			Priority:  50,
		},
		{
			EventType: "arena.fight_completed",
			Handler:   m.handleFightCompleted,
			Priority:  100,
		},
	}
}

// APIHandlers возвращает HTTP API handlers
func (m *ArenaModule) APIHandlers() []core.APIHandler {
	return []core.APIHandler{
		{
			Method:      "GET",
			Path:        "/opponents",
			Handler:     m.apiGetOpponents,
			Description: "Get list of opponents",
		},
		{
			Method:      "POST",
			Path:        "/fight",
			Handler:     m.apiStartFight,
			Description: "Start a fight",
		},
		{
			Method:      "GET",
			Path:        "/stats",
			Handler:     m.apiGetStats,
			Description: "Get arena statistics",
		},
	}
}

// === Handlers ===

// handleArenaMenu обрабатывает главное меню арены
func (m *ArenaModule) handleArenaMenu(ctx core.UniversalContext) core.Response {
	text := `⚔️ <b>Арена</b>
	
Добро пожаловать на арену!
Здесь вы можете сражаться с другими игроками.

💰 Победа: +100 монет
📈 Рейтинг: +10 очков
⚡ Энергия: -5`

	// Создаем клавиатуру
	keyboard := NewArenaKeyboard().MainMenu()

	return core.NewMessage(text).
		WithKeyboard(keyboard).
		WithParseMode(core.ParseModeHTML)
}

// handleFightCommand обрабатывает команду боя
func (m *ArenaModule) handleFightCommand(ctx core.UniversalContext) core.Response {
	userID := ctx.GetUserID()

	// Получаем список противников
	opponents, err := m.arenaService.GetOpponents(userID)
	if err != nil {
		return core.NewMessage("❌ Ошибка при поиске противников")
	}

	if len(opponents) == 0 {
		return core.NewMessage("🔍 Нет доступных противников")
	}

	text := "⚔️ <b>Выберите противника:</b>\n\n"

	// Создаем клавиатуру с противниками
	keyboard := NewArenaKeyboard()

	for i, opp := range opponents {
		text += fmt.Sprintf("%d. %s (Ур. %d, Сила: %d)\n",
			i+1, opp.Name, opp.Level, opp.Power)

		keyboard.AddOpponent(opp)
	}

	keyboard.AddBackButton()

	return core.NewMessage(text).
		WithKeyboard(keyboard.Build()).
		WithParseMode(core.ParseModeHTML)
}

// handleFightCallback обрабатывает callback выбора боя
func (m *ArenaModule) handleFightCallback(ctx core.UniversalContext) core.Response {
	return m.handleFightCommand(ctx)
}

// handleOpponentSelect обрабатывает выбор противника
func (m *ArenaModule) handleOpponentSelect(ctx core.UniversalContext) core.Response {
	opponentID, ok := ctx.GetIntParam("id")
	if !ok {
		return core.NewMessage("❌ Неверный ID противника")
	}

	userID := ctx.GetUserID()

	// Начинаем бой
	result, err := m.arenaService.StartFight(userID, int64(opponentID))
	if err != nil {
		return core.NewMessage("❌ Ошибка при начале боя: " + err.Error())
	}

	// Формируем отчет о бое
	text := m.formatFightResult(result)

	// Публикуем событие о завершении боя
	event := events.NewEvent("arena.fight_completed", m.name)
	event.SetUserID(userID).
		SetData("opponent_id", opponentID).
		SetData("winner", result.Winner).
		SetData("reward", result.Reward)

	m.eventBus.PublishAsync(ctx.Context(), event)

	// Клавиатура после боя
	keyboard := NewArenaKeyboard().AfterFight()

	return core.NewEditMessage(ctx.GetMessageID(), text).
		WithKeyboard(keyboard).
		WithParseMode(core.ParseModeHTML)
}

// handleStats обрабатывает статистику
func (m *ArenaModule) handleStats(ctx core.UniversalContext) core.Response {
	userID := ctx.GetUserID()

	stats, err := m.arenaService.GetStats(userID)
	if err != nil {
		return core.NewMessage("❌ Ошибка при получении статистики")
	}

	text := fmt.Sprintf(`📊 <b>Статистика арены</b>

⚔️ Всего боев: %d
✅ Побед: %d
❌ Поражений: %d
📈 Рейтинг: %d
🏆 Место в рейтинге: #%d
💰 Заработано монет: %d`,
		stats.TotalFights,
		stats.Wins,
		stats.Losses,
		stats.Rating,
		stats.Rank,
		stats.TotalEarned,
	)

	return core.NewMessage(text).
		WithParseMode(core.ParseModeHTML)
}

// === Event Handlers ===

func (m *ArenaModule) handleLevelUp(ctx context.Context, event core.Event) error {
	m.logger.Info("User leveled up", "user_id", event.UserID())
	// Логика обработки повышения уровня
	return nil
}

func (m *ArenaModule) handleFightCompleted(ctx context.Context, event core.Event) error {
	m.logger.Info("Fight completed",
		"user_id", event.UserID(),
		"winner", event.Data()["winner"],
	)
	// Логика обработки завершения боя
	return nil
}

// === API Handlers ===

func (m *ArenaModule) apiGetOpponents(ctx context.Context, req core.APIRequest) (core.APIResponse, error) {
	opponents, err := m.arenaService.GetOpponents(req.UserID)
	if err != nil {
		return core.APIResponse{Status: 500, Error: err}, err
	}

	return core.APIResponse{
		Status: 200,
		Body:   opponents,
	}, nil
}

func (m *ArenaModule) apiStartFight(ctx context.Context, req core.APIRequest) (core.APIResponse, error) {
	// Парсим тело запроса
	body := req.Body.(map[string]interface{})
	opponentID := int64(body["opponent_id"].(float64))

	result, err := m.arenaService.StartFight(req.UserID, opponentID)
	if err != nil {
		return core.APIResponse{Status: 500, Error: err}, err
	}

	return core.APIResponse{
		Status: 200,
		Body:   result,
	}, nil
}

func (m *ArenaModule) apiGetStats(ctx context.Context, req core.APIRequest) (core.APIResponse, error) {
	stats, err := m.arenaService.GetStats(req.UserID)
	if err != nil {
		return core.APIResponse{Status: 500, Error: err}, err
	}

	return core.APIResponse{
		Status: 200,
		Body:   stats,
	}, nil
}

// === Helper Methods ===

func (m *ArenaModule) formatFightResult(result *FightResult) string {
	var text string

	if result.Winner == "player" {
		text = "🎉 <b>Победа!</b>\n\n"
	} else {
		text = "😔 <b>Поражение</b>\n\n"
	}

	text += fmt.Sprintf(`📊 Результаты боя:
• Раундов: %d
• Нанесено урона: %d
• Получено урона: %d
• Критических ударов: %d

💰 Награда: %d монет
📈 Рейтинг: %+d`,
		result.Rounds,
		result.DamageDealt,
		result.DamageTaken,
		result.CriticalHits,
		result.Reward,
		result.RatingChange,
	)

	return text
}

// === Types ===

type Opponent struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Level int    `json:"level"`
	Power int    `json:"power"`
}

type FightResult struct {
	Winner       string    `json:"winner"`
	Rounds       int       `json:"rounds"`
	DamageDealt  int       `json:"damage_dealt"`
	DamageTaken  int       `json:"damage_taken"`
	CriticalHits int       `json:"critical_hits"`
	Reward       int       `json:"reward"`
	RatingChange int       `json:"rating_change"`
	Timestamp    time.Time `json:"timestamp"`
}

type ArenaStats struct {
	TotalFights int `json:"total_fights"`
	Wins        int `json:"wins"`
	Losses      int `json:"losses"`
	Rating      int `json:"rating"`
	Rank        int `json:"rank"`
	TotalEarned int `json:"total_earned"`
}
