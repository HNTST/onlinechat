package server

import (
	db "discord/database"
	md "discord/models"
	"encoding/json"
	"fmt"
	"net/http"
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
			ChatName string `json:"chat_name"`
			Author   int64  `json:"author"`
		}
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest) // 400
			fmt.Println("Ошибка декодирования джон запроса: ", err)
			return
		}
		id, err := sv.repo.SaveChatWithAuthor(input.ChatName, input.Author)
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
		var input struct {
			ID int64 `json:"user_id"`
		}
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, `{"error": "Некорректный ID"}`, http.StatusBadRequest)
			fmt.Println("Ошибка декодирования джон запроса: ", err)
			return
		}

		user, err = sv.repo.GetUser(input.ID)
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
		var input struct {
			UserID int `json:"user_id"`
		}
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, `{"error": "Некорректный UserID"}`, http.StatusBadRequest)
			fmt.Println("Ошибка декодирования джон запроса: ", err)
			return
		}
		chats, err = sv.repo.GetUserChats(int64(input.UserID))
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
		var input struct {
			ChatId int `json:"chat_id"`
		}
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, `{"error": "Некорректный ChatID"}`, http.StatusBadRequest)
			fmt.Println("Ошибка декодирования джон запроса: ", err)
			return
		}
		messages, err = sv.repo.GetChatMessages(int64(input.ChatId))
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
		var input struct {
			Username string `json:"username"`
		}
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, `{"error": "Некорректный ID"}`, http.StatusBadRequest)
			fmt.Println("Ошибка декодирования джон запроса: ", err)
			return
		}
		id, err = sv.repo.GetUserID(input.Username)
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
