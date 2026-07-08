# syntax=docker/dockerfile:1

# -----------------------------------------------------------------------------
# Frontend build stage
# -----------------------------------------------------------------------------
FROM node:22-alpine AS frontend-builder

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app/frontend

COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/ ./
RUN pnpm run generate

# Move generated SPA into the Go embed path
RUN rm -rf /app/internal/adapter/http/dist \
    && mkdir -p /app/internal/adapter/http/dist \
    && cp -R /app/frontend/.output/public/. /app/internal/adapter/http/dist/

# -----------------------------------------------------------------------------
# Go build stage
# -----------------------------------------------------------------------------
FROM golang:1.26-alpine AS go-builder

WORKDIR /app

# Install git for version injection and build dependencies
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY shared/ ./shared/

# Copy generated frontend dist into the Go embed path
COPY --from=frontend-builder /app/internal/adapter/http/dist ./internal/adapter/http/dist

ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /bin/vordoc \
    ./cmd/vordoc

# -----------------------------------------------------------------------------
# Runtime stage
# -----------------------------------------------------------------------------
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=go-builder /bin/vordoc /usr/local/bin/vordoc

# Default content directory; mount a volume here in production
RUN mkdir -p /app/content

ENV VORDOC_CONTENT=/app/content
ENV VORDOC_PORT=12300

EXPOSE 12300

ENTRYPOINT ["vordoc"]
CMD ["run"]
