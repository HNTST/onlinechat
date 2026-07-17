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
	pongWait   = 60 * time.Second
	pingPeriod = 54 * time.Second
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
	writeMu     sync.Mutex
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
			msg := wsStatus{Type: "online", UserID: id}
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

	msg := wsStatus{Type: status, UserID: userID}

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
//	@Description	Обновление HTTP до WebSocket для обмена сообщениями и отслеживания статуса онлайн/оффлайн в реальном времени.
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
//	@Description	2. Пользователь онлайн (приходит всем чат-партнёрам при подключении + новому пользователю о уже подключённых партнёрах):
//	@Description	```json
//	@Description	{ "type": "online", "user_id": 123 }
//	@Description	```
//	@Description	
//	@Description	3. Пользователь оффлайн (приходит всем чат-партнёрам при отключении):
//	@Description	```json
//	@Description	{ "type": "offline", "user_id": 123 }
//	@Description	```
//	@Description	
//	@Description	**Поведение:** При подключении нового пользователя сервер (1) уведомляет всех его чат-партнёров, что он онлайн, и (2) отправляет ему список уже подключённых партнёров. Уведомления приходят только для пользователей, состоящих в одном чате.
//	@Description	
//	@Description	**Heartbeat:** Сервер отправляет Ping каждые 54с. Клиент должен отвечать Pong (встроено в большинство WS-библиотек). Если Pong не получен в течение 60с — соединение закрывается.
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
