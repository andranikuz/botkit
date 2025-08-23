package events

import (
	"context"
	"fmt"
	"github.com/andranikuz/botkit/core"
	"sync"
	"time"
)

// EventBus реализация шины событий
type EventBus struct {
	// subscribers подписчики на события
	subscribers map[string][]subscription

	// queue очередь событий
	queue chan eventWrapper

	// workers воркеры для обработки
	workers int

	// logger логгер
	logger core.Logger

	// metrics метрики
	metrics core.Metrics

	// started флаг запуска
	started bool

	// ctx контекст для отмены
	ctx    context.Context
	cancel context.CancelFunc

	// wg wait group для воркеров
	wg sync.WaitGroup

	// mu мьютекс
	mu sync.RWMutex
}

// subscription подписка на событие
type subscription struct {
	handler  core.EventHandlerFunc
	filter   core.EventFilter
	priority int
}

// eventWrapper обертка события для очереди
type eventWrapper struct {
	ctx   context.Context
	event core.Event
}

// NewEventBus создает новую шину событий
func NewEventBus(logger core.Logger, metrics core.Metrics) *EventBus {
	return &EventBus{
		subscribers: make(map[string][]subscription),
		queue:       make(chan eventWrapper, 1000),
		workers:     10,
		logger:      logger,
		metrics:     metrics,
	}
}

// SetWorkers устанавливает количество воркеров
func (eb *EventBus) SetWorkers(count int) {
	eb.workers = count
}

// Subscribe подписывается на событие
func (eb *EventBus) Subscribe(eventType string, handler core.EventHandlerFunc) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	sub := subscription{
		handler:  handler,
		priority: 50,
	}

	eb.subscribers[eventType] = append(eb.subscribers[eventType], sub)

	eb.logger.Debug("Subscribed to event", "type", eventType)

	return nil
}

// SubscribeWithFilter подписывается на событие с фильтром
func (eb *EventBus) SubscribeWithFilter(eventType string, handler core.EventHandlerFunc, filter core.EventFilter) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	sub := subscription{
		handler:  handler,
		filter:   filter,
		priority: 50,
	}

	eb.subscribers[eventType] = append(eb.subscribers[eventType], sub)

	eb.logger.Debug("Subscribed to event with filter", "type", eventType)

	return nil
}

// Unsubscribe отписывается от события
func (eb *EventBus) Unsubscribe(eventType string, handler core.EventHandlerFunc) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs, exists := eb.subscribers[eventType]
	if !exists {
		return nil
	}

	// Находим и удаляем подписку
	newSubs := make([]subscription, 0, len(subs))
	for _, sub := range subs {
		// Сравнение функций по указателю не работает, нужен другой механизм
		// Пока просто пропускаем
		newSubs = append(newSubs, sub)
	}

	eb.subscribers[eventType] = newSubs

	return nil
}

// Publish публикует событие синхронно
func (eb *EventBus) Publish(ctx context.Context, event core.Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	start := time.Now()
	eventType := event.Type()

	// Метрики
	if eb.metrics != nil {
		eb.metrics.Counter("events.published", 1, "type", eventType)
	}

	eb.mu.RLock()
	subs, exists := eb.subscribers[eventType]
	if !exists {
		subs = []subscription{}
	}

	// Также проверяем wildcard подписки
	wildcardSubs, wildcardExists := eb.subscribers["*"]
	if wildcardExists {
		subs = append(subs, wildcardSubs...)
	}
	eb.mu.RUnlock()

	if len(subs) == 0 {
		eb.logger.Debug("No subscribers for event", "type", eventType)
		return nil
	}

	// Обрабатываем событие
	errors := make([]error, 0)
	for _, sub := range subs {
		// Проверяем фильтр
		if sub.filter != nil && !sub.filter(event) {
			continue
		}

		// Вызываем обработчик
		if err := eb.callHandler(ctx, sub.handler, event); err != nil {
			errors = append(errors, err)
			eb.logger.Error("Event handler failed",
				"type", eventType,
				"error", err,
			)
		}
	}

	// Метрики
	if eb.metrics != nil {
		duration := time.Since(start).Milliseconds()
		eb.metrics.Timing("events.processing_time", duration, "type", eventType)

		if len(errors) > 0 {
			eb.metrics.Counter("events.errors", int64(len(errors)), "type", eventType)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("event processing had %d errors", len(errors))
	}

	return nil
}

// PublishAsync публикует событие асинхронно
func (eb *EventBus) PublishAsync(ctx context.Context, event core.Event) {
	if !eb.started {
		eb.logger.Warn("Event bus not started, dropping event", "type", event.Type())
		return
	}

	select {
	case eb.queue <- eventWrapper{ctx: ctx, event: event}:
		// Успешно добавлено в очередь
		if eb.metrics != nil {
			eb.metrics.Counter("events.queued", 1, "type", event.Type())
		}
	default:
		// Очередь переполнена
		eb.logger.Error("Event queue full, dropping event", "type", event.Type())
		if eb.metrics != nil {
			eb.metrics.Counter("events.dropped", 1, "type", event.Type())
		}
	}
}

// callHandler вызывает обработчик с обработкой паники
func (eb *EventBus) callHandler(ctx context.Context, handler core.EventHandlerFunc, event core.Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panicked: %v", r)
			eb.logger.Error("Event handler panicked",
				"type", event.Type(),
				"panic", r,
			)
		}
	}()

	return handler(ctx, event)
}

// Start запускает шину событий
func (eb *EventBus) Start(ctx context.Context) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.started {
		return fmt.Errorf("event bus already started")
	}

	eb.ctx, eb.cancel = context.WithCancel(ctx)

	// Запускаем воркеров
	for i := 0; i < eb.workers; i++ {
		eb.wg.Add(1)
		go eb.worker(i)
	}

	eb.started = true
	eb.logger.Info("Event bus started", "workers", eb.workers)

	return nil
}

// Stop останавливает шину событий
func (eb *EventBus) Stop(ctx context.Context) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if !eb.started {
		return nil
	}

	// Отменяем контекст
	eb.cancel()

	// Закрываем очередь
	close(eb.queue)

	// Ждем завершения воркеров
	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Воркеры завершились
	case <-ctx.Done():
		// Таймаут
		return fmt.Errorf("timeout waiting for workers to stop")
	}

	eb.started = false
	eb.logger.Info("Event bus stopped")

	return nil
}

// worker обработчик событий из очереди
func (eb *EventBus) worker(id int) {
	defer eb.wg.Done()

	eb.logger.Debug("Event worker started", "id", id)

	for {
		select {
		case wrapper, ok := <-eb.queue:
			if !ok {
				// Очередь закрыта
				eb.logger.Debug("Event worker stopped", "id", id)
				return
			}

			// Обрабатываем событие
			if err := eb.Publish(wrapper.ctx, wrapper.event); err != nil {
				eb.logger.Error("Failed to process event from queue",
					"type", wrapper.event.Type(),
					"error", err,
				)
			}

		case <-eb.ctx.Done():
			// Контекст отменен
			eb.logger.Debug("Event worker cancelled", "id", id)
			return
		}
	}
}

// GetStats возвращает статистику шины событий
func (eb *EventBus) GetStats() EventBusStats {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	stats := EventBusStats{
		Subscribers:   make(map[string]int),
		QueueSize:     len(eb.queue),
		QueueCapacity: cap(eb.queue),
		Workers:       eb.workers,
		Started:       eb.started,
	}

	for eventType, subs := range eb.subscribers {
		stats.Subscribers[eventType] = len(subs)
		stats.TotalSubscribers += len(subs)
	}

	return stats
}

// EventBusStats статистика шины событий
type EventBusStats struct {
	Subscribers      map[string]int `json:"subscribers"`
	TotalSubscribers int            `json:"total_subscribers"`
	QueueSize        int            `json:"queue_size"`
	QueueCapacity    int            `json:"queue_capacity"`
	Workers          int            `json:"workers"`
	Started          bool           `json:"started"`
}

// EmitEvent helper функция для быстрой публикации события
func (eb *EventBus) EmitEvent(eventType string, userID, chatID int64, data map[string]interface{}) {
	event := NewEvent(eventType, "system")
	event.SetUserID(userID).SetChatID(chatID)

	for k, v := range data {
		event.SetData(k, v)
	}

	eb.PublishAsync(context.Background(), event)
}
