# Технический паспорт проекта

Backend мессенджера в стиле Discord. Монолит на Go с REST API + WebSocket.

---

## Архитектура

Проект использует **V2 feature-based архитектуру** в `internal/features/`. Каждый модуль:

```
module/
├── transport/http/     # HTTP-хендлеры, DTO, валидация
├── transport/ws/       # WebSocket-хендлеры (только chat)
├── service/            # Бизнес-логика
└── repository/postgres/# SQL-запросы через sqlc
```

Инверсия зависимостей: интерфейсы сервиса объявляются в слое транспорта, интерфейсы репозитория — в слое сервиса.

### Ядро (`internal/core/`)
Общая инфраструктура: конфигурации, доменные модели, пул соединений (`GetQueries(ctx)` для sqlc, `GetTx(ctx)` для транзакций), логер, middleware, транспортная утилита.

### Сгенерированный код (`internal/db/`)
Пакет `db` — автоматически сгенерированный sqlc. Содержит типизированные модели и методы запросов. **Не редактировать вручную.**

### SQL-запросы (`query/`)
Исходные SQL-запросы для sqlc, organised по фичам:

```
query/
├── auth/sessions.sql
├── chat/
│   ├── chats.sql
│   ├── chat_members.sql
│   └── messages.sql
├── user/users.sql
└── files/files.sql
```

---

## Стек технологий

| Компонент | Технология |
|-----------|-----------|
| Язык | Go 1.26.2 |
| Маршрутизация | `net/http.ServeMux` |
| БД | PostgreSQL 15 (pgx/v5 pgxpool, MaxConns=20) |
| SQL-генерация | sqlc v1.31.1 (pgx/v5 driver) |
| Миграции | `golang-migrate/v4` (auto-run при старте) |
| Объектное хранилище | MinIO (minio-go/v7) |
| WebSocket | `gorilla/websocket` |
| Валидация | `go-playground/validator/v10` |
| Логирование | `log/slog` (кастомный обёртка) |
| Тестирование | `testify` (assert, mock, suite) |
| Нагрузочные тесты | k6 |
| Документация | Swagger (swaggo) |

---

## Структура БД

### Таблица `files`
| Колонка | Тип | Ограничения |
|---------|-----|-------------|
| id | BIGSERIAL | PK |
| file_key | VARCHAR(255) | UNIQUE NOT NULL |
| original_file_name | VARCHAR(255) | NOT NULL |
| type_file | VARCHAR(255) | NOT NULL |
| size_file | BIGINT | NOT NULL |

### Таблица `users`
| Колонка | Тип | Ограничения |
|---------|-----|-------------|
| id | BIGSERIAL | PK |
| username | VARCHAR(255) | UNIQUE NOT NULL |
| password | VARCHAR(255) | NOT NULL |
| reg | TIMESTAMPTZ | NOT NULL DEFAULT CURRENT_TIMESTAMP |
| avatar_id | BIGINT | FK → files(id) ON DELETE SET NULL, DEFAULT 1 |
| last_seen_at | TIMESTAMPTZ | DEFAULT NULL |

### Таблица `chats`
| Колонка | Тип | Ограничения |
|---------|-----|-------------|
| id | BIGSERIAL | PK |
| name | VARCHAR(255) | NOT NULL |
| is_group | BOOLEAN | DEFAULT false |

### Таблица `chat_members`
| Колонка | Тип | Ограничения |
|---------|-----|-------------|
| chat_id | BIGINT | FK → chats(id) ON DELETE CASCADE |
| user_id | BIGINT | FK → users(id) |
| | | PK(chat_id, user_id) |

### Таблица `messages`
| Колонка | Тип | Ограничения |
|---------|-----|-------------|
| id | BIGSERIAL | PK |
| chat_id | BIGINT | NOT NULL FK → chats(id) ON DELETE CASCADE |
| sender_id | BIGINT | NOT NULL FK → users(id) ON DELETE CASCADE |
| message | TEXT | NOT NULL |
| sent_at | TIMESTAMPTZ | NOT NULL DEFAULT CURRENT_TIMESTAMP |
| is_edited | BOOLEAN | DEFAULT false |
| is_deleted | BOOLEAN | DEFAULT false |
| is_read | BOOLEAN | DEFAULT false |

### Таблица `sessions`
| Колонка | Тип | Ограничения |
|---------|-----|-------------|
| id | BIGSERIAL | PK |
| user_id | BIGINT | FK → users(id) ON DELETE CASCADE |
| token | VARCHAR(255) | NOT NULL, UNIQUE |
| expires_at | TIMESTAMPTZ | NOT NULL |

---

## Модули

### Auth (`internal/features/auth/`)
Регистрация, логин, логаут. bcrypt для паролей. Сессии — 64 hex-символа, TTL 24ч. Cookie `session_token` (HttpOnly).

### User (`internal/features/user/`)
Профиль пользователя, поиск по ID/username. `last_seen_at` для определения online/offline статуса фронтендом.

### Chat (`internal/features/chat/`)
CRUD чатов и сообщений. Личные и групповые чаты. Курсорная пагинация. Мягкое удаление. Отметка о прочтении. WebSocket для real-time доставки + online/offline статус.

### Files (`internal/features/files/`)
Загрузка файлов в MinIO + метаданные в PostgreSQL. Валидация MIME (JPEG/PNG/WebP, макс. 5MB).

---

## sqlc — интеграция

### Конфигурация
Файл `sqlc.yaml` в корне проекта. Генерирует Go-код в `internal/db/` (package `db`) из SQL-запросов в `query/`.

### Паттерн использования в репозиториях
```go
func (r *ChatRepo) SaveMessage(ctx context.Context, message domain.Messages) error {
    q := r.pool.GetQueries(ctx)  // ← берём *db.Queries из pool
    return q.InsertMessage(ctx, db.InsertMessageParams{
        ChatID:   message.ChatId,
        SenderID: message.SenderId,
        Message:  message.Message,
        SentAt:   message.Sent_at,
    })
}
```

### Транзакции
`ConnectionPool.GetQueries(ctx)` проверяет наличие `pgx.Tx` в контексте. Если транзакция активна — возвращает `q.WithTx(tx)`, иначе `q` (привязан к pgxpool). Транзакции по-прежнему открываются через `repo.Transaction(ctx, fn)`.

### Типы моделей
sqlc генерирует модели в `internal/db/models.go`. Маппинг на `domain.*` модели происходит в репозиторном слое. Nullable-поля (`avatar_id`, `last_seen_at`) конвертируются из `pgtype.Int8`/`pgtype.Timestamptz` в `*int64`/`*time.Time`.

### Обновление SQL
После изменения SQL-файлов в `query/` необходимо выполнить:
```bash
/root/go/bin/sqlc generate
```

---

## Доменные ошибки

| Ошибка | Описание | HTTP |
|--------|----------|------|
| `ErrUserAlreadyExists` | Пользователь уже существует | 409 |
| `ErrInternal` | Внутренняя ошибка сервера (маскируется) | 500 |
| `ErrInvalidJSON` | Неверный формат JSON | 400 |
| `ErrParamMissing` | Отсутствует обязательный параметр | 400 |
| `ErrEncodingRequest` | Ошибка кодирования JSON | 500 |
| `ErrUserNotFound` | Пользователь не найден | 404 |
| `ErrInvalidArgument` | Неверный аргумент | 400 |
| `ErrConfict` | Конфликт | 409 |
| `ErrChatNotExists` | Чат не существует | 404 |
| `ErrMessageNotFound` | Сообщение не найдено | 404 |
| `ErrUnauthorized` | Не авторизован | 401 |

---

## API V2 — Карта маршрутов

### Auth
| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v2/auth/register` | Регистрация + авто-логин. Body: `{"username","password"}`. Устанавливает cookie. |
| `POST` | `/api/v2/auth/login` | Логин. Body: `{"username","password"}`. Устанавливает cookie. |
| `POST` | `/api/v2/auth/logout` | Логаут. Удаляет сессию, очищает cookie. |

### User
| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/api/v2/user/profile` | Свой профиль (из контекста сессии). |
| `GET` | `/api/v2/users?id=N` | Профиль пользователя по ID. |
| `GET` | `/api/v2/user_id?username=X` | Поиск ID по username. |
| `POST` | `/api/v2/user/avatar` | Загрузка аватарки (multipart). |
| `GET` | `/api/v2/user/avatar?id=N` | Стриминг изображения. |

### Chat
| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v2/chat` | Создать чат. Body: `{"chat_name","author","target_user","is_group"}`. |
| `GET` | `/api/v2/chats` | Список чатов (с последним сообщением + unread_count). |
| `PUT` | `/api/v2/chat/{chat_id}/name` | Переименовать групповой чат. |
| `DELETE` | `/api/v2/chat/{chat_id}` | Удалить групповой чат (CASCADE). |
| `POST` | `/api/v2/chat/{chat_id}/members` | Добавить участников. Body: `{"user_ids":[...]}`. |
| `GET` | `/api/v2/chats/{chat_id}/members` | Список участников чата. |
| `POST` | `/api/v2/chats/{chat_id}/messages` | Отправить сообщение. Body: `{"sender_id","message"}`. Возвращает 201 без body. |
| `GET` | `/api/v2/chats/{chat_id}/messages` | История (курсорная пагинация: `last_message_id`, `limit`). |
| `PUT` | `/api/v2/chats/{chat_id}/messages/{message_id}` | Редактировать. Body: `{"text"}`. |
| `DELETE` | `/api/v2/chats/{chat_id}/messages/{message_id}` | Мягкое удаление. |
| `PATCH` | `/api/v2/chats/{chat_id}/messages/read-status` | Отметить как прочитанное. Body: `{"last_read_message_id"}`. |

### WebSocket
| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/api/v2/ws` | WebSocket (обмен сообщениями + online/offline статус). |

### Swagger
| Путь | Описание |
|------|----------|
| `/swagger/*` | Swagger UI (без auth). |

---

## WebSocket — Протокол

### Подключение
`GET /api/v2/ws` (требует cookie `session_token`). Upgrade → 101.

### Входящее (от клиента)
```json
{ "message": "text", "chat_id": 1, "sender_id": 123 }
```

### Исходящее (от сервера)

**Сообщение в чате:**
```json
{ "id": 456, "message": "text", "chat_id": 1, "sender_id": 123, "sent_at": "...", "is_edited": false, "is_deleted": false, "is_read": false }
```

**Пользователь онлайн:**
```json
{ "type": "online", "user_id": 123 }
```

**Пользователь оффлайн:**
```json
{ "type": "offline", "user_id": 123 }
```

### Поведение online/offline
- При подключении: сервер уведомляет чат-партнёров о новом пользователе + отправляет новому пользователю список уже подключённых партнёров.
- При отключении: сервер уведомляет чат-партнёров об отключении.
- Уведомления приходят только для пользователей, состоящих в одном чате.

### Heartbeat (Ping/Pong)
- Сервер отправляет Ping каждые 54с (`pingPeriod`).
- Клиент автоматически отвечает Pong (встроено в браузерные WS-API и gorilla/websocket).
- `ReadDeadline` = 60с (`pongWait`) — сбрасывается при получении любого пакета (включая Pong).
- Если Pong не получен за 60с → `ReadMessage()` возвращает ошибку → соединение закрывается, клиент удаляется из `clients`, партнёры получают offline.
- Все записи в `conn` защищены `writeMu` (ping goroutine + notifyPartners + processMessage).

---

## Middleware (цепочка справа налево)

`Cors` → `RequestID` → `Logger` → `Panic` → `Trace` → `Auth`

### Auth Middleware
Проверяет cookie `session_token`, извлекает `user_id` в контекст (`core_http_middleware.UserIDKey`).

**Исключения (без сессии):** `/swagger/`, `/api/v2/auth/login`, `/api/v2/auth/register`.

---

## Конфигурация

| Файл | Назначение |
|------|-----------|
| `internal/core/config/.env` | Продакшн |
| `internal/core/config/.env.test` | Тесты (localhost:5432, localhost:9000) |
| `internal/core/config/.env.example` | Шаблон |

**Ключи:** `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` (required), `DB_NAME`, `DB_SSLMODE`, `MINIO_HOST`, `MINIO_BUCKET_NAME`, `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD` (required), `MINIO_USE_SSL`, `HTTP_ADDR`, `LOGGER_LEVEL`.

**Загрузка:** `APP_ENV=test` → `.env.test`, иначе → `.env`.

---

## Тестирование

### Три типа тестов

1. **Unit-тесты сервисов** (`tests/features/*/service/`) — моки, без внешних зависимостей. Входят в `make check`.
2. **Интеграционные HTTP-тесты** (`tests/features/*/transport/http/`) — `httptest.Server`, подмоканные сервисы, реальная цепочка middleware. Входят в `make check`.
3. **Интеграционные тесты репозиториев** (`tests/features/*/repository/postgres/`) — требуют PostgreSQL. Упакованы в `testify/suite`.

### k6 Нагрузочные тесты (`tests/k6/`)

| Директория | Тесты |
|------------|-------|
| `auth/` | register, login, logout |
| `user/` | profile, get_user_id, upload_avatar, get_avatar |
| `chat/` | chats, messages, add_member, get_members, edit_message, delete_message, mark_read, rename_chat, delete_chat |
| `(root)` | smoke |

### Команды

| Команда | Описание |
|---------|----------|
| `make check` | golangci-lint + модульные + HTTP-тесты |
| `golangci-lint run ./...` | Только линтер (govet, staticcheck, ineffassign, unused) |
| `/root/go/bin/sqlc generate` | Регенерация sqlc-кода после изменения SQL |
| `make load-test SCRIPT=auth/login.js` | Запуск конкретного k6-теста |
| `make load-test VUS=50 DURATION=60s` | k6 с кастомными параметрами |
| `make app-up` | Запуск приложения (продакшн) |

---

## Известные ограничения

1. `POST /api/v2/chats/{chat_id}/messages` возвращает 201 без body — message_id нужно получать через `GET messages`.
2. `GetChatMessagesPaginated` и `GetUserChatsExtended` — raw SQL в репозитории (динамическая SQL, сложные LATERAL JOINs). Не через sqlc.

---

## Тестирование — итоги

| Suite | Результат |
|-------|-----------|
| auth/service | 7/7 PASS |
| auth/transport/http | 5/5 PASS |
| chat/repository/postgres | 35/35 PASS |
| chat/service | 31/31 PASS |
| chat/transport/http | 15/15 PASS |
| files/repository | 2/2 PASS |
| files/service | 2/2 PASS |
| files/transport/http | 6/6 PASS |
| user/repository/postgres | 8/8 PASS |
| user/service | 10/10 PASS |
| user/transport/http | 6/6 PASS |
| **ИТОГО** | **124/124 PASS** |

---

## Правила для ИИ-агента

1. **Язык** — все ответы строго на русском языке.
2. **Изоляция изменений** — работать только с тем, что указано в запросе. Не модифицировать смежные файлы без прямой необходимости.
3. **Ядро (internal/core/) не трогать** — категорически запрещено изменять или удалять файлы в `internal/core/` без прямого указания. Можно только использовать.
4. **WebSocket (internal/features/chat/transport/ws/) не трогать** — без прямого указания. Логика heartbeat, online/offline, writeMu — неизменна.
5. **Тестирование обязательно** — после каждого изменения кода необходимо:
   - Запустить `golangci-lint run ./...` (или `make check`)
   - Запустить `make check` (линтер + модульные + HTTP-тесты)
   - Если что-то не проходит — исправить и повторить цикл до зелёного состояния
6. **Swagger** — если меняется API (новый эндпоинт, изменение ответа), обновлять swagger-комментарии в хендлере и перегенерировать: `/root/go/bin/swag init -d ./cmd/app,./internal/features -o ./docs --parseDependency --parseInternal`
7. **sqlc** — после изменения SQL-файлов в `query/` выполнять `/root/go/bin/sqlc generate`. Генерированный код в `internal/db/` не редактировать вручную.
8. **Маскировка ошибок** — все 500 ошибки маскируются для клиента. Сырые ошибки БД/MinIO never не отдаются наружу, только в лог.
9. **Конфиденциальность** — запрещено логировать пароли и секретные ключи даже на уровне Debug.
10. **Транзакции** — в сервисах v2 транзакции открываются через `repo.Transaction(ctx, func(txCtx) { ... })`. Запрещено вызывать Begin/Commit/Rollback напрямую.
11. **Коммиты и пуш** — категорически запрещено делать git commit или git push без прямого указания пользователя. Даже если задача подразумевает изменения — показать результат, дождаться подтверждения, и только потом коммитить по команде. Сообщение коммита — строго на русском языке.
