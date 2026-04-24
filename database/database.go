package database

import (
	"database/sql"
	md "discord/models"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type DBRepo interface {
	InitDB() error

	GetUser(id int64) (md.User, error)
	GetUserID(username string) (int64, error)
	GetUserChats(user_id int64) ([]md.Chat, error)
	GetChatMessages(chat_id int64) ([]md.Messages, error)

	SaveUser(username, password string) (int64, error)
	SaveChatWithAuthor(name string, user_id int64) (int64, error)
	SaveMessage(message md.Messages) error
	SaveUserToChat(user_id, chat_id int64) error
}

const SAVE_MESSAGE = "SaveMessage"
const GET_USER = "GetUser"
const GET_USER_ID = "GetUserId"
const SAVE_USER_TO_CHAT = "GetUserToChat"

type Service struct {
	db   *sql.DB
	Stmt map[string]*sql.Stmt
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("%s не найден, использую дефолт: %s", key, fallback)
	return fallback
}

func (sv *Service) InitDB() error {

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "192.168.31.234"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "discord_user"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "discord"),
		getEnv("SSLMODE", "disable"),
	)

	// 2. Открываем соединение
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	sv.db = db

	// 3. Проверяем соединение (sql.Open не создает само подключение сразу)
	err = db.Ping()
	if err != nil {
		return err
	}

	// создание таблиц
	if _, err = db.Exec(Users); err != nil {
		return err
	}
	if _, err = db.Exec(Chat); err != nil {
		return err
	}
	if _, err = db.Exec(Chat_members); err != nil {
		return err
	}
	if _, err = db.Exec(Messages); err != nil {
		return err
	}
	if err = sv.PrepareSQL(); err != nil {
		return err
	}
	fmt.Println("Успешное подключение к базе данных!")
	return nil
}
func (sv *Service) PrepareSQL() error {
	pattern := func(methodName, query string) error {
		stmt, err := sv.db.Prepare(query)
		if err != nil {
			return err
		}
		sv.Stmt[methodName] = stmt
		return nil
	}

	if err := pattern(SAVE_MESSAGE, "insert into messages (chat_id, sender_id, message) values($1, $2,$3)"); err != nil {
		return err
	}
	if err := pattern(GET_USER, "select username, reg from users where id = $1"); err != nil {
		return err
	}
	if err := pattern(GET_USER_ID, "select id from users where username = $1"); err != nil {
		return err
	}
	if err := pattern(SAVE_USER_TO_CHAT, "insert into chat_members (chat_id, user_id) values($1,$2)"); err != nil {
		return err
	}
	return nil
}

// --- получение
func (sv *Service) GetUserID(username string) (int64, error) {
	var id int64
	err := sv.Stmt[GET_USER_ID].QueryRow(username).Scan(&id)
	if err != nil {
		return id, err
	}
	return id, nil
}
func (sv *Service) GetUser(id int64) (md.User, error) {
	var user md.User
	err := sv.Stmt[GET_USER].QueryRow(id).Scan(&user.Name, &user.Reg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, fmt.Errorf("пользователь с id %d не найден", id)
		}
		return user, err
	}
	return user, nil
}

func (sv *Service) GetUserChats(user_id int64) ([]md.Chat, error) {
	var result []md.Chat
	rows, err := sv.db.Query(`
		SELECT c.id, c.name
		FROM chats c
		JOIN chat_members cm ON c.id = cm.chat_id
		WHERE cm.user_id = $1;`, user_id)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var c md.Chat
		if err := rows.Scan(&c.Id, &c.Name); err != nil {
			return result, err
		}
		result = append(result, c)
	}
	return result, nil
}
func (sv *Service) GetChatMessages(chat_id int64) ([]md.Messages, error) {
	var result []md.Messages
	rows, err := sv.db.Query(`select sender_id, message from messages where chat_id =$1 ORDER BY id ASC`, chat_id)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var m md.Messages
		if err := rows.Scan(&m.SenderId, &m.Message); err != nil {
			return result, err
		}
		result = append(result, m)
	}
	return result, nil
}

// --- сохранение
func (sv *Service) SaveUser(username, password string) (int64, error) {
	user := md.User{
		Password: password,
		Reg:      time.Now(),
	}
	err := sv.db.QueryRow("insert into users (username, password,reg) values($1,$2,$3) RETURNING id",
		username, user.Password, user.Reg).Scan(&user.Id)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return 0, fmt.Errorf("пользователь с таким именем уже существует")
			}
		}
		return 0, err
	}
	return user.Id, nil
}
func (sv *Service) SaveChatWithAuthor(name string, user_id int64) (int64, error) {
	tx, err := sv.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var chat_id int64
	err = tx.QueryRow("insert into chats (name) values($1) RETURNING id", name).Scan(&chat_id)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES($1, $2)", chat_id, user_id)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return chat_id, nil
}
func (sv *Service) SaveMessage(message md.Messages) error {

	_, err := sv.Stmt[SAVE_MESSAGE].Exec(message.ChatId, message.SenderId, message.Message)
	if err != nil {
		return err
	}
	return nil
}
func (sv *Service) SaveUserToChat(user_id, chat_id int64) error {
	_, err := sv.Stmt[SAVE_USER_TO_CHAT].Exec(chat_id, user_id)
	if err != nil {
		return err
	}
	return err
}
