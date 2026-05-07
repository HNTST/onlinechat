package postgres

import (
	"context"
	"database/sql"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"errors"
	"fmt"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Драйвер для PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const SAVE_MESSAGE = "SaveMessage"
const GET_USER = "GetUser"
const GET_USER_ID = "GetUserId"
const SAVE_USER_TO_CHAT = "GetUserToChat"

// type Database struct {
// 	cfg  config.Config_DB
// 	db   *sql.DB
// 	Stmt map[string]*sql.Stmt
// }

// func NewDatabse(cfg config.Config_DB) *Database {
// 	return &Database{cfg: cfg,
// 		Stmt: make(map[string]*sql.Stmt)}
// }
// func (sv *Database) InitDB() error {

// 	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
// 		sv.cfg.Host, sv.cfg.Port, sv.cfg.User, sv.cfg.Password, sv.cfg.Name, sv.cfg.Sslmode,
// 	)

// 	// 2. Открываем соединение
// 	db, err := sql.Open("postgres", connStr)
// 	if err != nil {
// 		return err
// 	}
// 	sv.db = db

// 	// 3. Проверяем соединение (sql.Open не создает само подключение сразу)
// 	err = db.Ping()
// 	if err != nil {
// 		return err
// 	}

// 	// создание таблиц
// 	// if _, err = db.Exec(Users); err != nil {
// 	// 	return err
// 	// }
// 	// if _, err = db.Exec(Chat); err != nil {
// 	// 	return err
// 	// }
// 	// if _, err = db.Exec(Chat_members); err != nil {
// 	// 	return err
// 	// }
// 	// if _, err = db.Exec(Messages); err != nil {
// 	// 	return err
// 	// }

// 	// МИГРАЦИИ
// 	migrationConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
// 		sv.cfg.User,
// 		sv.cfg.Password,
// 		sv.cfg.Host,
// 		sv.cfg.Port,
// 		sv.cfg.Name,
// 		sv.cfg.Sslmode,
// 	)
// 	m, err := migrate.New(
// 		"file://migrations",
// 		migrationConnStr,
// 	)

// 	if err != nil {
// 		return err
// 	}

// 	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
// 		return err
// 	}

// 	if err = sv.PrepareSQL(); err != nil {
// 		return err
// 	}
// 	return nil
// }
// func (sv *Database) Close() error {
// 	return sv.db.Close()
// }
// func (sv *Database) PrepareSQL() error {
// 	pattern := func(methodName, query string) error {
// 		stmt, err := sv.db.Prepare(query)
// 		if err != nil {
// 			return err
// 		}
// 		sv.Stmt[methodName] = stmt
// 		return nil
// 	}

// 	if err := pattern(SAVE_MESSAGE, "insert into messages (chat_id, sender_id, message, sent_at) values($1, $2,$3, $4)"); err != nil {
// 		return err
// 	}
// 	if err := pattern(GET_USER, "select username, reg from users where id = $1"); err != nil {
// 		return err
// 	}
// 	if err := pattern(GET_USER_ID, "select id from users where username = $1"); err != nil {
// 		return err
// 	}
// 	if err := pattern(SAVE_USER_TO_CHAT, "insert into chat_members (chat_id, user_id) values($1,$2)"); err != nil {
// 		return err
// 	}
// 	return nil
// }

// --- получение
func (sv *OldDbREpository) GetUserID(ctx context.Context, username string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, sv.pool.OpTimeout())
	defer cancel()

	query := `
		select id
		from users
		where username=$1
	`
	row := sv.pool.QueryRow(ctx, query, username)
	var id int64
	err := row.Scan(
		&id,
	)
	if err != nil {
		return id, fmt.Errorf("scan error: %w", err)
	}
	return id, nil
}
func (sv *OldDbREpository) GetUser(ctx context.Context, id int64) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, sv.pool.OpTimeout())
	defer cancel()

	query := `
		select username, reg
		from users
		where id = $1
	`

	row := sv.pool.QueryRow(ctx, query, id)
	var user domain.User

	err := row.Scan(
		&user.Name,
		&user.Reg,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, fmt.Errorf("пользователь с id %d не найден", id)
		}
		return user, fmt.Errorf("scan error: %w", err)
	}
	return user, nil
}
func (sv *OldDbREpository) GetUserChats(ctx context.Context, user_id int64) ([]domain.Chat, error) {
	ctx, cancel := context.WithTimeout(ctx, sv.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT c.id, c.name
		FROM chats c
		JOIN chat_members cm ON c.id = cm.chat_id
		WHERE cm.user_id = $1;
		`
	var result []domain.Chat
	rows, err := sv.pool.Query(ctx, query, user_id)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var c domain.Chat
		if err := rows.Scan(&c.Id, &c.Name); err != nil {
			return result, fmt.Errorf("scan error: %w", err)
		}
		result = append(result, c)
	}
	return result, nil
}
func (sv *OldDbREpository) GetChatMessages(ctx context.Context, chat_id int64) ([]domain.Messages, error) {
	ctx, cancel := context.WithTimeout(ctx, sv.pool.OpTimeout())
	defer cancel()

	query := `
		select sender_id, message, sent_at
		from messages
		where chat_id =$1
		ORDER BY id ASC
	`
	var result []domain.Messages
	rows, err := sv.pool.Query(ctx, query, chat_id)
	if err != nil {
		return result, fmt.Errorf("get rows error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m domain.Messages
		if err := rows.Scan(&m.SenderId, &m.Message, &m.Sent_at); err != nil {
			return result, fmt.Errorf("scan error: %w", err)
		}
		result = append(result, m)
	}
	return result, nil
}
func (sv *OldDbREpository) GetChatMembers(ctx context.Context, chat_id int64) ([]int64, error) {
	ctx, cancel := context.WithTimeout(ctx, sv.pool.OpTimeout())
	defer cancel()
	query := `
		select user_id
		from chat_members
		where chat_id =$1
	`
	var result []int64
	rows, err := sv.pool.Query(ctx, query, chat_id)
	if err != nil {
		return result, fmt.Errorf("get rows error: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return result, fmt.Errorf("scan error: %w", err)
		}
		result = append(result, id)
	}
	return result, nil
}
func (sv *OldDbREpository) FindExistingChat(ctx context.Context, user1, user2 int64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, sv.pool.OpTimeout())
	defer cancel()

	query := `
        SELECT chat_id
        FROM chat_members
        WHERE user_id IN ($1, $2)
        GROUP BY chat_id
        HAVING COUNT(DISTINCT user_id) = 2
        `

	var chatID int64

	row := sv.pool.QueryRow(ctx, query, user1, user2)
	err := row.Scan(&chatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("get rows error: %w", err)
		}
		return 0, fmt.Errorf("get rows error: %w", err)
	}

	return chatID, nil
}

// --- сохранение
func (sv *OldDbREpository) SaveUser(ctx context.Context, username, password string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, sv.pool.OpTimeout())
	defer cancel()

	user := domain.User{
		Password: password,
		Reg:      time.Now(),
	}

	query := `
		insert into users (username, password,reg)
		values($1,$2,$3)
		RETURNING id
	`
	row := sv.pool.QueryRow(ctx, query, username, user.Password, user.Reg)
	err := row.Scan(&user.Id)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return 0, fmt.Errorf("scan error: %w", core_errors.ErrUserAlreadyExists)
			}
		}
		return 0, fmt.Errorf("scan error: %w", err)
	}
	return user.Id, nil
}

func (sv *OldDbREpository) SaveChatWithAuthor(
	ctx context.Context,
	name string,
	user_id, target_user_id int64,
) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, sv.pool.OpTimeout())
	defer cancel()
	// ШАГ 0: Проверяем, а не общаются ли они уже?
	existingID, err := sv.FindExistingChat(ctx, user_id, target_user_id)
	if err != nil {
		return 0, err
	}
	if existingID > 0 {
		return existingID, nil // Возвращаем старый чат, не создавая новый
	}

	// Если старого чата нет — погнали создавать новый
	tx, err := sv.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var chat_id int64
	err = tx.QueryRow(ctx, "insert into chats (name) values($1) RETURNING id", name).Scan(&chat_id)
	if err != nil {
		return 0, err
	}

	// Добавляем первого
	_, err = tx.Exec(ctx, "INSERT INTO chat_members (chat_id, user_id) VALUES($1, $2)", chat_id, user_id)
	if err != nil {
		return 0, err
	}

	// Добавляем второго
	_, err = tx.Exec(ctx, "INSERT INTO chat_members (chat_id, user_id) VALUES($1, $2)", chat_id, target_user_id)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return chat_id, nil
}
func (sv *OldDbREpository) SaveMessage(ctx context.Context, message domain.Messages) error {
	ctx, cancel := context.WithTimeout(ctx, sv.pool.OpTimeout())
	defer cancel()

	_, err := sv.pool.Exec(ctx, "insert into messages (chat_id, sender_id, message, sent_at) values($1, $2,$3, $4)", message.ChatId, message.SenderId, message.Message, message.Sent_at)
	if err != nil {
		return err
	}
	return nil
}
func (sv *OldDbREpository) SaveUserToChat(ctx context.Context, user_id, chat_id int64) error {
	ctx, cancel := context.WithTimeout(ctx, sv.pool.OpTimeout())
	defer cancel()
	_, err := sv.pool.Exec(ctx, "insert into chat_members (chat_id, user_id) values($1,$2)", chat_id, user_id)
	if err != nil {
		return err
	}
	return nil
}
func (sv *OldDbREpository) UpdateUserAvatarPath(ctx context.Context, user_id, size int64, file_key, orig_file_name, type_file string) (int64, error) {

	tx, err := sv.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	var file_id int64
	err = tx.QueryRow(ctx, "insert into files (file_key, original_file_name, type_file, size_file) values ($1,$2,$3,$4) RETURNING ID",
		file_key, orig_file_name, type_file, size).Scan(&file_id)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(ctx, "update users set avatar_id = $1 where id = $2", file_id, user_id)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return file_id, nil
}
