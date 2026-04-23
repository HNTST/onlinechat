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
	SaveChat(name string) (int64, error)
	SaveMessage(message md.Messages) error
	SaveUserToChat(user_id, chat_id int64) error
}

type Service struct {
	db *sql.DB
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
	fmt.Println("Успешное подключение к базе данных!")
	return nil
}

// --- получение
func (sv *Service) GetUser(id int64) (md.User, error) {
	var user md.User
	err := sv.db.QueryRow("select username, reg from users where id = $1", id).Scan(&user.Name, &user.Reg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, fmt.Errorf("пользователь с id %d не найден", id)
		}
		// Если произошла какая-то другая ошибка (проблемы с БД, сетью и т.д.)
		return user, err
	}
	return user, nil
}
func (sv *Service) GetUserID(username string) (int64, error) {
	var id int64
	err := sv.db.QueryRow("select id from users where username = $1", username).Scan(&id)
	if err != nil {
		return id, err
	}
	return id, nil
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
		// Приводим ошибку к типу *pq.Error
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return 0, fmt.Errorf("пользователь с таким именем уже существует")
			}
		}
		return 0, err
	}
	return user.Id, nil
}
func (sv *Service) SaveChat(name string) (int64, error) {
	var id int64
	err := sv.db.QueryRow("insert into chats (name) values($1) RETURNING id",
		name).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func (sv *Service) SaveMessage(message md.Messages) error {
	_, err := sv.db.Exec("insert into messages (chat_id, sender_id, message) values($1, $2,$3)",
		message.ChatId, message.SenderId, message.Message)
	if err != nil {
		return err
	}
	return nil
}
func (sv *Service) SaveUserToChat(user_id, chat_id int64) error {
	_, err := sv.db.Exec("insert into chat_members (chat_id, user_id) values($1,$2)", chat_id, user_id)
	if err != nil {
		return err
	}
	return err
}
