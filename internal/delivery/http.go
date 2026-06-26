package delivery

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	core_http_server "discord/internal/core/transport/server"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type OLD_SERVER_HANDLERS struct {
	service   Service
	clients   map[int64]*websocket.Conn
	clientsMu sync.Mutex
	logger    *core_logger.Logger
}
type Service interface {
	GetUserID(ctx context.Context, username string) (int64, error)
	SaveUser(ctx context.Context, username, password string) (int64, error)
	SaveChat(ctx context.Context, chat_name string, user_id, target_user_id int64) (int64, error)
	GetUser(ctx context.Context, user_id int64) (domain.User, error)
	GetUserChats(ctx context.Context, user_id int64) ([]domain.Chat, error)
	GetUserMessages(ctx context.Context, chat_id int64) ([]domain.Messages, error)
	GetChatMembers(ctx context.Context, chat_id int64) ([]int64, error)
	SaveMessage(ctx context.Context, message domain.Messages) error
	UploadUserAvatar(ctx context.Context, user_id int64, file io.Reader, orig_file_name string, size int64, content_type string) (int64, error)
	AddUserToChat(ctx context.Context, chat_id, target_user_id int64) error
}

func NewServer(logger *core_logger.Logger, s Service) *OLD_SERVER_HANDLERS {
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
			Method:  http.MethodGet,
			Path:    "/ws",
			Handler: h.HandleConnections,
		},
		{
			Method:  http.MethodPost,
			Path:    "/chats/add_user",
			Handler: h.AddUserToChat,
		},
	}
}

// SaveUser creates a new user
//
//	@Summary		Регистрация пользователя
//	@Description	Создание новой учётной записи пользователя
//	@Tags			v1 / Пользователи
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{username=string,password=string}	true	"Данные для регистрации"
//	@Success		201		{object}	object{user_id=int64}
//	@Failure		400		{object}	object{}
//	@Router			/api/v1/user [post]
func (sv *OLD_SERVER_HANDLERS) SaveUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)
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
		responseHandler.ErrorResponse(err, "failed to decode json")
		return
	}
	id, err := sv.service.SaveUser(ctx, input.Username, input.Password)
	if err != nil {
		if errors.Is(err, core_errors.ErrUserAlreadyExists) {
			responseHandler.ErrorResponse(err, "user already exists")
			return
		}
		responseHandler.ErrorResponse(err, "failed to save user")
		return
	}

	var resp struct {
		Id int64 `json:"user_id"`
	}
	resp.Id = id
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to encode json")
		return
	}
}

// SaveChat creates a new chat
//
//	@Summary		Создание чата
//	@Description	Создание нового чата с двумя участниками
//	@Tags			v1 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{chat_name=string,author=int64,target_user=int64}	true	"Данные чата"
//	@Success		201		{object}	object{chat_id=int64}
//	@Failure		400		{object}	object{}
//	@Router			/api/v1/chat [post]
func (sv *OLD_SERVER_HANDLERS) SaveChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)
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
		responseHandler.ErrorResponse(err, "failed to decode json")
		return
	}

	id, err := sv.service.SaveChat(ctx, input.ChatName, input.Author, input.TargetUser)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to save chat")
		return
	}
	var resp struct {
		Id int64 `json:"chat_id"`
	}
	resp.Id = id
	w.WriteHeader(http.StatusCreated) // 201
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to encode json")
		return
	}
}

// UpdateUserAvatar uploads an avatar image for a user
//
//	@Summary		Загрузка аватара
//	@Description	Загрузка файла аватара пользователя в MinIO и сохранение метаданных
//	@Tags			v1 / Файлы
//	@Accept			mpfd
//	@Produce		json
//	@Param			user_id			query		int		true	"ID пользователя"
//	@Param			new_user_avatar	formData	file	true	"Файл изображения аватара"
//	@Success		201				{object}	object{file_id=int64}
//	@Failure		400				{object}	object{}
//	@Router			/api/v1/user/avatar [post]
func (sv *OLD_SERVER_HANDLERS) UpdateUserAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)
	w.Header().Set("Content-Type", "application/json")

	user_idParam := r.URL.Query().Get("user_id")
	if user_idParam == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}
	user_id, err := strconv.Atoi(user_idParam)
	if err != nil {
		responseHandler.ErrorResponse(err, "param is incorrect")
		return
	}

	file, header, err := r.FormFile("new_user_avatar")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get avatar")
		return
	}
	defer file.Close()

	file_id, err := sv.service.UploadUserAvatar(ctx, int64(user_id), file, header.Filename, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to save user avatar")
		return
	}
	var resp struct {
		Id int64 `json:"file_id"`
	}
	resp.Id = file_id
	w.WriteHeader(http.StatusCreated) // 201
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to encode json")
		return
	}
}

// AddUserToChat adds a user to an existing chat
//
//	@Summary		Добавление пользователя в чат
//	@Description	Добавление пользователя в существующий чат по ID чата
//	@Tags			v1 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{chat_id=int64,target_user_id=int64}	true	"Данные для добавления"
//	@Success		200		{object}	object{}
//	@Failure		400		{object}	object{}
//	@Router			/api/v1/chats/add_user [post]
func (sv *OLD_SERVER_HANDLERS) AddUserToChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	var input struct {
		Chat_id     int64 `json:"chat_id"`
		Target_user int64 `json:"target_user_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to decode json")
		return
	}
	err = sv.service.AddUserToChat(ctx, input.Chat_id, input.Target_user)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to add user")
		return
	}
	w.WriteHeader(http.StatusOK) // 200 — Всё Ок
}

// GetUser returns a user by ID
//
//	@Summary		Получение пользователя
//	@Description	Получение данных пользователя по его ID
//	@Tags			v1 / Пользователи
//	@Accept			json
//	@Produce		json
//	@Param			user_id	query		int	true	"ID пользователя"
//	@Success		200		{object}	domain.User
//	@Failure		400		{object}	object{}
//	@Router			/api/v1/user [get]
func (sv *OLD_SERVER_HANDLERS) GetUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	log.Warn("Hello from GetUser HTTP")
	w.Header().Set("Content-Type", "application/json")
	var (
		err  error
		user domain.User
	)
	user_idParam := r.URL.Query().Get("user_id")
	if user_idParam == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}
	user_id, err := strconv.Atoi(user_idParam)
	if err != nil {
		responseHandler.ErrorResponse(err, "param is incorrect")
		return
	}

	user, err = sv.service.GetUser(ctx, int64(user_id))
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user")
		return
	}
	w.WriteHeader(http.StatusOK) // 200 — Всё Ок
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		sv.sendError(w, core_errors.ErrEncodingRequest.Error(), http.StatusInternalServerError, err)
		return
	}
}

// GetUserChats returns all chats for a user
//
//	@Summary		Чаты пользователя
//	@Description	Получение всех чатов, в которых состоит пользователь
//	@Tags			v1 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			user_id	query		int	true	"ID пользователя"
//	@Success		200		{array}		domain.Chat
//	@Failure		400		{object}	object{}
//	@Router			/api/v1/chats [get]
func (sv *OLD_SERVER_HANDLERS) GetUserChats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)
	var (
		err   error
		chats []domain.Chat
	)
	user_idParam := r.URL.Query().Get("user_id")
	if user_idParam == "" {
		responseHandler.ErrorResponse(core_errors.ErrParamMissing, "param is incorrect")
		return
	}
	user_id, err := strconv.Atoi(user_idParam)
	if err != nil {
		responseHandler.ErrorResponse(err, "param is incorrect")
		return
	}
	chats, err = sv.service.GetUserChats(ctx, int64(user_id))
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user's chats")
		return
	}
	w.WriteHeader(http.StatusOK) // 200
	err = json.NewEncoder(w).Encode(chats)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to encode json")
		return
	}
}

// GetUserMessages returns messages for a chat
//
//	@Summary		Сообщения чата
//	@Description	Получение всех сообщений в чате
//	@Tags			v1 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			chat_id	query		int	true	"ID чата"
//	@Success		200		{array}		domain.Messages
//	@Failure		400		{object}	object{}
//	@Router			/api/v1/messages [get]
func (sv *OLD_SERVER_HANDLERS) GetUserMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)
	w.Header().Set("Content-Type", "application/json")
	var (
		err      error
		messages []domain.Messages
	)
	chat_idParam := r.URL.Query().Get("chat_id")
	if chat_idParam == "" {
		responseHandler.ErrorResponse(core_errors.ErrParamMissing, "param is incorrect")
		return
	}
	chat_id, err := strconv.Atoi(chat_idParam)
	if err != nil {
		responseHandler.ErrorResponse(err, "param is incorrect")
		return
	}

	messages, err = sv.service.GetUserMessages(ctx, int64(chat_id))
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user's messages")
		return
	}
	w.WriteHeader(http.StatusOK) // 200
	err = json.NewEncoder(w).Encode(messages)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to encode json")
		return
	}
}

// GetUserID returns a user ID by username
//
//	@Summary		Получение ID пользователя
//	@Description	Получение ID пользователя по его имени
//	@Tags			v1 / Пользователи
//	@Accept			json
//	@Produce		json
//	@Param			username	query		string	true	"Имя пользователя"
//	@Success		200			{object}	object{user_id=int64}
//	@Failure		400			{object}	object{}
//	@Router			/api/v1/user_id [get]
func (sv *OLD_SERVER_HANDLERS) GetUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)
	w.Header().Set("Content-Type", "application/json")
	var (
		err error
		id  int64
	)
	username := r.URL.Query().Get("username")
	if username == "" {
		responseHandler.ErrorResponse(core_errors.ErrParamMissing, "param is incorrect")
		return
	}

	id, err = sv.service.GetUserID(ctx, username)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user_id")
		return
	}
	var resp struct {
		Id int64 `json:"user_id"`
	}
	resp.Id = id
	w.WriteHeader(http.StatusOK) // 200 — Всё Ок
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to encode json")
		return
	}
}

// GetChatMembers returns member IDs for a chat
//
//	@Summary		Участники чата
//	@Description	Получение ID всех участников чата
//	@Tags			v1 / Чаты
//	@Accept			json
//	@Produce		json
//	@Param			chat_id	query		int	true	"ID чата"
//	@Success		200		{object}	object{members=[]int64}
//	@Failure		400		{object}	object{}
//	@Router			/api/v1/chats/members [get]
func (sv *OLD_SERVER_HANDLERS) GetChatMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)
	w.Header().Set("Content-Type", "application/json")
	var (
		err error
		id  []int64
	)

	chat_idParam := r.URL.Query().Get("chat_id")
	if chat_idParam == "" {
		responseHandler.ErrorResponse(core_errors.ErrParamMissing, "param is incorrect")
		return
	}
	chat_id, err := strconv.Atoi(chat_idParam)
	if err != nil {
		responseHandler.ErrorResponse(err, "param is incorrect")
		return
	}

	id, err = sv.service.GetChatMembers(ctx, int64(chat_id))
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get chat members")
		return
	}
	var resp struct {
		Id []int64 `json:"members"`
	}
	resp.Id = id
	w.WriteHeader(http.StatusOK) // 200 — Всё Ок
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to encode json")
		return
	}
}
