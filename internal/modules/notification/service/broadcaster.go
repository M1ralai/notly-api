package notification

import (
	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
	"github.com/M1ralai/notly-api/internal/infrastructure/websocket"
)

type Broadcaster struct {
	hub    *websocket.Hub
	logger *logger.ZapLogger
}

func NewBroadcaster(hub *websocket.Hub, logger *logger.ZapLogger) *Broadcaster {
	return &Broadcaster{hub: hub, logger: logger}
}

func (b *Broadcaster) SyncSignal(userID int, excludeCID string, eventType string, payload map[string]interface{}) {
	b.Publish(userID, excludeCID, eventType, payload)
}

func (b *Broadcaster) Publish(userID int, excludeCID string, eventType string, data map[string]interface{}) {
	/*
		b.logger.Info("Broadcasting notification", map[string]interface{}{
			"user_id":     userID,
			"exclude_cid": excludeCID,
			"type":        eventType,
			"action":      "BROADCAST_NOTIFICATION",
		})
	*/

	message := websocket.NewMessage(eventType, userID, data)
	message.ExcludeCID = excludeCID
	b.hub.PublishToUser(userID, message)
}

func (b *Broadcaster) TaskCreated(userID int, excludeCID string, data map[string]interface{}) {
	b.Publish(userID, excludeCID, websocket.TypeTaskCreated, data)
}

func (b *Broadcaster) TaskCompleted(userID int, excludeCID string, data map[string]interface{}) {
	b.Publish(userID, excludeCID, websocket.TypeTaskCompleted, data)
}

func (b *Broadcaster) HabitCompleted(userID int, excludeCID string, habitID int, title string, streak int) {
	b.Publish(userID, excludeCID, websocket.TypeHabitCompleted, map[string]interface{}{
		"habit_id":       habitID,
		"title":          title,
		"current_streak": streak,
	})
}

func (b *Broadcaster) StreakMilestone(userID int, excludeCID string, habitID int, title string, streak int) {
	b.Publish(userID, excludeCID, websocket.TypeHabitMilestone, map[string]interface{}{
		"habit_id": habitID,
		"title":    title,
		"streak":   streak,
	})
}

func (b *Broadcaster) SyncProgress(userID int, excludeCID string, provider string, progress int) {
	b.Publish(userID, excludeCID, websocket.TypeJobProgress, map[string]interface{}{
		"provider": provider,
		"progress": progress,
	})
}

func (b *Broadcaster) SyncCompleted(userID int, excludeCID string, provider string) {
	b.Publish(userID, excludeCID, websocket.TypeJobCompleted, map[string]interface{}{
		"provider": provider,
	})
}

func (b *Broadcaster) ConflictDetected(userID int, excludeCID string, reason string, start, end string) {
	b.Publish(userID, excludeCID, websocket.TypeCalendarConflict, map[string]interface{}{
		"reason": reason,
		"start":  start,
		"end":    end,
	})
}
