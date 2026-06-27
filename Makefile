app-up:
	@go run cmd/app/main.go

check:
	go vet ./...
	go test -p 1 -count=1 -v ./...

test-integration:
	@echo "=== Поднимаем тестовую PostgreSQL ==="
	docker compose -f docker-compose.test.yml up -d --wait
	@echo "=== Запускаем тесты ==="
	go test -p 1 -count=1 -v ./tests/... || true
	@echo "=== Останавливаем и чистим контейнер ==="
	docker compose -f docker-compose.test.yml down -v

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
