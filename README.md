# Vordoc

Self-hosted Markdown documentation viewer. Content lives in the filesystem. No database. No user authentication.

## Features

- **Filesystem-based content**: All docs are plain Markdown files with YAML frontmatter
- **Multiple documentations**: Each folder under `content/` is a separate doc collection
- **Themes**: Customizable CSS variables and optional layout overrides
- **Password protection**: Protect specific pages with bcrypt passwords (no auth system)
- **Dark mode**: Automatic light/dark theme switching

## Tech Stack

- **Backend**: Go 1.22 + chi
- **Frontend**: Nuxt 3 + Vue 3 + Tailwind CSS + shadcn/ui
- **Package manager**: pnpm (frontend only)

## Prerequisites

- Go 1.22+
- pnpm

## Development

Two terminals needed:

**Terminal 1 — Backend:**

```bash
# First time only
go mod tidy

# Run with hot reload (requires `air`)
air

# Or run directly
go run ./cmd/vordoc
```

Backend runs on `http://localhost:8080`.

**Terminal 2 — Frontend:**

```bash
cd frontend
pnpm install
pnpm dev
```

Frontend dev server runs on `http://localhost:3000` and proxies `/api` and `/themes` to the Go backend.

## Content Structure

```
content/
  public/
    config.yaml          # title, theme, sidebar order
    access.yaml          # access rules (public / password)
    index.md             # landing page
    guide/
      intro.md
      setup.md
```

- `config.yaml`: `title`, `description`, `theme`, `sidebar`
- `access.yaml`: per-page/section access rules with bcrypt `password_hash`
- `.md` files: YAML frontmatter (`title`, `order`) + markdown body

## Access Levels

- `public` — anyone can view
- `password` — requires entering the correct page password

Password verification sets a signed cookie valid for 24 hours.

## Demo Content

Two demo docs included:

- **`content/public/`** — public doc with guide pages
- **`content/admin/`** — doc with a password-protected `settings` page
  - Password: `admin123`

## API

- `GET /api/v1` — list all docs
- `GET /api/v1/:doc` — doc metadata + sidebar
- `GET /api/v1/:doc/:page` — page content (respects access rules)
- `POST /api/v1/:doc/:page` — verify page password
- `GET /api/config/public` — public runtime config
- `/themes/:name/vars.css` — theme CSS variables (static)

## License

MIT
