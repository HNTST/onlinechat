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
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatService interface {
	SaveMessage(ctx context.Context, message domain.Messages) (int64, error)
	GetChatMembers(ctx context.Context, chat_id int64) ([]int64, error)
	GetChatPartnerIDs(ctx context.Context, userID int64) ([]int64, error)
	GetUserChatIDs(ctx context.Context, userID int64) ([]int64, error)
}

type UserService interface {
	SaveLastSeen(ctx context.Context, userID int64, lastSeen time.Time) error
	GetLastSeen(ctx context.Context, userID int64) (*time.Time, error)
}

type UserStatus struct {
	Status   string
	LastSeen time.Time
}

type safeConn struct {
	conn         *websocket.Conn
	writeMu      sync.Mutex
	closeOnce    sync.Once
	allowedChats map[int64]struct{}
}

func newSafeConn(conn *websocket.Conn, allowedChats map[int64]struct{}) *safeConn {
	return &safeConn{
		conn:         conn,
		allowedChats: allowedChats,
	}
}

func (sc *safeConn) WriteJSON(v interface{}) error {
	sc.writeMu.Lock()
	defer sc.writeMu.Unlock()
	return sc.conn.WriteJSON(v)
}

func (sc *safeConn) WriteMessage(messageType int, data []byte) error {
	sc.writeMu.Lock()
	defer sc.writeMu.Unlock()
	return sc.conn.WriteMessage(messageType, data)
}

func (sc *safeConn) Close() error {
	sc.closeOnce.Do(func() {
		sc.conn.Close()
	})
	return nil
}

func (sc *safeConn) SetPongHandler(h func(string) error) {
	sc.conn.SetPongHandler(h)
}

func (sc *safeConn) SetReadDeadline(t time.Time) {
	sc.conn.SetReadDeadline(t)
}

func (sc *safeConn) SetReadLimit(limit int64) {
	sc.conn.SetReadLimit(limit)
}

func (sc *safeConn) ReadMessage() (messageType int, p []byte, err error) {
	return sc.conn.ReadMessage()
}

func (sc *safeConn) isAllowedChat(chatID int64) bool {
	_, ok := sc.allowedChats[chatID]
	return ok
}

type ChatWSHandler struct {
	log          *core_logger.Logger
	chatService  ChatService
	userService  UserService
	clients      map[int64]map[*safeConn]struct{}
	clientsMu    sync.RWMutex
	userStatuses map[int64]*UserStatus
	statusMu     sync.RWMutex
}

func NewChatWSHandler(ctx context.Context, logger *core_logger.Logger, chatService ChatService, userService UserService) *ChatWSHandler {
	h := &ChatWSHandler{
		log:          logger,
		chatService:  chatService,
		userService:  userService,
		clients:      make(map[int64]map[*safeConn]struct{}),
		userStatuses: make(map[int64]*UserStatus),
	}
	go h.cleanupStaleStatuses(ctx)
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

	lastSeenDB, err := h.userService.GetLastSeen(context.Background(), userID)
	if err != nil || lastSeenDB == nil {
		return "offline", time.Time{}
	}
	return "offline", *lastSeenDB
}

func (h *ChatWSHandler) cleanupStaleStatuses(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.statusMu.Lock()
			now := time.Now()
			for userID, s := range h.userStatuses {
				if s.Status == "offline" && now.Sub(s.LastSeen) > statusCleanup {
					if err := h.userService.SaveLastSeen(context.Background(), userID, s.LastSeen); err != nil {
						h.log.Warn("failed to save last seen on cleanup", "user_id", userID, "error", err)
					}
					delete(h.userStatuses, userID)
				}
			}
			h.statusMu.Unlock()
		}
	}
}

func (h *ChatWSHandler) addClient(userID int64, sc *safeConn) {
	h.clientsMu.Lock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*safeConn]struct{})
	}
	h.clients[userID][sc] = struct{}{}
	h.clientsMu.Unlock()
}

func (h *ChatWSHandler) removeClient(userID int64, sc *safeConn) {
	h.clientsMu.Lock()
	if sockets, ok := h.clients[userID]; ok {
		delete(sockets, sc)
		if len(sockets) == 0 {
			delete(h.clients, userID)
			h.clientsMu.Unlock()
			h.removeUserStatus(userID)
			return
		}
	}
	h.clientsMu.Unlock()
}

func (h *ChatWSHandler) hasAnyConnection(userID int64) bool {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	sockets, ok := h.clients[userID]
	return ok && len(sockets) > 0
}

func (h *ChatWSHandler) writeToAllUserSockets(userID int64, msg interface{}) {
	h.clientsMu.RLock()
	sockets := make([]*safeConn, 0)
	if s, ok := h.clients[userID]; ok {
		for sc := range s {
			sockets = append(sockets, sc)
		}
	}
	h.clientsMu.RUnlock()

	for _, sc := range sockets {
		if err := sc.WriteJSON(msg); err != nil {
			h.removeClient(userID, sc)
		}
	}
}

func (h *ChatWSHandler) notifyUserAboutOnlinePartners(ctx context.Context, userID int64) {
	partnerIDs, err := h.chatService.GetChatPartnerIDs(ctx, userID)
	if err != nil {
		log := core_logger.FromContext(ctx)
		log.Warn("failed to get chat partners for self status check", "user_id", userID, "error", err)
		return
	}

	type partnerStatus struct {
		id       int64
		status   string
		lastSeen time.Time
	}
	var statuses []partnerStatus

	for _, id := range partnerIDs {
		if h.hasAnyConnection(id) {
			status, lastSeen := h.getStatusForPartners(id)
			statuses = append(statuses, partnerStatus{id: id, status: status, lastSeen: lastSeen})
		}
	}

	if len(statuses) == 0 {
		return
	}

	h.clientsMu.RLock()
	sockets := make([]*safeConn, 0)
	if s, ok := h.clients[userID]; ok {
		for sc := range s {
			sockets = append(sockets, sc)
		}
	}
	h.clientsMu.RUnlock()

	for _, ps := range statuses {
		msg := wsStatus{Type: ps.status, UserID: ps.id, LastSeen: ps.lastSeen.Format(time.RFC3339)}
		for _, sc := range sockets {
			if err := sc.WriteJSON(msg); err != nil {
				h.removeClient(userID, sc)
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

	for _, id := range partnerIDs {
		h.writeToAllUserSockets(id, msg)
	}
}

// HandleConnections upgrades to WebSocket
//
//	@Summary		WebSocket-соединение (v2)
//	@Description	Обновление HTTP до WebSocket для обмена сообщениями и отслеживания статуса онлайн/оффлайн/away в реальном времени.
//	@Description	
//	@Description	**Входящее сообщение (от клиента, sender_id берётся из сессии):**
//	@Description	```json
//	@Description	{ "message": "text", "chat_id": 1 }
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

	chatIDs, err := h.chatService.GetUserChatIDs(r.Context(), user_id)
	if err != nil {
		conn.Close()
		return
	}
	allowedChats := make(map[int64]struct{}, len(chatIDs))
	for _, id := range chatIDs {
		allowedChats[id] = struct{}{}
	}

	sc := newSafeConn(conn, allowedChats)
	defer sc.Close()

	h.addClient(user_id, sc)
	h.setUserStatus(user_id, "online")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	log := core_logger.FromContext(ctx)

	sc.SetPongHandler(func(string) error {
		sc.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	sc.SetReadDeadline(time.Now().Add(pongWait))
	sc.SetReadLimit(maxMessageSize)

	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				if err := sc.WriteMessage(websocket.PingMessage, nil); err != nil {
					cancel()
					return
				}
			}
		}
	}()

	log.Info("websocket", "user_id", user_id, "msg", "user connected")

	h.notifyPartners(ctx, user_id, "online")
	h.notifyUserAboutOnlinePartners(ctx, user_id)

	for {
		_, msg, err := sc.ReadMessage()
		if err != nil {
			h.removeClient(user_id, sc)
			if !h.hasAnyConnection(user_id) {
				if err := h.userService.SaveLastSeen(ctx, user_id, time.Now()); err != nil {
					log.Warn("failed to save last seen on disconnect", "user_id", user_id, "error", err)
				}
				h.notifyPartners(ctx, user_id, "offline")
			}
			break
		}
		h.processMessage(ctx, msg, user_id, sc)
	}
}

type wsInput struct {
	Message string `json:"message"`
	ChatID  int64  `json:"chat_id"`
}

func (h *ChatWSHandler) processMessage(ctx context.Context, data []byte, actualSenderID int64, sc *safeConn) {
	var input wsInput
	if err := json.Unmarshal(data, &input); err != nil {
		return
	}
	if !sc.isAllowedChat(input.ChatID) {
		return
	}
	if input.Message == "" {
		return
	}

	m := domain.Messages{
		Message:  input.Message,
		SenderId: actualSenderID,
		ChatId:   input.ChatID,
		Sent_at:  time.Now(),
	}
	id, err := h.chatService.SaveMessage(ctx, m)
	if err != nil {
		return
	}
	m.Id = id

	membersIDs, err := h.chatService.GetChatMembers(ctx, m.ChatId)
	if err != nil {
		return
	}

	for _, memberID := range membersIDs {
		h.writeToAllUserSockets(memberID, m)
	}
}
