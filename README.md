# Vordoc

Vordoc — это генератор/вьюер документации: markdown-страницы из папки `content` отдаются через Go-бэкенд, а интерфейс представляет собой встроенный Nuxt SPA.

## Содержание

- [Стек и архитектура](#стек-и-архитектура)
- [Структура папки `content`](#структура-папки-content)
- [Переменные окружения](#переменные-окружения)
- [Локальная разработка через Lota](#локальная-разработка-через-lota)
- [Docker-сборка и запуск](#docker-сборка-и-запуск)
- [Полезные команды](#полезные-команды)

## Стек и архитектура

- **Бэкенд:** Go 1.26 (`module vordoc`), фреймворк маршрутизации — `go-chi/chi/v5`.
- **Фронтенд:** Nuxt 3, статическая генерация (`pnpm run generate`).
- **Сборщик задач:** [Lota](https://github.com/quonaro/Lota).
- **Контент:** Markdown-файлы, YAML-конфигурации, JSON-строки интерфейса.

### Как собирается приложение

1. Фронтенд собирается в статику в `frontend/.output/public`.
2. Статика копируется в `internal/adapter/http/dist`, куда бэкенд подкладывает её через `embed`.
3. Go-билд формирует единый бинарник `./vordoc`.

Этот же пайплайн используется и в Docker — см. `Dockerfile` (`frontend-builder` → `go-builder` → `alpine`).

```@/home/quonaro/CascadeProjects/my/Vordoc/Dockerfile:6-48
FROM node:22-alpine AS frontend-builder
...
RUN pnpm run generate
...
COPY --from=frontend-builder /app/internal/adapter/http/dist ./internal/adapter/http/dist
...
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /bin/vordoc \
    ./cmd/vordoc
```

## Структура папки `content`

`content` — это монтируемый том с пользовательским контентом. В Docker он подключается read-only:

```@/home/quonaro/CascadeProjects/my/Vordoc/docker-compose.yml:9-11
volumes:
  - ./content:/app/content:ro
```

### Корень `content`

| Файл/папка     | Назначение                                                                                                                                  |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `config.yaml`  | Глобальная конфигурация сайта: заголовок, логотип, тема, favicon.                                                                           |
| `text.json`    | Переводы UI. Переопределяет встроенный `internal/adapter/content/default_text.json` рекурсивно: достаточно указать только изменённые ключи. |
| `logotype.svg` | Логотип по умолчанию.                                                                                                                       |
| Подпапки       | Каждая подпапка — отдельная документация (`doc`).                                                                                           |

### Подпапка документации

```
content/
├── config.yaml
├── text.json
├── logotype.svg
├── public/
│   ├── config.yaml
│   └── guide/
│       └── getting-started.md
└── admin/
    ├── config.yaml
    ├── access.yaml
    ├── index.md
    ├── settings.md
    └── public/
```

| Файл          | Назначение                                                                           |
| ------------- | ------------------------------------------------------------------------------------ |
| `config.yaml` | `title` документации, настройки `header`, `access` и `password_hash`.                |
| `access.yaml` | Альтернативный/дополнительный файл правил доступа (может дублировать `config.yaml`). |
| `index.md`    | Главная страница документации.                                                       |
| `*.md`        | Остальные страницы.                                                                  |
| `public/`     | Статические ресурсы документации (изображения, шрифты и т.д.).                       |

### Права доступа

- По умолчанию документация публичная.
- `access: password` + `password_hash` включает авторизацию по паролю.
- Правила наследуются вниз по дереву: дочерняя страница может использовать хеш ближайшего предка.
- `access: none` или `access: public` сбрасывает наследование.

### Настройка интерфейса

`content/config.yaml` поддерживает, например:

```yaml
root:
  enable: true
  title: "Vordoc"

favicon: "favicon.ico"

header:
  enable: true
  elements: ["logo", "search", "theme-switch"]
  title: "Vordoc"
  logo:
    path: "logotype.svg"
    size: 40
  font:
    name: "FabergeDigital"
    size: 42

theme:
  default: "system"
  accent-color: "#3b82f6"
```

## Переменные окружения

Приложение читает `.env` автоматически (`godotenv.Load()`). В продакшене переменные задаются через реальное окружение.

```@/home/quonaro/CascadeProjects/my/Vordoc/.env:1-10
# Vordoc runtime configuration
...
VORDOC_PORT=12300
VORDOC_CONTENT=./content
VORDOC_LOG_LEVEL=info
VORDOC_LOG_TYPE=pretty
VORDOC_SHUTDOWN_GRACE=10s
VORDOC_PAGE_SECRET=CHANGE_ME
```

| Переменная              | По умолчанию          | Описание                                                                                        |
| ----------------------- | --------------------- | ----------------------------------------------------------------------------------------------- |
| `VORDOC_PORT`           | `12300`               | Порт HTTP-сервера бэкенда.                                                                      |
| `VORDOC_CONTENT`        | `./content`           | Путь к корню пользовательского контента.                                                        |
| `VORDOC_LOG_LEVEL`      | `info`                | Уровень логирования (`debug`, `info`, `warn`, `error`).                                         |
| `VORDOC_LOG_TYPE`       | `pretty`              | Формат логов (`pretty`, `json`, `text`).                                                        |
| `VORDOC_SHUTDOWN_GRACE` | `10s`                 | Таймаут graceful shutdown.                                                                      |
| `VORDOC_PAGE_SECRET`    | генерируется случайно | Секрет для cookie защищённых страниц. **Задайте явно**, иначе сессии сбросятся при перезапуске. |

## Локальная разработка через Lota

Для разработки предпочтительно использовать [Lota](https://github.com/quonaro/Lota).

```bash
# Установка Lota (один раз)
go install github.com/quonaro/lota@latest
```

### Команды

```bash
# Запуск бэкенда (air) и фронтенда (pnpm dev) параллельно
lota dev

# Сборка единого бинарника ./vordoc
lota build

# Сборка + запуск
lota run
```

```@/home/quonaro/CascadeProjects/my/Vordoc/lota.yaml:28-62
echo "==> Installing frontend dependencies..."
...
pnpm run generate
...
cp -R frontend/.output/public/. internal/adapter/http/dist/
...
go build -ldflags="-s -w -X main.version=$COMMIT" -o vordoc ./cmd/vordoc
```

- Бэкенд доступен на `http://localhost:12300`.
- Фронтенд в dev-режиме на `http://localhost:12301`.
- Lota сама прибивает занятые порты перед стартом.

## Docker-сборка и запуск

Проект содержит готовый `Dockerfile` и `docker-compose.yml`.

```bash
docker compose up --build --force-recreate
```

### Что происходит внутри Docker

1. `frontend-builder` (Node 22 alpine) — устанавливает зависимости через pnpm и генерирует SPA.
2. Сгенерированная статика копируется в `internal/adapter/http/dist`.
3. `go-builder` (Go 1.26 alpine) — собирает бинарник `vordoc`.
4. `runtime` (Alpine 3.21) — запускает бинарник.

Контейнер:

- экспонирует порт `12300`;
- читает `.env`;
- монтирует `./content` в `/app/content` только для чтения.

### Продакшен

- **Обязательно** замените `VORDOC_PAGE_SECRET`.
- Используйте внешний том или bind-mount для `content`.
- При запуске без Docker после `lota build` запустите `./vordoc run`.

## Полезные команды

```bash
# Показать версию
./vordoc version

# Сгенерировать bcrypt-хеш пароля для config.yaml
./vordoc pass "your-secret"
# или: ./vordoc pass password="your-secret"

# Запуск вне Lota
go run ./cmd/vordoc run

# Сборка бинарника без Lota
go build -ldflags="-s -w -X main.version=$(git rev-parse --short HEAD)" -o vordoc ./cmd/vordoc
```

---

Лицензия: [MIT](LICENSE) © 2026 quonaro.
