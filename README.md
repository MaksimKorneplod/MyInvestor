# MyInvestor

MyInvestor — симулятор трейдинга, который позволит пользователям практиковаться в покупке и продаже финансовых активов в условиях, приближённых к реальным.

> Проект находится в разработке. Сейчас доступны проверка состояния API, регистрация школы и получение новостной ленты. API использует PostgreSQL.

## Требования

- Go 1.26.10
- Git
- Docker Desktop (Windows и macOS) или Docker Engine с Compose (Linux)
- `migrate` для обновления схемы базы и GNU Make для команд из `Makefile` — см. [установку инструментов](#1-установить-инструменты).

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

Контракт API описан в [OpenAPI](./docs/openapi.yaml). Источник новостей и правила фильтрации описаны в [документации новостной ленты](./docs/news.md). Регистрация пользователя и авторизация в таблице ниже пока только описаны в контракте.

| Метод | Путь | Назначение |
|---|---|---|
| `GET` | `/api/v1/health` | Проверка доступности сервера |
| `POST` | `/api/v1/schools/register` | [Регистрация школы](./docs/school-registration.md) |
| `GET` | `/api/v1/news` | Новости Интерфакса категорий «Экономика» и «Бизнес» |
| `POST` | `/api/v1/register` | Регистрация пользователя (планируется) |
| `POST` | `/api/v1/login` | Вход и получение bearer-токена (планируется) |
| `POST` | `/api/v1/logout` | Завершение сессии (планируется, требует bearer-токен) |

## Проверки

Обычные тесты можно запустить без PostgreSQL:

```bash
go test ./...
```

Чтобы проверить подключение к PostgreSQL, применение и откат миграции, запустите отдельную тестовую базу. Передайте её адрес в `TEST_DATABASE_URL`; имя базы должно оканчиваться на `_test`.

```bash
make test-db-up
make test-integration TEST_DATABASE_URL='postgres://investor_test:investor_test@localhost:5433/myinvestor_test?sslmode=disable'
make test-db-down
```

Последняя команда останавливает и удаляет тестовый контейнер.

## Запуск backend для разработки

Все команды выполняйте из корневой папки репозитория.

### 1. Установить инструменты

Скачайте зависимости, установите Air для перезапуска API при изменении кода и `migrate` с поддержкой PostgreSQL:

```bash
go mod download
go install github.com/air-verse/air@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3
```

Если терминал не находит `air` или `migrate`, добавьте каталог с инструментами Go в `PATH`.

**Bash (в том числе Git Bash) или Zsh:**

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

**PowerShell:**

```powershell
$env:Path = "$(go env GOPATH)\bin;$env:Path"
```

Эти команды действуют только в текущем терминале. Чтобы настройка сохранялась, добавьте каталог `bin` из `go env GOPATH` в пользовательский `PATH`.

#### Установить Make

На macOS Make входит в инструменты командной строки Xcode (`xcode-select --install`). На Debian и Ubuntu его можно установить командой `sudo apt install make`. На Windows, например, через Chocolatey:

```powershell
choco install make
```

На Windows выполняйте команды `make` в **Git Bash**: `Makefile` использует утилиты, доступные в этой оболочке. Из PowerShell миграции можно запускать [напрямую](#миграции-без-make).

Проверьте, что инструменты доступны в выбранном терминале:

```bash
air -v
migrate -version
make --version
```

Если при подключении `migrate` пишет `unknown driver postgres`, переустановите его командой выше с `-tags 'postgres'`.

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

Миграции — это SQL-файлы, которые обновляют схему базы. API их сам не запускает. После запуска PostgreSQL и после получения новых файлов в `migrations/` выполните:

```bash
make migrate-up
make migrate-version
```

`make migrate-up` применяет все новые миграции, а `make migrate-version` показывает текущую версию. `Makefile` берёт `DATABASE_URL` из окружения или из `.env`. Для другой базы передайте адрес явно: `make migrate-up DATABASE_URL='postgres://...'`.

Каждая миграция состоит из двух файлов с одним номером: `.up.sql` меняет схему, `.down.sql` отменяет это изменение. `migrate` применяет их по порядку и записывает текущую версию в служебную таблицу `schema_migrations`. Первая миграция `000001_init_schema` задаёт начальную версию, а `000002_school_accounts` создаёт таблицы школ, учеников, их профилей и избранного.

#### Добавить миграцию

1. Создайте пару файлов со следующим свободным номером и коротким именем:

   ```bash
   migrate create -ext sql -dir migrations -seq add_example_table
   ```

2. Запишите изменение в `.up.sql`, а его отмену — в `.down.sql`. Включите оба файла в ту же ветку, что и код, который использует новую схему.
3. На отдельной тестовой базе примените миграцию, откатите её и примените снова. `make migrate-down` отменяет **одну** последнюю миграцию и может удалить данные. Перед проверкой убедитесь, что `DATABASE_URL` указывает на тестовую базу.

Не меняйте миграции, которые уже попали в `develop` или `main`: на других базах их номер может быть уже отмечен как применённый. Создайте новую пару файлов. Если в двух ветках заняли один номер, перед слиянием перенумеруйте свою миграцию.

#### Если версия помечена как dirty

Пометка `dirty` означает, что миграция не завершилась и `migrate` остановил дальнейшие обновления. Посмотрите номер через `make migrate-version` и проверьте, какие изменения успели примениться. Приведите схему к последней успешно применённой версии. Только после этого снимите пометку `dirty`, указав номер последней успешно применённой миграции:

```bash
migrate -path ./migrations -database "$DATABASE_URL" force НОМЕР_ВЕРСИИ
make migrate-up
```

Перед запуском задайте `DATABASE_URL` в текущем терминале. Команда `force` меняет запись о версии; она не выполняет и не откатывает SQL.

#### Миграции без Make

В PowerShell можно запустить `migrate` напрямую. Сначала задайте `$env:DATABASE_URL`: используйте значение `DATABASE_URL` из своего `.env`.

```powershell
migrate -path ./migrations -database $env:DATABASE_URL up
migrate -path ./migrations -database $env:DATABASE_URL version
```

Для отката одной миграции замените `up` на `down 1`.

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
