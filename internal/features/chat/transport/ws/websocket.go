package chat_transport_ws

import (
	"context"
	"discord/internal/core/domain"
	core_logger "discord/internal/core/logger"
	"encoding/json"
	"net/http"
	"strconv"
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

// HandleConnections upgrades to WebSocket
//
//	@Summary		WebSocket connection
//	@Description	Upgrade HTTP to WebSocket for real-time messaging (v2)
//	@Tags			v2 / WebSocket
//	@Param			user_id	query	int	true	"User ID"
//	@Success		101		{object}	object{}
//	@Router			/api/v2/ws [get]
func (h *ChatWSHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	uidStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(uidStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}
	user_id := int64(userID)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.clientsMu.Lock()
	h.clients[user_id] = conn
	h.clientsMu.Unlock()

	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	log.Info("websocket", "user_id", userID, "msg", "user connected")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			h.clientsMu.Lock()
			delete(h.clients, user_id)
			h.clientsMu.Unlock()
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
