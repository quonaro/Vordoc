# Vordoc

Vordoc is a documentation generator/viewer: Markdown pages from the `content` folder are served by a Go backend, while the UI is an embedded Nuxt SPA.

## Table of contents

- [Stack and architecture](#stack-and-architecture)
- [Content folder structure](#content-folder-structure)
- [Environment variables](#environment-variables)
- [Local development with Lota](#local-development-with-lota)
- [Docker build and run](#docker-build-and-run)
- [Useful commands](#useful-commands)

## Stack and architecture

- **Backend:** Go 1.26 (`module vordoc`), routing framework — `go-chi/chi/v5`.
- **Frontend:** Nuxt 3, static generation (`pnpm run generate`).
- **Task runner:** [Lota](https://github.com/quonaro/Lota).
- **Content:** Markdown files, YAML configurations, JSON UI strings.

### How the application is built

1. The frontend is generated into static files in `frontend/.output/public`.
2. The static files are copied to `internal/adapter/http/dist`, which the backend embeds.
3. The Go build produces a single `./vordoc` binary.

The same pipeline is used inside Docker — see `Dockerfile` (`frontend-builder` → `go-builder` → `alpine`).

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

## Content folder structure

`content` is a mountable volume with user content. In Docker it is mounted read-only:

```@/home/quonaro/CascadeProjects/my/Vordoc/docker-compose.yml:9-11
volumes:
  - ./content:/app/content:ro
```

### Content root

| File/folder    | Purpose                                                                                                  |
| -------------- | -------------------------------------------------------------------------------------------------------- |
| `config.yaml`  | Global site configuration: header, logo, theme, favicon.                                                   |
| `text.json`    | UI translations. Recursively overrides the embedded `internal/adapter/content/default_text.json`; only changed keys are needed. |
| `logotype.svg` | Default logotype.                                                                                        |
| Subfolders     | Each subfolder is a separate documentation (`doc`).                                                      |

### Documentation subfolder

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

| File          | Purpose                                                                           |
| ------------- | --------------------------------------------------------------------------------- |
| `config.yaml` | Documentation `title`, `header`, `access`, and `password_hash` settings.          |
| `access.yaml` | Alternative/additional access rules file (may duplicate `config.yaml`).           |
| `index.md`    | Documentation home page.                                                          |
| `*.md`        | Other pages.                                                                      |
| `public/`     | Documentation static resources (images, fonts, etc.).                             |

### Access control

- By default, documentation is public.
- `access: password` + `password_hash` enables password protection.
- Rules are inherited down the tree: a child page can use the nearest ancestor's hash.
- `access: none` or `access: public` resets inheritance.

### UI configuration

`content/config.yaml` supports, for example:

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

## Environment variables

The application reads `.env` automatically (`godotenv.Load()`). In production, set variables through the real environment.

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

| Variable                | Default               | Description                                                                                       |
| ----------------------- | --------------------- | ------------------------------------------------------------------------------------------------- |
| `VORDOC_PORT`           | `12300`               | Backend HTTP server port.                                                                         |
| `VORDOC_CONTENT`        | `./content`           | Path to the user content root.                                                                    |
| `VORDOC_LOG_LEVEL`      | `info`                | Log level (`debug`, `info`, `warn`, `error`).                                                     |
| `VORDOC_LOG_TYPE`       | `pretty`              | Log format (`pretty`, `json`, `text`).                                                            |
| `VORDOC_SHUTDOWN_GRACE` | `10s`                 | Graceful shutdown timeout.                                                                        |
| `VORDOC_PAGE_SECRET`    | generated randomly    | Secret for cookie-protected pages. **Set explicitly**, otherwise sessions reset on restart.        |

## Local development with Lota

For development, prefer using [Lota](https://github.com/quonaro/Lota).

```bash
# Install Lota (once)
go install github.com/quonaro/lota@latest
```

### Commands

```bash
# Run backend (air) and frontend (pnpm dev) in parallel
lota dev

# Build the single ./vordoc binary
lota build

# Build + run
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

- Backend is available at `http://localhost:12300`.
- Frontend in dev mode is at `http://localhost:12301`.
- Lota kills occupied ports before starting.

## Docker build and run

The project includes a ready-to-use `Dockerfile` and `docker-compose.yml`.

```bash
docker compose up --build --force-recreate
```

### What happens inside Docker

1. `frontend-builder` (Node 22 alpine) installs dependencies via pnpm and generates the SPA.
2. The generated static files are copied to `internal/adapter/http/dist`.
3. `go-builder` (Go 1.26 alpine) builds the `vordoc` binary.
4. `runtime` (Alpine 3.21) runs the binary.

Container:

- exposes port `12300`;
- reads `.env`;
- mounts `./content` to `/app/content` read-only.

### Production

- **Be sure** to replace `VORDOC_PAGE_SECRET`.
- Use an external volume or bind-mount for `content`.
- When running without Docker, after `lota build` run `./vordoc run`.

## Useful commands

```bash
# Show version
./vordoc version

# Generate a bcrypt password hash for config.yaml
./vordoc pass "your-secret"
# or: ./vordoc pass password="your-secret"

# Run without Lota
go run ./cmd/vordoc run

# Build the binary without Lota
go build -ldflags="-s -w -X main.version=$(git rev-parse --short HEAD)" -o vordoc ./cmd/vordoc
```

---

License: [MIT](LICENSE) © 2026 quonaro.
