# CI-пайплайн: правильная настройка деплоя на Dokploy

> Заметка о том, как корректно связать GitHub Actions CI с деплоем на Dokploy,
> чтобы деплой происходил **только после успешного прохождения проверок**.

## Цель

- Push в `main` → запуск CI (go vet, golangci-lint, go test).
- Только если все проверки зелёные → запуск деплоя на Dokploy.
- Push с упавшим CI **не должен** деплоить.

## Два механизма деплоя в Dokploy (не путать)

1. **Auto Deploy (GitHub-интеграция / Git Source)** — Dokploy сам слушает push в репозиторий
   и деплоит **независимо от GitHub Actions**. Если включён — код уедет на прод даже при
   упавших тестах. Для схемы «деплой только после CI» — **выключить**.
2. **Webhook URL / API** — деплой запускается внешним запросом (например, из пайплайна).

### Подводный камень webhook

Webhook для Git-провайдеров сверяет **ветку из полезной нагрузки** с веткой, настроенной
в Dokploy. Запрос `curl -X POST <webhook_url>` без тела даёт:

```json
{ "message": "Branch Not Match" }
```

А при выключенном Auto Deploy webhook отвечает:

```json
{ "message": "Automatic deployments are disabled for this application" }
```

## Правильный способ: Dokploy API (рекомендуется)

Работает из CI, не зависит от push-событий и от тумблера Auto Deploy.

### Шаги в Dokploy

1. Профиль → настройки → **Generate API Token**.
2. Получить `applicationId` своего приложения:

```bash
curl -X 'GET' 'https://dokploy.yourdomain/api/project.all' \
  -H 'accept: application/json' \
  -H 'x-api-key: <token>'
```

3. В настройках приложения выключить **Auto Deploy** (вкладка General), чтобы push
   сам по себе ничего не деплоил.

### Секреты в GitHub (Settings → Secrets and variables → Actions)

| Секрет                  | Значение                                   |
|-------------------------|--------------------------------------------|
| `DOKPLOY_API_URL`       | `https://dokploy.yourdomain`               |
| `DOKPLOY_API_KEY`       | API-токен                                  |
| `DOKPLOY_APPLICATION_ID`| ID приложения из `project.all`             |

### Шаг в `.github/workflows/ci.yml`

```yaml
- name: Deploy на Dokploy
  if: success()
  env:
    DOKPLOY_API_URL: ${{ secrets.DOKPLOY_API_URL }}
    DOKPLOY_API_KEY: ${{ secrets.DOKPLOY_API_KEY }}
    DOKPLOY_APPLICATION_ID: ${{ secrets.DOKPLOY_APPLICATION_ID }}
  run: |
    if [ -z "$DOKPLOY_API_URL" ] || [ -z "$DOKPLOY_API_KEY" ] || [ -z "$DOKPLOY_APPLICATION_ID" ]; then
      echo "Секреты Dokploy не заданы — шаг деплоя пропущен"
      exit 0
    fi
    curl -fsSL -X POST "$DOKPLOY_API_URL/api/application.deploy" \
      -H "Content-Type: application/json" \
      -H "x-api-key: $DOKPLOY_API_KEY" \
      -d "{\"applicationId\": \"$DOKPLOY_APPLICATION_ID\"}"
```

### Ключевое условие — гейт

Шаг деплоя стоит **в конце** пайплайна с `if: success()`: если CI упал, шаг не выполнится
и деплой не запустится. Дополнительно: **Auto Deploy в Dokploy выключен**, поэтому
единственный путь к деплою — этот шаг.

## Итоговая схема

```
push в main → GitHub Actions CI (vet + lint + tests)
                  │ зелёный?
                  ├── нет → стоп (деплоя нет)
                  └── да → POST /api/application.deploy → Dokploy
```
