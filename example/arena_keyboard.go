package example

import (
	"fmt"
	"github.com/andranikuz/botkit/core"
)

// ArenaKeyboard клавиатура для модуля арены
type ArenaKeyboard struct {
	buttons [][]core.Button
}

// NewArenaKeyboard создает новую клавиатуру арены
func NewArenaKeyboard() *ArenaKeyboard {
	return &ArenaKeyboard{
		buttons: make([][]core.Button, 0),
	}
}

// Type возвращает тип клавиатуры
func (k *ArenaKeyboard) Type() core.KeyboardType {
	return core.KeyboardTypeInline
}

// Buttons возвращает кнопки
func (k *ArenaKeyboard) Buttons() [][]core.Button {
	return k.buttons
}

// Options возвращает опции клавиатуры
func (k *ArenaKeyboard) Options() core.KeyboardOptions {
	return core.KeyboardOptions{}
}

// Build возвращает клавиатуру как интерфейс
func (k *ArenaKeyboard) Build() core.Keyboard {
	return k
}

// MainMenu создает главное меню арены
func (k *ArenaKeyboard) MainMenu() *ArenaKeyboard {
	k.buttons = [][]core.Button{
		{
			{
				Text: "⚔️ В бой",
				Type: core.ButtonTypeCallback,
				Route: &core.Route{
					Module: "arena",
					Action: "fight",
				},
			},
			{
				Text: "📊 Статистика",
				Type: core.ButtonTypeCallback,
				Route: &core.Route{
					Module: "arena",
					Action: "stats",
				},
			},
		},
		{
			{
				Text: "🏆 Рейтинг",
				Type: core.ButtonTypeCallback,
				Route: &core.Route{
					Module: "arena",
					Action: "leaderboard",
				},
			},
			{
				Text: "ℹ️ Правила",
				Type: core.ButtonTypeCallback,
				Route: &core.Route{
					Module: "arena",
					Action: "rules",
				},
			},
		},
		{
			{
				Text: "⬅️ В город",
				Type: core.ButtonTypeCallback,
				Route: &core.Route{
					Module: "city",
					Action: "main",
				},
			},
		},
	}

	return k
}

// AddOpponent добавляет кнопку противника
func (k *ArenaKeyboard) AddOpponent(opp Opponent) *ArenaKeyboard {
	btn := core.Button{
		Text: fmt.Sprintf("%s (Ур. %d)", opp.Name, opp.Level),
		Type: core.ButtonTypeCallback,
		Route: &core.Route{
			Module: "arena",
			Action: "opponent",
			Params: map[string]interface{}{
				"id": opp.ID,
			},
		},
	}

	// Добавляем по 2 кнопки в строку
	if len(k.buttons) == 0 || len(k.buttons[len(k.buttons)-1]) >= 2 {
		k.buttons = append(k.buttons, []core.Button{btn})
	} else {
		lastRow := len(k.buttons) - 1
		k.buttons[lastRow] = append(k.buttons[lastRow], btn)
	}

	return k
}

// AddBackButton добавляет кнопку "Назад"
func (k *ArenaKeyboard) AddBackButton() *ArenaKeyboard {
	k.buttons = append(k.buttons, []core.Button{
		{
			Text: "⬅️ Назад",
			Type: core.ButtonTypeCallback,
			Route: &core.Route{
				Module: "arena",
				Action: "back",
			},
		},
	})

	return k
}

// AfterFight клавиатура после боя
func (k *ArenaKeyboard) AfterFight() *ArenaKeyboard {
	k.buttons = [][]core.Button{
		{
			{
				Text: "🔄 Еще бой",
				Type: core.ButtonTypeCallback,
				Route: &core.Route{
					Module: "arena",
					Action: "fight",
				},
			},
			{
				Text: "📊 Статистика",
				Type: core.ButtonTypeCallback,
				Route: &core.Route{
					Module: "arena",
					Action: "stats",
				},
			},
		},
		{
			{
				Text: "⬅️ В меню арены",
				Type: core.ButtonTypeCallback,
				Route: &core.Route{
					Module: "arena",
					Action: "back",
				},
			},
		},
	}

	return k
}
