# MyInvestor

MyInvestor — симулятор трейдинга, который позволит пользователям практиковаться в покупке и продаже финансовых активов в условиях, приближённых к реальным.

> Проект находится в разработке. Сейчас доступен базовый HTTP-сервер с health-check endpoint; торговые сценарии и подключение к базе данных ещё не реализованы.

## Требования

- Go 1.26.10
- Git
- Docker Desktop (Windows/macOS) либо Docker Engine и Compose plugin (Linux) — только если нужна локальная PostgreSQL

## Структура проекта

```text
.
├── cmd/
│   └── api/
│       └── main.go          # точка входа HTTP API
├── internal/
│   ├── config/              # конфигурация этого приложения
│   ├── handler/             # HTTP-обработчики
│   ├── repository/          # доступ к данным и внешним системам
│   └── service/             # бизнес-логика
├── pkg/                     # переиспользуемые компоненты (если появятся)
├── .air.toml
├── .dockerignore
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── CONTRIBUTING.md
```

Каталоги внутри `internal/` подготовлены для дальнейшей разработки; некоторые из них пока пусты. Конфигурация MyInvestor размещается в `internal/config`, поскольку она относится только к этому приложению. Каталог `pkg/` следует использовать для компонентов, которые действительно планируется переиспользовать в других проектах.

## API

| Метод | Путь | Назначение |
|---|---|---|
| `GET` | `/api/v1/health` | Проверка доступности сервера |

## Проверки

Запустить тесты:

```bash
go test ./...
```

## Запуск backend для разработки

Все команды выполняйте из корневой папки репозитория.

### 1. Установить зависимости проекта и Air

```bash
go mod download
go install github.com/air-verse/air@latest
```

Убедитесь, что каталог `bin` внутри GOPATH находится в `PATH`. Узнать расположение GOPATH можно командой `go env GOPATH`. Если `air` не находится, добавьте каталог GOPATH/bin в PATH.

Для текущей сессии терминала:

**macOS / Linux (Bash или Zsh):**

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

**Windows (PowerShell):**

```powershell
$env:Path = "$(go env GOPATH)\bin;$env:Path"
```

Для постоянного использования добавьте этот каталог в пользовательскую переменную PATH средствами настроек операционной системы. После настройки можно проверить доступность Air командой `air -v`.

### 2. (Необязательно) Запустить PostgreSQL

Текущая версия API не использует базу данных, поэтому этот шаг можно пропустить. Если для разработки нужна локальная PostgreSQL, запустите Docker Desktop (Windows/macOS) или Docker Engine с Compose plugin (Linux), затем выполните:

```bash
docker compose up -d
```

Compose запустит PostgreSQL в Docker. По умолчанию используются база `myinvestor`, пользователь `investor` и пароль `investor_dev`. Для переопределения параметров создайте в корне файл `.env` в формате `KEY=value` и задайте `POSTGRES_DB`, `POSTGRES_USER` и `POSTGRES_PASSWORD`. Значения по умолчанию предназначены только для локальной разработки.

### 3. Запустить backend через Air

В том же или другом терминале, из корня репозитория:

```bash
air
```

Air пересобирает и перезапускает API при изменениях Go-файлов. Сервер доступен на `http://localhost:8080`. Откройте `http://localhost:8080/api/v1/health` в браузере, чтобы проверить health-check. Ожидаемый JSON-ответ:

```json
{"status": "ok", "app": "MyInvestor"}
```

Остановите Air сочетанием `Ctrl+C`. Если запускали PostgreSQL, остановить и удалить контейнер можно командой:

```bash
docker compose down
```

Данные PostgreSQL сохраняются в Docker volume и не удаляются командой выше. Чтобы удалить также базу и её данные, используйте `docker compose down --volumes`.

`Dockerfile` собирает отдельный образ API; описанный выше Compose-сценарий запускает только PostgreSQL, а backend запускается локально через Air.

## Участие в разработке

Правила работы с ветками, Pull Request и структурой проекта описаны в [CONTRIBUTING.md](./CONTRIBUTING.md).
