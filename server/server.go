package server

import (
	db "discord/database"
	md "discord/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type serverRepo interface {
	SaveUser() http.HandlerFunc
	SaveChat() http.HandlerFunc

	GetUser() http.HandlerFunc
	GetUserID() http.HandlerFunc
	GetUserChats() http.HandlerFunc
	GetUserMessages() http.HandlerFunc

	HandleConnections() http.HandlerFunc
	Broadcast(msg []byte)
}
type Server struct {
	repo      db.DBRepo
	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex
}

func New(r db.DBRepo) *Server {
	return &Server{
		repo:      r,
		clients:   make(map[*websocket.Conn]bool),
		clientsMu: sync.Mutex{},
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
			http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest) // 400
			fmt.Println("Ошибка декодирования джон запроса: ", err)
			return
		}
		id, err := sv.repo.SaveUser(input.Username, input.Password)
		if err != nil {
			http.Error(w, `{"error": "Ошибка сохранения в базу"}`, http.StatusInternalServerError) // 500
			fmt.Println("Ошибка сохранения в базу данных: ", err)
			return
		}
		var resp struct {
			Id int64 `json:"user_id"`
		}
		resp.Id = id
		w.WriteHeader(http.StatusCreated) // 201 — Успешно создано
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			fmt.Println("Ошибка кодирования джон ответа: ", err)
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
			http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest) // 400
			fmt.Println("Ошибка декодирования джон запроса: ", err)
			return
		}
		id, err := sv.repo.SaveChatWithAuthor(input.ChatName, input.Author, input.TargetUser)
		if err != nil {
			http.Error(w, `{"error": "Ошибка создания чата или приязки пользователя к чату"}`, http.StatusInternalServerError) // 500
			fmt.Println("Ошибка сохранения в базу данных: ", err)
			return
		}
		var resp struct {
			Id int64 `json:"chat_id"`
		}
		resp.Id = id
		w.WriteHeader(http.StatusCreated) // 201
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			fmt.Println("Ошибка кодирования джон ответа: ", err)
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
			user md.User
		)
		user_idParam := r.URL.Query().Get("user_id")
		if user_idParam == "" {
			http.Error(w, `{"error": "Отсутствует параметр user_id"}`, http.StatusBadRequest)
			fmt.Println("Ошибка: user_id не указан в запросе")
			return
		}
		user_id, err := strconv.Atoi(user_idParam)
		if err != nil {
			http.Error(w, `{"error": "Ошибка конвертации параметра"}`, http.StatusBadRequest)
			fmt.Println("Ошибка конвертации параметра")
			return
		}

		user, err = sv.repo.GetUser(int64(user_id))
		if err != nil {
			http.Error(w, `{"error": "Пользователь не найден"}`, http.StatusNotFound) // 404
			fmt.Println("Ошибка сохранения в базу данных: ", err)
			return
		}
		w.WriteHeader(http.StatusOK) // 200 — Всё Ок
		err = json.NewEncoder(w).Encode(user)
		if err != nil {
			fmt.Println("Ошибка кодирования джон ответа: ", err)
			return
		}
	}
}
func (sv *Server) GetUserChats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var (
			err   error
			chats []md.Chat
		)
		user_idParam := r.URL.Query().Get("user_id")
		if user_idParam == "" {
			http.Error(w, `{"error": "Отсутствует параметр user_id"}`, http.StatusBadRequest)
			fmt.Println("Ошибка: user_id не указан в запросе")
			return
		}
		user_id, err := strconv.Atoi(user_idParam)
		if err != nil {
			http.Error(w, `{"error": "Ошибка конвертации параметра"}`, http.StatusBadRequest)
			fmt.Println("Ошибка конвертации параметра")
			return
		}
		chats, err = sv.repo.GetUserChats(int64(user_id))
		if err != nil {
			http.Error(w, `{"error": "Ошибка получения чатов"}`, http.StatusInternalServerError)
			fmt.Println("Ошибка сохранения в базу данных: ", err)
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		err = json.NewEncoder(w).Encode(chats)
		if err != nil {
			fmt.Println("Ошибка кодирования джон ответа: ", err)
			return
		}
	}
}
func (sv *Server) GetUserMessages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var (
			err      error
			messages []md.Messages
		)
		chat_idParam := r.URL.Query().Get("chat_id")
		if chat_idParam == "" {
			http.Error(w, `{"error": "Отсутствует параметр chat_id"}`, http.StatusBadRequest)
			fmt.Println("Ошибка: chat_id не указан в запросе")
			return
		}
		chat_id, err := strconv.Atoi(chat_idParam)
		if err != nil {
			http.Error(w, `{"error": "Ошибка конвертации параметра"}`, http.StatusBadRequest)
			fmt.Println("Ошибка конвертации параметра")
			return
		}

		messages, err = sv.repo.GetChatMessages(int64(chat_id))
		if err != nil {
			http.Error(w, `{"error": "Ошибка получения истории"}`, http.StatusInternalServerError)
			fmt.Println("Ошибка сохранения в базу данных: ", err)
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		err = json.NewEncoder(w).Encode(messages)
		if err != nil {
			fmt.Println("Ошибка кодирования джон ответа: ", err)
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
			http.Error(w, `{"error": "Отсутствует параметр username"}`, http.StatusBadRequest)
			fmt.Println("Ошибка: username не указан в запросе")
			return
		}

		id, err = sv.repo.GetUserID(username)
		if err != nil {
			http.Error(w, `{"error": "Пользователь не найден"}`, http.StatusNotFound) // 404
			fmt.Println("Ошибка получения из базы данных: ", err)
			return
		}
		var resp struct {
			Id int64 `json:"user_id"`
		}
		resp.Id = id
		w.WriteHeader(http.StatusOK) // 200 — Всё Ок
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			fmt.Println("Ошибка кодирования джон ответа: ", err)
			return
		}
	}
}
