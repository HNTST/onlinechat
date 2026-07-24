package chat_transport_ws

import (
	"context"
	"discord/internal/core/domain"
	core_logger "discord/internal/core/logger"
	core_http_middleware "discord/internal/core/transport/middleware"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait      = 60 * time.Second
	pingPeriod    = 54 * time.Second
	statusCleanup = 24 * time.Hour
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatService interface {
	SaveMessage(ctx context.Context, message domain.Messages) error
	GetChatMembers(ctx context.Context, chat_id int64) ([]int64, error)
	GetChatPartnerIDs(ctx context.Context, userID int64) ([]int64, error)
	SaveLastSeen(ctx context.Context, userID int64, lastSeen time.Time) error
	GetLastSeen(ctx context.Context, userID int64) (*time.Time, error)
}

type UserStatus struct {
	Status   string
	LastSeen time.Time
}

type ChatWSHandler struct {
	chatService  ChatService
	clients      map[int64]*websocket.Conn
	clientsMu    sync.Mutex
	writeMu      sync.Mutex
	userStatuses map[int64]*UserStatus
	statusMu     sync.RWMutex
}

func NewChatWSHandler(chatService ChatService) *ChatWSHandler {
	h := &ChatWSHandler{
		chatService:  chatService,
		clients:      make(map[int64]*websocket.Conn),
		userStatuses: make(map[int64]*UserStatus),
	}
	go h.cleanupStaleStatuses()
	return h
}

type wsStatus struct {
	Type     string `json:"type"`
	UserID   int64  `json:"user_id"`
	LastSeen string `json:"last_seen,omitempty"`
}

func (h *ChatWSHandler) setUserStatus(userID int64, status string) {
	h.statusMu.Lock()
	h.userStatuses[userID] = &UserStatus{
		Status:   status,
		LastSeen: time.Now(),
	}
	h.statusMu.Unlock()
}

func (h *ChatWSHandler) removeUserStatus(userID int64) {
	h.statusMu.Lock()
	if s, ok := h.userStatuses[userID]; ok {
		s.Status = "offline"
		s.LastSeen = time.Now()
	}
	h.statusMu.Unlock()
}

func (h *ChatWSHandler) getUserStatus(userID int64) (string, time.Time, bool) {
	h.statusMu.RLock()
	defer h.statusMu.RUnlock()
	s, ok := h.userStatuses[userID]
	if !ok {
		return "", time.Time{}, false
	}
	return s.Status, s.LastSeen, true
}

func (h *ChatWSHandler) getStatusForPartners(userID int64) (string, time.Time) {
	status, lastSeen, ok := h.getUserStatus(userID)
	if ok {
		return status, lastSeen
	}

	lastSeenDB, err := h.chatService.GetLastSeen(context.Background(), userID)
	if err != nil || lastSeenDB == nil {
		return "offline", time.Time{}
	}
	return "offline", *lastSeenDB
}

func (h *ChatWSHandler) cleanupStaleStatuses() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		h.statusMu.Lock()
		now := time.Now()
		for userID, s := range h.userStatuses {
			if s.Status == "offline" && now.Sub(s.LastSeen) > statusCleanup {
				if err := h.chatService.SaveLastSeen(context.Background(), userID, s.LastSeen); err != nil {
					log := core_logger.NewLogger()
					log.Warn("failed to save last seen on cleanup", "user_id", userID, "error", err)
				}
				delete(h.userStatuses, userID)
			}
		}
		h.statusMu.Unlock()
	}
}

func (h *ChatWSHandler) notifyUserAboutOnlinePartners(ctx context.Context, userID int64) {
	partnerIDs, err := h.chatService.GetChatPartnerIDs(ctx, userID)
	if err != nil {
		log := core_logger.FromContext(ctx)
		log.Warn("failed to get chat partners for self status check", "user_id", userID, "error", err)
		return
	}

	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	selfConn, ok := h.clients[userID]
	if !ok {
		return
	}

	for _, id := range partnerIDs {
		if _, online := h.clients[id]; online {
			status, lastSeen := h.getStatusForPartners(id)
			msg := wsStatus{Type: status, UserID: id, LastSeen: lastSeen.Format(time.RFC3339)}
			h.writeMu.Lock()
			err := selfConn.WriteJSON(msg)
			h.writeMu.Unlock()
			if err != nil {
				selfConn.Close()
				delete(h.clients, userID)
				return
			}
		}
	}
}

func (h *ChatWSHandler) notifyPartners(ctx context.Context, userID int64, status string) {
	partnerIDs, err := h.chatService.GetChatPartnerIDs(ctx, userID)
	if err != nil {
		log := core_logger.FromContext(ctx)
		log.Warn("failed to get chat partners for status notification", "user_id", userID, "error", err)
		return
	}

	h.statusMu.RLock()
	us := h.userStatuses[userID]
	h.statusMu.RUnlock()

	msg := wsStatus{Type: status, UserID: userID}
	if us != nil {
		msg.LastSeen = us.LastSeen.Format(time.RFC3339)
	}

	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	for _, id := range partnerIDs {
		if conn, ok := h.clients[id]; ok {
			h.writeMu.Lock()
			err := conn.WriteJSON(msg)
			h.writeMu.Unlock()
			if err != nil {
				conn.Close()
				delete(h.clients, id)
			}
		}
	}
}

func (h *ChatWSHandler) removeClient(userID int64) {
	h.clientsMu.Lock()
	if conn, ok := h.clients[userID]; ok {
		conn.Close()
		delete(h.clients, userID)
	}
	h.clientsMu.Unlock()
}

// HandleConnections upgrades to WebSocket
//
//	@Summary		WebSocket-соединение (v2)
//	@Description	Обновление HTTP до WebSocket для обмена сообщениями и отслеживания статуса онлайн/оффлайн/away в реальном времени.
//	@Description	
//	@Description	**Входящее сообщение (от клиента):**
//	@Description	```json
//	@Description	{ "message": "text", "chat_id": 1, "sender_id": 123 }
//	@Description	```
//	@Description	
//	@Description	**Исходящие сообщения (от сервера):**
//	@Description	
//	@Description	1. Новое сообщение в чате:
//	@Description	```json
//	@Description	{ "id": 456, "message": "text", "chat_id": 1, "sender_id": 123, "sent_at": "2025-01-01T00:00:00Z" }
//	@Description	```
//	@Description	
//	@Description	2. Статус пользователя:
//	@Description	```json
//	@Description	{ "type": "online|away|offline", "user_id": 123, "last_seen": "2025-01-01T00:00:00Z" }
//	@Description	```
//	@Description	
//	@Description	**Статусы:** online (в приложении), away (свернул/бездействует), offline (отключился)
//	@Description	**Heartbeat:** Ping каждые 54с, Pong ожидается в течение 60с.
//	@Tags			v2 / WebSocket
//	@Success		101		{object}	object{}
//	@Failure		401		{object}	object{}
//	@Router			/api/v2/ws [get]
func (h *ChatWSHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	user_id, ok := r.Context().Value(core_http_middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "требуется авторизация", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.clientsMu.Lock()
	h.clients[user_id] = conn
	h.clientsMu.Unlock()

	h.setUserStatus(user_id, "online")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	log := core_logger.FromContext(ctx)

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	conn.SetReadDeadline(time.Now().Add(pongWait))

	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				h.writeMu.Lock()
				err := conn.WriteMessage(websocket.PingMessage, nil)
				h.writeMu.Unlock()
				if err != nil {
					return
				}
			}
		}
	}()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	log.Info("websocket", "user_id", user_id, "msg", "user connected")

	h.notifyPartners(ctx, user_id, "online")
	h.notifyUserAboutOnlinePartners(ctx, user_id)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			h.removeClient(user_id)
			h.removeUserStatus(user_id)
			if err := h.chatService.SaveLastSeen(ctx, user_id, time.Now()); err != nil {
				log.Warn("failed to save last seen on disconnect", "user_id", user_id, "error", err)
			}
			h.notifyPartners(ctx, user_id, "offline")
			break
		}
		h.processMessage(ctx, msg, user_id)
	}
}

type wsInput struct {
	Message  string `json:"message"`
	ChatID   int64  `json:"chat_id"`
	SenderID int64  `json:"sender_id"`
}

func (h *ChatWSHandler) processMessage(ctx context.Context, data []byte, actualSenderID int64) {
	var input wsInput
	if err := json.Unmarshal(data, &input); err != nil {
		return
	}
	if input.SenderID != actualSenderID {
		return
	}
	m := domain.Messages{
		Message:  input.Message,
		SenderId: input.SenderID,
		ChatId:   input.ChatID,
		Sent_at:  time.Now(),
	}
	if err := h.chatService.SaveMessage(ctx, m); err != nil {
		return
	}

	membersIDs, err := h.chatService.GetChatMembers(ctx, m.ChatId)
	if err != nil {
		return
	}
	h.clientsMu.Lock()
	for _, id := range membersIDs {
		if clientConn, ok := h.clients[id]; ok {
			h.writeMu.Lock()
			err := clientConn.WriteJSON(m)
			h.writeMu.Unlock()
			if err != nil {
				clientConn.Close()
				delete(h.clients, id)
			}
		}
	}
	h.clientsMu.Unlock()
}
