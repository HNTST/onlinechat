# Инструкция для ИИ-агента (Контекст проекта)

Ты являешься опытным backend-разработчиком на Go. Твоя задача — помогать в разработке, рефакторинге, исправлении багов, покрытии тестами и поддержке этого проекта, строго соблюдая архитектурные правила, актуальный стек технологий, правила маскирования ошибок и ограничения безопасности.

---

## 1. Общее описание проекта и архитектура
* **Проект**: Backend мессенджера / чат-приложения в стиле Discord (монолит).
* **Тип**: HTTP REST API + WebSocket-сервер.
* **Архитектура**: Гибридная. В проекте сосуществуют два поколения кода под единым `net/http.ServeMux`:
  1. **Legacy (v1)** (`/api/v1/`): Старая 3-слойная архитектура (`internal/delivery/` -> `internal/service/` -> `internal/repository/`). Хендлеры парсят JSON вручную. Эндпоинты v1 и v2 частично пересекаются, но имеют разные сигнатуры (например, добавление участников в чат в v1 принимает одного пользователя, а в v2 — массив IDs). **Удалять или модифицировать v1-эндпоинты без прямого указания запрещено.** Файл `routes.md` в корне устарел и является архивным.
  2. **Новый слой (v2)** (`/api/v2/`): Чистая модульная архитектура (Feature-based) в `internal/features/`. Каждый модуль разделен на `transport/http/` (DTO + валидация), `service/` (бизнес-логика) и `repository/postgres/` (SQL). У модуля `chat` также присутствует `transport/ws/` для WebSocket v2. Инверсия зависимостей (DIP): интерфейсы сервиса объявляются в слое транспорта, интерфейсы репозитория — в слое сервиса.
  3. **Ядро проекта (Core)** (`internal/core/`): Общая инфраструктура приложения. Содержит глобальные конфигурации, доменные модели, пулы соединений, кастомный логер и middleware.
  4. **Auth + Сессии** (`internal/features/auth/`): Модуль аутентификации v2. Использует bcrypt для паролей, сессии хранятся в таблице `sessions` (токен 64 hex-символа, expires_at). Регистрация встроена в auth, а не в user module. Парольный хардкод "-1" удалён, `bcrypt` интегрирован в `user service SaveUser`.

### Защищённые эндпоинты и user_id из контекста сессии
* AuthMiddleware (`internal/core/transport/middleware/auth.go`) проверяет cookie `session_token`, извлекает `user_id` (int64) и кладёт в контекст по ключу `core_http_middleware.UserIDKey`.
* Исключения (без сессии): `/swagger/`, `/api/v2/auth/login`, `/api/v2/auth/register`.
* Все защищённые хендлеры v2 (чаты: список, rename, delete, mark_read; профиль пользователя) **читают `user_id` строго из контекста**, а не из query-параметров.
* Все эндпоинты v2 (кроме `/swagger/`, `/api/v2/auth/login`, `/api/v2/auth/register`) требуют валидную сессию через cookie `session_token`.

---

## 2. Актуальный технологический стек
* **Язык**: Go 1.26.2 (используй встроенные возможности и строго подключенные зависимости).
* **Маршрутизация**: Встроенный `net/http.ServeMux`, обернутый в кастомный `APIVersionRouter` (`internal/core/transport/server/router.go`), разделяющий потоки на `/api/v1/` и `/api/v2/`.
* **Цепочка Middleware (справа налево)**: `Cors` → `RequestID` → `Logger` → `Panic` → `Trace` → `Auth` (инициализируется в `main.go`). Auth middleware проверяет cookie `session_token`, кладёт `user_id` в контекст. Исключения: `/swagger/`, `/api/v2/auth/login`, `/api/v2/auth/register`.
* **База данных**: PostgreSQL. Работа через пул соединений `pgx/v5` (`jackc/pgx/v5/pgxpool`) с интерфейсом `Pool`, включающим `CtxWithTransaction` и `GetTx`.
* **Миграции**: Автоматические через `golang-migrate/v4`. Функция `RunMigrations(pgPool)` вызывается в `main.go` при старте. Файлы миграций расположены в `migrations/000001_init_schema.{up,down}.sql`.
* **Real-time Доставка**: Реализована на базе библиотеки `github.com/gorilla/websocket`. Утечки горутин при graceful shutdown устранены в обеих версиях (v1 и v2) путем отслеживания `ctx.Done()` и принудительного закрытия `conn.Close()`.
* **Объектное хранилище (Двойное хранилище фичи `files`)**: За работу с файлами отвечает выделенный v2-модуль `internal/features/files/`. Метаданные пишутся в `repository/postgres/`, а бинарные файлы — в хранилище MinIO через `repository/minio.go` (используется пакет `minio-go/v7`). Интерфейсы `FileRepository` и `FileStorage` определены в слое `service/service.go`.
* **Валидация**: `go-playground/validator/v10` через теги структур в DTO (применяется строго в слое v2).
* **Логирование**: Встроенный структурированный логер `log/slog`, обернутый в `internal/core/logger/`.
* **Тестирование**: Фреймворк `github.com/stretchr/testify` (assert, mock, suite).

---

## 3. Картография маршрутов v2 (API v2 Routes)
* `POST /api/v2/auth/register` — Регистрация + авто-логин (body: `{"username": "...", "password": "..."}`). При успехе устанавливает HttpOnly cookie `session_token` и возвращает профиль (201). AuthMiddleware не применяется.
* `POST /api/v2/auth/login` — Аутентификация пользователя (body: `{"username": "...", "password": "..."}`). При успехе устанавливает HttpOnly cookie `session_token` и возвращает профиль пользователя. AuthMiddleware не применяется.
* `POST /api/v2/auth/logout` — Завершение сессии. Требует cookie `session_token`. Удаляет сессию из БД и очищает cookie. Возвращает 200 OK.
* `GET /api/v2/user/profile` — Получить свой профиль (требует сессию). `user_id` читается из контекста. Возвращает `id`, `username`, `avatar_id`.
* `GET /api/v2/users` — Получить произвольного пользователя по id (query: `id`). Требует сессию. Возвращает `id`, `username`, `avatar_id`.
* `GET /api/v2/user_id` — Получить ID пользователя по его имени (query: `username`). Возвращает `{"user_id": int64}`.
* `POST /api/v2/chat` — Создать новый чат. Тело: `{"chat_name": "...", "author": int64, "target_user": int64, "is_group": bool}`. Флаг `is_group` опционален (по умолчанию `false`). Для групповых чатов проверка существующего чата пропускается.
* `PUT /api/v2/chat/{chat_id}/name` — Переименовать групповой чат (body: `{"name": "..."}`). `user_id` читается из контекста. Только для `is_group=true`. Любой участник может переименовать.
* `DELETE /api/v2/chat/{chat_id}` — Удалить групповой чат. `user_id` читается из контекста. Только для `is_group=true`. Каскадно удаляет сообщения и участников (ON DELETE CASCADE).
* `POST /api/v2/chat/{chat_id}/members` — Добавить пользователей в чат (path-параметр `chat_id`, body: `{"users_ids": [...]}`).
* `GET /api/v2/chats` — Получить чаты пользователя с расширенной информацией. `user_id` читается из контекста. Возвращает массив чатов, каждый содержит: `id`, `name`, `is_group`, `last_message` (объект с полями `id`, `text`, `sender_id`, `sent_at`, `is_read` или `null`), `unread_count` (число непрочитанных сообщений от других пользователей).
* `GET /api/v2/chats/{chat_id}/messages` — Получить историю сообщений чата с курсорной пагинацией (query: `last_message_id` — опциональный курсор, `limit` — опциональный лимит, по умолчанию 50, макс. 200). Удалённые сообщения возвращаются с `is_deleted=true` и текстом "Сообщение удалено".
* `GET /api/v2/chats/members` — Получить список участников чата (query: `chat_id`).
* `POST /api/v2/chats/{chat_id}/messages` — Отправить/сохранить сообщение (body: `{"sender_id": int64, "message": "..."}`).
* `PUT /api/v2/chats/{chat_id}/messages/{message_id}` — Редактировать текст сообщения (body: `{"text": "..."}`). Устанавливает `is_edited=true`.
* `DELETE /api/v2/chats/{chat_id}/messages/{message_id}` — Мягкое удаление сообщения (устанавливает `is_deleted=true`, очищает текст).
* `PATCH /api/v2/chats/{chat_id}/messages/read-status` — Отметить сообщения как прочитанные (body: `{"last_read_message_id": int64}`). `user_id` читается из контекста. Помечает все чужие сообщения в чате до указанного ID как `is_read=true`.
* `GET /api/v2/ws` — WebSocket-соединение нового поколения. Поддерживает: обмен сообщениями в реальном времени + уведомления о статусе онлайн/оффлайн пользователей.
* `POST /api/v2/user/avatar` — Загрузка файла аватарки (multipart/form-data).
* `GET /api/v2/user/avatar` — Прямой бинарный стриминг изображения аватарки (query: `id` — соответствует `avatar_id`). Устанавливает `Content-Type` на основе данных из БД (`type_file`). Ошибка `io.Copy` здесь логируется на уровне `Debug` (маскирует штатный обрыв соединения клиентом).


### Структуры ответов

**`GET /api/v2/chats` (расширенный ответ):**
```json
{
  "chats": [
    {
      "id": 1,
      "name": "chat-name",
      "is_group": false,
      "last_message": {
        "id": 10,
        "text": "Последнее сообщение",
        "sender_id": 2,
        "sent_at": "2026-06-28T12:00:00Z",
        "is_read": false
      },
      "unread_count": 3
    }
  ]
}
```

**`GET /api/v2/messages` (пагинация):**
```json
{
  "messages": [
    {
      "id": 1,
      "sender_id": 2,
      "message": "Текст сообщения",
      "sent_at": "2026-06-28T12:00:00Z",
      "is_edited": false,
      "is_deleted": false,
      "is_read": false
    }
  ]
}
```

Для удалённых сообщений: `is_deleted=true`, `message="Сообщение удалено"`.

### WebSocket: обмен сообщениями и статус онлайн/оффлайн

**Входящее сообщение (от клиента):**
```json
{ "message": "text", "chat_id": 1, "sender_id": 123 }
```

**Исходящие сообщения (от сервера):**

1. Новое сообщение в чате:
```json
{ "id": 456, "message": "text", "chat_id": 1, "sender_id": 123, "sent_at": "2025-01-01T00:00:00Z", "is_edited": false, "is_deleted": false, "is_read": false }
```

2. Пользователь онлайн (подключился к WebSocket):
```json
{ "type": "online", "user_id": 123 }
```

3. Пользователь оффлайн (отключился от WebSocket):
```json
{ "type": "offline", "user_id": 123 }
```

**Логика:** При подключении пользователя сервер находит всех его чат-партнёров через `GetChatPartnerIDs`, проверяет какие из них сейчас в `clients` map (подключены), и отправляет им `{"type":"online","user_id":X}`. При отключении — аналогично отправляет `{"type":"offline","user_id":X}`.

### Структура БД (миграция 000002)

Таблица `messages` получила три новых колонки:
| Колонка      | Тип     | По умолч. | Описание                              |
|-------------|---------|-----------|---------------------------------------|
| `is_edited` | BOOLEAN | FALSE     | Было ли сообщение отредактировано     |
| `is_deleted`| BOOLEAN | FALSE     | Флаг мягкого удаления                 |
| `is_read`   | BOOLEAN | FALSE     | Прочитано ли сообщение получателем    |

### Структура БД (миграция 000004)

| Таблица    | Колонка     | Тип                       | Ограничения      | Описание                     |
|-----------|-------------|---------------------------|------------------|------------------------------|
| `sessions`| `id`        | BIGSERIAL                 | PRIMARY KEY      | Идентификатор сессии         |
|           | `user_id`   | BIGINT                    | NOT NULL, FK -> users(id) ON DELETE CASCADE | Владелец сессии |
|           | `token`     | VARCHAR(255)              | NOT NULL, UNIQUE | Случайный токен (64 hex символа) |
|           | `expires_at`| TIMESTAMP WITH TIME ZONE  | NOT NULL         | Время истечения сессии       |

### Структура БД (миграция 000003)

| Таблица  | Изменение                                       |
|----------|-------------------------------------------------|
| `users`  | `avatar_id` теперь имеет `DEFAULT 1`            |
| `chats`  | Добавлена колонка `is_group BOOLEAN NOT NULL DEFAULT false` |

| Колонка    | Тип     | По умолч. | Описание             |
|-----------|---------|-----------|----------------------|
| `is_group` | BOOLEAN | FALSE     | Флаг группового чата |

### Оптимизированный SQL-запрос для `GET /api/v2/chats`

```sql
SELECT
    c.id, c.name, c.is_group,
    lm.id, lm.message, lm.sender_id, lm.sent_at, lm.is_read,
    COALESCE(uc.unread_count, 0)
FROM chats c
JOIN chat_members cm ON c.id = cm.chat_id
LEFT JOIN LATERAL (
    SELECT id, message, sender_id, sent_at, is_read
    FROM messages
    WHERE chat_id = c.id AND is_deleted = false
    ORDER BY id DESC
    LIMIT 1
) lm ON true
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS unread_count
    FROM messages
    WHERE chat_id = c.id AND sender_id != $2 AND is_read = false
) uc ON true
WHERE cm.user_id = $1
ORDER BY c.id
```

---

## 4. Доменные ошибки (internal/core/errors/errs.go)

| Ошибка | Описание | HTTP статус |
|--------|----------|-------------|
| `ErrUserAlreadyExists` | Пользователь уже существует | 409 |
| `ErrInternal` | Внутренняя ошибка сервера | 500 (маскируется) |
| `ErrInvalidJSON` | Неверный формат JSON | 400 |
| `ErrParamMissing` | Отсутствует обязательный параметр | 500 (не замаплен в ErrorResponse, попадает в default) |
| `ErrEncodingRequest` | Ошибка кодирования JSON | 500 |
| `ErrUserNotFound` | Пользователь не найден | 404 |
| `ErrInvalidArgument` | Неверный аргумент | 400 |
| `ErrConfict` | Конфликт | 409 |
| `ErrChatNotExists` | Чат не существует | 404 |
| `ErrMessageNotFound` | Сообщение не найдено | 404 |
| `ErrUnauthorized` | Не авторизован | 401 |

## 5. Критические правила разработки и безопасности

### Маскирование ошибок и Логирование
1. Централизованная обработка ответов в `HTTPResponseHandler.errorResponse` строго фильтрует ошибки. 
2. **Все ответы со статусом 500 (Internal Server Error) маскируются.** Клиент в JSON получает обобщенный текст `"Внутренняя ошибка сервера"`. Сырые технические ошибки базы данных (ошибки сканирования `pgx`, `no rows`), MinIO, или сетевые таймауты **никогда** не должны отдаваться наружу. Реальная ошибка пишется исключительно в системный лог через `slog`.
3. Доменные и валидационные ошибки (`ErrInvalidArgument` -> 400, `ErrUserNotFound` -> 404, `ErrConflict` -> 409) отдаются клиенту как есть. Репозитории обязаны явно переводить ошибки драйвера в доменные (например, `pgx.ErrNoRows` -> `core_errors.ErrUserNotFound`).
4. **Важно**: Ошибки от `DecodeAndValidateRequest` (валидатор `go-playground/validator`) не являются экземплярами `core_errors.ErrInvalidArgument`. Хендлеры v2 оборачивают их: `responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, msg)` — **всегда** передавай `core_errors.ErrInvalidArgument` вместо сырой ошибки валидации, чтобы получить HTTP 400.
5. **Запрещено** логировать сырые конфиденциальные данные (пароли пользователей, секретные ключи MinIO/Postgres) даже на уровне `Debug`.

### Управление транзакциями и Контекст
1. Запрещено использовать сырые строки в качестве ключей контекста. Для транзакций используется кастомный неэкспортируемый тип `txKeyType` и переменная `txKey`.
2. В сервисах v2 транзакции открываются через `repo.Transaction(ctx, func(txCtx) { ... })`. Метод `CtxWithTransaction` автоматически подхватывает уже открытую транзакцию через `GetTx`. Если транзакция уже открыта — колбэк выполняется внутри неё; если нет — создаётся новая. Запрещено вызывать `Begin/Commit/Rollback` напрямую в репозиториях в обход пула.

### Конфигурация и Окружение
1. Все файлы конфигурации (`.env`, `.env.test`, `.env.example`) находятся строго в директории `internal/core/config/`. Загрузка в коде выполняется через `godotenv.Load("internal/core/config/.env")`.
2. Критические параметры (пароли `DB_PASSWORD`, `MINIO_ROOT_PASSWORD`) считываются через метод `requireEnv()`. При их отсутствии в рантайме приложение падает через `log.Fatalf`. Запрещено хардкодить девелоперские IP-адреса и дефолтные пароли в кодовой базе.
3. *Внимание на баг конфигурации:* В `.env.example` используется ошибочный ключ `SSLMODE`, тогда как `config.go` ожидает `DB_SSLMODE`. Исправь при первой работе с файлом.

### Ограничения по работе со структурами
1. **Категорически запрещено изменять или удалять файлы внутри общей инфраструктуры ядра `internal/core/`** без прямой, отдельной команды пользователя.
2. Запрещено модифицировать или очищать закомментированные legacy-блоки в v1-коде, если на это нет прямого указания.
3. *Внимание на баг legacy:* В файле `internal/delivery/http.go` на строке 336 вызывается метод `sv.sendError()`, который отсутствует в структуре. Это мертвый код (dead code), но не трогай его без команды.

### Технический долг (Целевое состояние миграций)
* Приоритетная задача рефакторинга инфраструктуры миграций: необходимо перевести загрузку файлов из папки `migrations/` с дискового пути (`file://migrations`) на вшивание бинарника через встроенный пакет Go `//go:embed migrations/*.sql`.

---

## 6. Правила работы ИИ-агента

1. **Языковой стандарт**: На любые запросы пользователя, вопросы или в финальных отчетах о проделанной работе ты обязан **отвечать строго на русском языке**.
2. **Категорический запрет на автоматические коммиты**: Тебе **строго запрещено** запускать таргет `make commit`, `git commit` или осуществлять любые другие действия по фиксации изменений в репозитории, даже если таргет `make commit` запрашивает интерактивное подтверждение (y/N). Коммиты делаются исключительно пользователем вручную в терминале хоста.
3. **Изоляция изменений**: Изменения по задачам v2 должны вноситься изолированно, в рамках папки конкретной фичи.

---

## 7. Тестовая инфраструктура (Docker Compose)

Тестовая кодовая база строго разделена на три типа тестов, все они используют фреймворк `testify`:

1. **Модульные тесты сервисов (Unit)**: Находятся в `tests/features/*/service/`. Не имеют внешних зависимостей, используют изолированные моки (`testify/mock`). Входят в `make check`.
2. **Интеграционные HTTP-тесты транспорта**: Находятся в `tests/features/*/transport/http/`. Используют `httptest.Server`, подмоканные сервисы, но прогоняют запросы через **реальную цепочку middleware** (Logger, Panic). Следуют паттерну `setupServer(t, mockSvc)` (и `setupServer(t, mockSvc, userID)` для тестов с сессией). Для симуляции аутентифицированного пользователя в тестах используется хелпер `withUser(userID)`, который подкладывает `user_id` в контекст. Входят в `make check`.
3. **Интеграционные тесты репозиториев (Docker)**: Находятся в `tests/features/*/repository/postgres/`, `tests/features/files/repository/` (`minio_repository_test.go`) и `tests/repository/postgres/`. Тесты баз данных и хранилищ используют паттерн `testify/suite` (`SetupSuite`/`TearDownSuite`) и требуют запущенных контейнеров.
4. **Новые тесты v2 (чат)**: Добавлены тесты для: `EditMessage`, `SoftDeleteMessage`, `MarkMessagesAsRead`, `GetChatMessagesPaginated`, `GetUserChatsExtended`, `RenameChat`, `DeleteChat`, `UpdateChatName`, `IsUserInChat` — во всех трёх слоях (service, transport/http, repository/postgres).

### Команды для работы:
* `make app-up` — Запуск основного приложения (требует наличия настроенного `.env` файла в `internal/core/config/`).
* `make check` — Быстрый запуск линтера, `go vet`, модульных тестов и интеграционных тестов транспорта. Выполняй после любых изменений кода.
* `make test-integration` — Полный цикл интеграционного тестирования инфраструктуры. Команда разворачивает окружение в Docker на основе файла `docker-compose.test.yml` (поднимает контейнеры `postgres:15-alpine` и `minio/minio:latest`), считывает параметры из `internal/core/config/.env.test`, автоматически накатывает миграции, прогоняет тесты репозиториев и полностью уничтожает контейнеры вместе со всеми созданными томами (`docker compose down -v`).

### Нагрузочное тестирование (K6)

Тесты расположены в `tests/k6/`. Используют фреймворк [k6](https://k6.io/) для нагрузочного и функционального тестирования API.

**Структура:**
```
tests/k6/
├── fixtures/
│   └── test.png            — Минимальный 1x1 PNG для тестов загрузки аватарок
├── auth/
│   ├── register.js          — Регистрация (iterations:1, уникальные данные)
│   ├── login.js             — Логин + профиль + негативные сценарии (duration)
│   └── logout.js            — Логаут + проверка сессии (duration)
├── user/
│   ├── profile.js           — Профиль + поиск пользователя (duration)
│   ├── get_user_id.js       — Поиск ID по username (duration)
│   ├── upload_avatar.js     — Загрузка аватарки (duration)
│   └── get_avatar.js        — Получение аватарки по file_id (duration)
├── chat/
│   ├── chats.js             — Создание и получение чатов (duration)
│   ├── messages.js          — Отправка и получение сообщений (duration)
│   ├── add_member.js        — Добавление участника в чат (duration)
│   ├── get_members.js       — Список участников чата (duration)
│   ├── edit_message.js      — Редактирование сообщения (duration)
│   ├── delete_message.js    — Удаление сообщения (duration)
│   ├── mark_read.js         — Отметка прочтения (duration)
│   ├── rename_chat.js       — Переименование группового чата (duration)
│   └── delete_chat.js       — Удаление группового чата (duration)
└── smoke.js                 — Базовый smoke-тест (duration)
```

**Паттерны:**
- Аутентификация через cookie `session_token` (хелпер `extractToken()`)
- Уникальные данные через тег `${__VU}_${__ITER}` в username
- Все эндпоинты (кроме register/login) требуют cookie в заголовке `Cookie: session_token=...`
- Для проверки отсутствия авторизации используется трюк `BASE_URL.replace('localhost', '127.0.0.1')` — cookie jar k6 привязан к origins
- k6 `open()` для чтения файлов работает только в init stage (глобальная область)
- k6 `http.file()` принимает `{ field_name: http.file(data, filename) }` как body
- `POST /api/v2/chats/{chat_id}/messages` возвращает 201 без body — message_id нужно получать через `GET messages`
- JSON-поле сообщения: `id` (не `message_id`)
- Migration 000005 вставляет `id=1` в `files` с `ON CONFLICT DO NOTHING` — необходим `setval()` после вставки для корректной последовательности SERIAL

**Типы тестов:**
| Тест | Тип | Режим k6 | Описание |
|------|-----|----------|----------|
| `register.js` | Уникальные данные | `iterations: 1` | Одна регистрация, проверка дубликата |
| `login.js` | Уникальные данные | `duration` | Регистрация → логин → профиль → негативные |
| `logout.js` | Уникальные данные | `duration` | Регистрация → логаут → проверка сброса |
| `profile.js` | Уникальные данные | `duration` | Регистрация → профиль → поиск по id |
| `get_user_id.js` | Уникальные данные | `duration` | Регистрация → поиск по username → 404 → без auth |
| `upload_avatar.js` | Уникальные данные | `duration` | Регистрация → multipart загрузка PNG → file_id |
| `get_avatar.js` | Уникальные данные | `duration` | Загрузка → получение бинарного потока → 404 |
| `chats.js` | Уникальные данные | `duration` | 2 пользователя → чат → список → дубликат |
| `messages.js` | Уникальные данные | `duration` | 2 пользователя → чат → сообщения → пагинация |
| `add_member.js` | Уникальные данные | `duration` | 3 пользователя → чат → добавление участника |
| `get_members.js` | Уникальные данные | `duration` | 2 пользователя → чат → список участников |
| `edit_message.js` | Уникальные данные | `duration` | Отправка → редактирование → проверка is_edited |
| `delete_message.js` | Уникальные данные | `duration` | Отправка → мягкое удаление |
| `mark_read.js` | Уникальные данные | `duration` | Отправка → отметка прочтения |
| `rename_chat.js` | Уникальные данные | `duration` | Групповой чат → переименование |
| `delete_chat.js` | Уникальные данные | `duration` | Групповой чат → удаление |
| `smoke.js` | Нагрузка | `duration` | Запрос пользователя по id |

**Команды:**
```bash
make load-test                                  # Регистрация (по умолчанию)
make load-test SCRIPT=auth/login.js             # Конкретный тест
make load-test VUS=50 DURATION=60s              # Кастомные параметры
make load-test-register                         # Ярлык: регистрация
make load-test-login                            # Ярлык: логин
make load-test-logout                           # Ярлык: логаут
make load-test-smoke                            # Ярлык: smoke
make load-test-profile                          # Ярлык: профиль
make load-test-chats                            # Ярлык: чаты
make load-test-messages                         # Ярлык: сообщения
make load-test-get-user-id                      # Ярлык: поиск user_id
make load-test-upload-avatar                    # Ярлык: загрузка аватарки
make load-test-get-avatar                       # Ярлык: получение аватарки
make load-test-add-member                       # Ярлык: добавление участника
make load-test-get-members                      # Ярлык: список участников
make load-test-edit-message                     # Ярлык: редактирование сообщения
make load-test-delete-message                   # Ярлык: удаление сообщения
make load-test-mark-read                        # Ярлык: отметка прочтения
make load-test-rename-chat                      # Ярлык: переименование чата
make load-test-delete-chat                      # Ярлык: удаление чата
```

**Важно:** Каждый `load-test-*` таргет автоматически:
1. Поднимает Docker окружение (`docker-compose.test.yml`) с предварительной очисткой (`down -v`)
2. Собирает бинарник (`APP_ENV=test go build -o .bin/test-server ./cmd/app`)
3. Запускает Go-сервер в тестовом режиме (`APP_ENV=test .bin/test-server`) с трекингом PID
4. Останавливает сервер по PID после завершения k6
5. Останавливает и чистит Docker (`docker compose down -v`)

**Известные ошибки сервера (не баги тестов):**
- `ErrParamMissing` не замаплен в `ErrorResponse` — возвращает 500 вместо 400 (см. `handler.go:56 default`)
- Migration 000005:序列 SERIAL `files_id_seq` не обновляется после `INSERT ... ON CONFLICT DO NOTHING` с явным `id=1` — исправлено добавлением `SELECT setval(...)`
