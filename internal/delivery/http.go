package delivery

import (
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_server "discord/internal/core/transport/server"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type OLD_SERVER_HANDLERS struct {
	service   domain.Service
	clients   map[int64]*websocket.Conn
	clientsMu sync.Mutex
	logger    *core_logger.Logger
}

func NewServer(logger *core_logger.Logger, s domain.Service) *OLD_SERVER_HANDLERS {
	return &OLD_SERVER_HANDLERS{
		service:   s,
		clients:   make(map[int64]*websocket.Conn),
		clientsMu: sync.Mutex{},
		logger:    logger,
	}
}

func (h *OLD_SERVER_HANDLERS) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{Method: http.MethodPost,
			Path:    "/user",
			Handler: h.SaveUser,
		},
		{Method: http.MethodPost,
			Path:    "/chat",
			Handler: h.SaveChat,
		},
		{Method: http.MethodPost,
			Path:    "/user/avatar",
			Handler: h.UpdateUserAvatar,
		},
		{Method: http.MethodGet,
			Path:    "/user_id",
			Handler: h.GetUserID,
		},
		{Method: http.MethodGet,
			Path:    "/user",
			Handler: h.GetUser,
		},
		{Method: http.MethodGet,
			Path:    "/chats",
			Handler: h.GetUserChats,
		},
		{Method: http.MethodGet,
			Path:    "/messages",
			Handler: h.GetUserMessages,
		},
		{Method: http.MethodGet,
			Path:    "/chats/members",
			Handler: h.GetChatMembers,
		},
		{
			Method: http.MethodGet,
			Path: "/ws",
			Handler: h.HandleConnections,
		}
	}
}

func (sv *OLD_SERVER_HANDLERS) SaveUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var (
		err error
	)
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		sv.sendError(w, core_errors.ErrInvalidJSON.Error(), http.StatusBadRequest, err)
		return
	}
	id, err := sv.service.SaveUser(input.Username, input.Password)
	if err != nil {
		if errors.Is(err, core_errors.ErrUserAlreadyExists) {
			sv.sendError(w, core_errors.ErrUserAlreadyExists.Error(), http.StatusConflict, err)
			return
		}
		sv.sendError(w, core_errors.ErrInternal.Error(), http.StatusInternalServerError, err)
		return
	}

	var resp struct {
		Id int64 `json:"user_id"`
	}
	resp.Id = id
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		sv.sendError(w, core_errors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
		return
	}
}

func (sv *OLD_SERVER_HANDLERS) SaveChat(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var (
		err error
	)
	var input struct {
		ChatName   string `json:"chat_name"`
		Author     int64  `json:"author"`
		TargetUser int64  `json:"target_user"`
	}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		sv.sendError(w, core_errors.ErrInvalidJSON.Error(), http.StatusBadRequest, err)
		return
	}

	id, err := sv.service.SaveChat(input.ChatName, input.Author, input.TargetUser)
	if err != nil {
		sv.sendError(w, core_errors.ErrInternal.Error(), http.StatusInternalServerError, err)
		return
	}
	var resp struct {
		Id int64 `json:"chat_id"`
	}
	resp.Id = id
	w.WriteHeader(http.StatusCreated) // 201
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		sv.sendError(w, core_errors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
		return
	}
}

func (sv *OLD_SERVER_HANDLERS) UpdateUserAvatar(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	user_idParam := r.URL.Query().Get("user_id")
	if user_idParam == "" {
		sv.sendError(w, core_errors.ErrParamMissing.Error(), http.StatusBadRequest, nil)
		return
	}
	user_id, err := strconv.Atoi(user_idParam)
	if err != nil {
		sv.sendError(w, core_errors.ErrParamMissing.Error(), http.StatusBadRequest, err)
		return
	}

	file, header, err := r.FormFile("new_user_avatar")
	if err != nil {
		sv.sendError(w, "ошибка получения файла", http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	file_id, err := sv.service.UploadUserAvatar(r.Context(), int64(user_id), file, header.Filename, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		sv.sendError(w, "ошибка сохранения файла", http.StatusInternalServerError, err)
		return
	}
	var resp struct {
		Id int64 `json:"file_id"`
	}
	resp.Id = file_id
	w.WriteHeader(http.StatusCreated) // 201
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		sv.sendError(w, core_errors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
		return
	}
}

// --- GET
func (sv *OLD_SERVER_HANDLERS) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	log.Warn("Hello from GetUser HTTP")
	w.Header().Set("Content-Type", "application/json")
	var (
		err  error
		user domain.User
	)
	user_idParam := r.URL.Query().Get("user_id")
	if user_idParam == "" {
		sv.sendError(w, core_errors.ErrParamMissing.Error(), http.StatusBadRequest, nil)
		return
	}
	user_id, err := strconv.Atoi(user_idParam)
	if err != nil {
		sv.sendError(w, core_errors.ErrParamMissing.Error(), http.StatusBadRequest, err)
		return
	}

	user, err = sv.service.GetUser(int64(user_id))
	if err != nil {
		sv.sendError(w, core_errors.ErrUserNotFound.Error(), http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusOK) // 200 — Всё Ок
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		sv.sendError(w, core_errors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
		return
	}
}

func (sv *OLD_SERVER_HANDLERS) GetUserChats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var (
		err   error
		chats []domain.Chat
	)
	user_idParam := r.URL.Query().Get("user_id")
	if user_idParam == "" {
		sv.sendError(w, core_errors.ErrParamMissing.Error(), http.StatusBadRequest, nil)
		return
	}
	user_id, err := strconv.Atoi(user_idParam)
	if err != nil {
		sv.sendError(w, core_errors.ErrParamMissing.Error(), http.StatusBadRequest, err)
		return
	}
	chats, err = sv.service.GetUserChats(int64(user_id))
	if err != nil {
		sv.sendError(w, core_errors.ErrInternal.Error(), http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusOK) // 200
	err = json.NewEncoder(w).Encode(chats)
	if err != nil {
		sv.sendError(w, core_errors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
		return
	}
}

func (sv *OLD_SERVER_HANDLERS) GetUserMessages(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var (
		err      error
		messages []domain.Messages
	)
	chat_idParam := r.URL.Query().Get("chat_id")
	if chat_idParam == "" {
		sv.sendError(w, core_errors.ErrParamMissing.Error(), http.StatusBadRequest, nil)
		return
	}
	chat_id, err := strconv.Atoi(chat_idParam)
	if err != nil {
		sv.sendError(w, core_errors.ErrParamMissing.Error(), http.StatusBadRequest, err)
		return
	}

	messages, err = sv.service.GetUserMessages(int64(chat_id))
	if err != nil {
		sv.sendError(w, core_errors.ErrInternal.Error(), http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusOK) // 200
	err = json.NewEncoder(w).Encode(messages)
	if err != nil {
		sv.sendError(w, core_errors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
		return
	}
}

func (sv *OLD_SERVER_HANDLERS) GetUserID(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var (
		err error
		id  int64
	)
	username := r.URL.Query().Get("username")
	if username == "" {
		sv.sendError(w, core_errors.ErrParamMissing.Error(), http.StatusBadRequest, nil)
		return
	}

	id, err = sv.service.GetUserID(username)
	if err != nil {
		sv.sendError(w, core_errors.ErrUserNotFound.Error(), http.StatusNotFound, err)
		return
	}
	var resp struct {
		Id int64 `json:"user_id"`
	}
	resp.Id = id
	w.WriteHeader(http.StatusOK) // 200 — Всё Ок
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		sv.sendError(w, core_errors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
		return
	}
}

func (sv *OLD_SERVER_HANDLERS) GetChatMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var (
		err error
		id  []int64
	)
	chat_idParam := r.URL.Query().Get("chat_id")
	if chat_idParam == "" {
		sv.sendError(w, core_errors.ErrParamMissing.Error(), http.StatusBadRequest, nil)
		return
	}
	chat_id, err := strconv.Atoi(chat_idParam)
	if err != nil {
		sv.sendError(w, core_errors.ErrParamMissing.Error(), http.StatusBadRequest, err)
		return
	}

	id, err = sv.service.GetChatMembers(int64(chat_id))
	if err != nil {
		sv.sendError(w, core_errors.ErrUserNotFound.Error(), http.StatusNotFound, err)
		return
	}
	var resp struct {
		Id []int64 `json:"members"`
	}
	resp.Id = id
	w.WriteHeader(http.StatusOK) // 200 — Всё Ок
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		sv.sendError(w, core_errors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
		return
	}
}
