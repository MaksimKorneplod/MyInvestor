# MyInvestor

MyInvestor — симулятор трейдинга, который позволит пользователям практиковаться в покупке и продаже финансовых активов в условиях, приближённых к реальным.

> Проект находится в разработке. Сейчас доступен базовый HTTP-сервер с health-check endpoint и подключением к PostgreSQL.

## Требования

- Go 1.26.10
- Git
- Docker Desktop (Windows/macOS) либо Docker Engine и Compose plugin (Linux)
- `migrate` CLI для управления миграциями

## Структура проекта

```text
.
├── cmd/
│   └── api/
│       └── main.go          # точка входа HTTP API
├── internal/
│   ├── config/              # конфигурация этого приложения
│   ├── database/            # подключение к PostgreSQL
│   ├── handler/             # HTTP-обработчики
│   ├── repository/          # доступ к данным и внешним системам
│   └── service/             # бизнес-логика
├── migrations/              # версионируемые SQL-миграции
├── pkg/                     # переиспользуемые компоненты (если появятся)
├── .air.toml
├── .dockerignore
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── CONTRIBUTING.md
```

Каталоги внутри `internal/` подготовлены для дальнейшей разработки; некоторые из них пока пусты. Конфигурация MyInvestor размещается в `internal/config`, поскольку она относится только к этому приложению. Каталог `pkg/` следует использовать для компонентов, которые действительно планируется переиспользовать в других проектах.

## API

| Метод | Путь | Назначение |
|---|---|---|
| `GET` | `/api/v1/health` | Проверка доступности сервера |

## Проверки

Запустить unit-тесты:

```bash
go test ./...
```

Интеграционный тест PostgreSQL использует отдельную временную базу. Запустите тестовый сервис и тест:

```bash
make test-db-up
TEST_DATABASE_URL='postgres://investor_test:investor_test@localhost:5433/myinvestor_test?sslmode=disable' make test-integration
make test-db-down
```

Тест проверяет подключение и ping. Он откажется работать с базой, имя которой не оканчивается на `_test`.

## Запуск backend для разработки

Все команды выполняйте из корневой папки репозитория.

### 1. Установить инструменты

```bash
go mod download
go install github.com/air-verse/air@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3
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

### 2. Настроить и запустить PostgreSQL

Создайте в корне `.env` (он игнорируется Git) и задайте параметры PostgreSQL и URL подключения приложения. Пример для локальных значений по умолчанию:

```dotenv
POSTGRES_DB=myinvestor
POSTGRES_USER=investor
POSTGRES_PASSWORD=investor_dev
DATABASE_URL=postgres://investor:investor_dev@localhost:5432/myinvestor?sslmode=disable
```

Эти значения предназначены только для локальной разработки. Не используйте их для внешних сред.

Запустите Docker Desktop (Windows/macOS) либо Docker Engine с Compose plugin (Linux), затем выполните:

```bash
docker compose up -d
```

Compose запустит PostgreSQL. Убедитесь, что она готова, командой `docker compose ps`.

### 3. Применить и проверить миграции

Миграции не запускаются автоматически при старте API. Применяйте их отдельно перед запуском:
Make-команды берут `DATABASE_URL` из окружения или из строки `DATABASE_URL=...` в `.env`.

```bash
make migrate-up
make migrate-version
```

Для отката последней миграции:

```bash
make migrate-down
```

Начальная миграция фиксирует baseline и не создаёт доменные таблицы; пользовательские таблицы должны появиться в миграции задачи, которая вводит соответствующую модель.

### 4. Запустить backend через Air

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

`Dockerfile` собирает отдельный образ API; описанный выше Compose-сценарий запускает PostgreSQL, а backend запускается локально через Air.

## Участие в разработке

Правила работы с ветками, Pull Request и структурой проекта описаны в [CONTRIBUTING.md](./CONTRIBUTING.md).
