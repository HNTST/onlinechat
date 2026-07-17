app-up:
	APP_ENV=prod go run cmd/app/main.go

check:
	golangci-lint run ./...
	APP_ENV=test go test -p 1 -count=1 -v ./...

# === Тестовая инфраструктура (Docker) ===
test-infra-up:
	docker compose -f docker-compose.test.yml down -v
	docker compose -f docker-compose.test.yml up --build --wait -d

test-infra-down:
	docker compose -f docker-compose.test.yml down -v

# === Интеграционные тесты ===
test-integration: test-infra-up
	@echo "=== Запускаем тесты ==="
	@APP_ENV=test go test -p 1 -count=1 -v ./tests/...; \
	rc=$$?; \
	echo "=== Останавливаем и чистим контейнер ==="; \
	make test-infra-down; \
	exit $$rc

# === Запуск сервера в тестовом режиме ===
run-test-api: test-infra-up
	@echo "=== Запускаем API-сервер в тестовом режиме ==="
	APP_ENV=test go run cmd/app/main.go

# === Нагрузочные тесты (k6) ===
K6_RUN = VUS=$(or $(VUS),10) DURATION=$(or $(DURATION),30s) k6 run

load-test: test-infra-up
	@echo "=== Сборка тестового бинарника ==="
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@echo "=== Запуск API-сервера ==="
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid
	@sleep 3
	@$(K6_RUN) tests/k6/$(or $(SCRIPT),auth/register).js; \
	rc=$$?; \
	echo "=== Останавливаем API-сервер ==="; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	echo "=== Останавливаем Docker ==="; \
	make test-infra-down; \
	exit $$rc

load-test-register: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/auth/register.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-login: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/auth/login.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-logout: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/auth/logout.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-smoke: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/smoke.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-profile: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/user/profile.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-chats: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/chat/chats.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-messages: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/chat/messages.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-get-user-id: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/user/get_user_id.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-upload-avatar: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/user/upload_avatar.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-get-avatar: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/user/get_avatar.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-add-member: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/chat/add_member.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-get-members: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/chat/get_members.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-edit-message: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/chat/edit_message.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-delete-message: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/chat/delete_message.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-mark-read: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/chat/mark_read.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-rename-chat: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/chat/rename_chat.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

load-test-delete-chat: test-infra-up
	@APP_ENV=test go build -o .bin/test-server ./cmd/app
	@APP_ENV=test .bin/test-server & echo $$! > .server.pid; \
	sleep 3; \
	$(K6_RUN) tests/k6/chat/delete_chat.js; \
	rc=$$?; \
	kill $$(cat .server.pid) 2>/dev/null || true; \
	rm -f .server.pid .bin/test-server; \
	make test-infra-down; \
	exit $$rc

commit:
	@git add .; \
	changes=$$(git diff --cached --name-status); \
	if [ -z "$$changes" ]; then \
		echo "Нет изменений для коммита"; \
		exit 0; \
	fi; \
	echo "===== Стейджeные изменения ====="; \
	git diff --cached --stat; \
	echo ""; \
	echo "===== Предлагаемое сообщение коммита ====="; \
	msg=""; \
	while IFS= read -r line; do \
		case "$$line" in \
			M*) prefix="Modified:" ;; \
			A*) prefix="Added:" ;; \
			D*) prefix="Deleted:" ;; \
			R*) prefix="Renamed:" ;; \
			*) prefix="Changed:" ;; \
		esac; \
		file=$$(echo "$$line" | cut -c3-); \
		item="$$prefix $$file"; \
		if [ -z "$$msg" ]; then \
			msg="$$item"; \
		else \
			msg="$$msg"$$'\n'"$$item"; \
		fi; \
	done <<< "$$changes"; \
	echo "$$msg"; \
	echo ""; \
	echo "Команда для коммита:"; \
	echo "  git commit -m \"$$(echo "$$msg" | head -1)\" -m \"$$(echo "$$msg" | tail -n +2)\""; \
	echo ""; \
	read -p "Выполнить коммит? (y/N): " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		tmpfile=$$(mktemp); \
		echo "$$msg" > "$$tmpfile"; \
		git commit -F "$$tmpfile"; \
		rm -f "$$tmpfile"; \
	else \
		echo "Коммит отменён"; \
	fi
