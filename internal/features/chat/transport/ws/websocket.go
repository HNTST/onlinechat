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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatService interface {
	SaveMessage(ctx context.Context, message domain.Messages) error
	GetChatMembers(ctx context.Context, chat_id int64) ([]int64, error)
	GetChatPartnerIDs(ctx context.Context, userID int64) ([]int64, error)
}

type ChatWSHandler struct {
	chatService ChatService
	clients     map[int64]*websocket.Conn
	clientsMu   sync.Mutex
}

func NewChatWSHandler(chatService ChatService) *ChatWSHandler {
	return &ChatWSHandler{
		chatService: chatService,
		clients:     make(map[int64]*websocket.Conn),
	}
}

type wsStatus struct {
	Type   string `json:"type"`
	UserID int64  `json:"user_id"`
}

func (h *ChatWSHandler) notifyPartners(ctx context.Context, userID int64, status string) {
	partnerIDs, err := h.chatService.GetChatPartnerIDs(ctx, userID)
	if err != nil {
		log := core_logger.FromContext(ctx)
		log.Warn("failed to get chat partners for status notification", "user_id", userID, "error", err)
		return
	}

	msg := wsStatus{Type: status, UserID: userID}

	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	for _, id := range partnerIDs {
		if conn, ok := h.clients[id]; ok {
			if err := conn.WriteJSON(msg); err != nil {
				conn.Close()
				delete(h.clients, id)
			}
		}
	}
}

// HandleConnections upgrades to WebSocket
//
//	@Summary		WebSocket-соединение (v2)
//	@Description	Обновление HTTP до WebSocket для обмена сообщениями и отслеживания статуса онлайн/оффлайн в реальном времени.
//	@Description	
//	@Description	**Входящие сообщения (от клиента):**
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
//	@Description	2. Пользователь онлайн:
//	@Description	```json
//	@Description	{ "type": "online", "user_id": 123 }
//	@Description	```
//	@Description	
//	@Description	3. Пользователь оффлайн:
//	@Description	```json
//	@Description	{ "type": "offline", "user_id": 123 }
//	@Description	```
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

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	log := core_logger.FromContext(ctx)
	log.Info("websocket", "user_id", user_id, "msg", "user connected")

	h.notifyPartners(ctx, user_id, "online")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			h.clientsMu.Lock()
			delete(h.clients, user_id)
			h.clientsMu.Unlock()
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
			err := clientConn.WriteJSON(m)
			if err != nil {
				clientConn.Close()
				delete(h.clients, id)
			}
		}
	}
	h.clientsMu.Unlock()
}
