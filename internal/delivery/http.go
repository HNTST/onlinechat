package delivery

import (
	"discord/internal/domain"
	"discord/internal/errrors"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	service   domain.Service
	clients   map[int64]*websocket.Conn
	clientsMu sync.Mutex
	logger    slog.Logger
}

func NewServer(logger *slog.Logger, s domain.Service) *Server {
	return &Server{
		service:   s,
		clients:   make(map[int64]*websocket.Conn),
		clientsMu: sync.Mutex{},
		logger:    *logger,
	}
}

func (sv *Server) SaveUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			sv.sendError(w, errrors.ErrInvalidJSON.Error(), http.StatusBadRequest, err)
			return
		}
		id, err := sv.service.SaveUser(input.Username, input.Password)
		if err != nil {
			if errors.Is(err, errrors.ErrUserAlreadyExists) {
				sv.sendError(w, errrors.ErrUserAlreadyExists.Error(), http.StatusConflict, err)
				return
			}
			sv.sendError(w, errrors.ErrInternal.Error(), http.StatusInternalServerError, err)
			return
		}
		var resp struct {
			Id int64 `json:"user_id"`
		}
		resp.Id = id
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			sv.sendError(w, errrors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
			return
		}
	}
}
func (sv *Server) SaveChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			sv.sendError(w, errrors.ErrInvalidJSON.Error(), http.StatusBadRequest, err)
			return
		}

		id, err := sv.service.SaveChat(input.ChatName, input.Author, input.TargetUser)
		if err != nil {
			sv.sendError(w, errrors.ErrInternal.Error(), http.StatusInternalServerError, err)
			return
		}
		var resp struct {
			Id int64 `json:"chat_id"`
		}
		resp.Id = id
		w.WriteHeader(http.StatusCreated) // 201
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			sv.sendError(w, errrors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
			return
		}
	}
}
func (sv *Server) UpdateUserAvatar() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		user_idParam := r.URL.Query().Get("user_id")
		if user_idParam == "" {
			sv.sendError(w, errrors.ErrParamMissing.Error(), http.StatusBadRequest, nil)
			return
		}
		user_id, err := strconv.Atoi(user_idParam)
		if err != nil {
			sv.sendError(w, errrors.ErrParamMissing.Error(), http.StatusBadRequest, err)
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
			sv.sendError(w, errrors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
			return
		}
	}
}

// --- GET
func (sv *Server) GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var (
			err  error
			user domain.User
		)
		user_idParam := r.URL.Query().Get("user_id")
		if user_idParam == "" {
			sv.sendError(w, errrors.ErrParamMissing.Error(), http.StatusBadRequest, nil)
			return
		}
		user_id, err := strconv.Atoi(user_idParam)
		if err != nil {
			sv.sendError(w, errrors.ErrParamMissing.Error(), http.StatusBadRequest, err)
			return
		}

		user, err = sv.service.GetUser(int64(user_id))
		if err != nil {
			sv.sendError(w, errrors.ErrUserNotFound.Error(), http.StatusNotFound, err)
			return
		}
		w.WriteHeader(http.StatusOK) // 200 — Всё Ок
		err = json.NewEncoder(w).Encode(user)
		if err != nil {
			sv.sendError(w, errrors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
			return
		}
	}
}
func (sv *Server) GetUserChats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var (
			err   error
			chats []domain.Chat
		)
		user_idParam := r.URL.Query().Get("user_id")
		if user_idParam == "" {
			sv.sendError(w, errrors.ErrParamMissing.Error(), http.StatusBadRequest, nil)
			return
		}
		user_id, err := strconv.Atoi(user_idParam)
		if err != nil {
			sv.sendError(w, errrors.ErrParamMissing.Error(), http.StatusBadRequest, err)
			return
		}
		chats, err = sv.service.GetUserChats(int64(user_id))
		if err != nil {
			sv.sendError(w, errrors.ErrInternal.Error(), http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		err = json.NewEncoder(w).Encode(chats)
		if err != nil {
			sv.sendError(w, errrors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
			return
		}
	}
}
func (sv *Server) GetUserMessages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var (
			err      error
			messages []domain.Messages
		)
		chat_idParam := r.URL.Query().Get("chat_id")
		if chat_idParam == "" {
			sv.sendError(w, errrors.ErrParamMissing.Error(), http.StatusBadRequest, nil)
			return
		}
		chat_id, err := strconv.Atoi(chat_idParam)
		if err != nil {
			sv.sendError(w, errrors.ErrParamMissing.Error(), http.StatusBadRequest, err)
			return
		}

		messages, err = sv.service.GetUserMessages(int64(chat_id))
		if err != nil {
			sv.sendError(w, errrors.ErrInternal.Error(), http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		err = json.NewEncoder(w).Encode(messages)
		if err != nil {
			sv.sendError(w, errrors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
			return
		}
	}
}
func (sv *Server) GetUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var (
			err error
			id  int64
		)
		username := r.URL.Query().Get("username")
		if username == "" {
			sv.sendError(w, errrors.ErrParamMissing.Error(), http.StatusBadRequest, nil)
			return
		}

		id, err = sv.service.GetUserID(username)
		if err != nil {
			sv.sendError(w, errrors.ErrUserNotFound.Error(), http.StatusNotFound, err)
			return
		}
		var resp struct {
			Id int64 `json:"user_id"`
		}
		resp.Id = id
		w.WriteHeader(http.StatusOK) // 200 — Всё Ок
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			sv.sendError(w, errrors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
			return
		}
	}
}
