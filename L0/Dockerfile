FROM golang:1.24-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

RUN --mount=type=cache,target=/go/pkg/mod go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG DATE
ENV CGO_ENABLED=0 GOOS=linux

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache \
    go build -trimpath -ldflags "-s -w -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o /out/api ./cmd/service && \
    go build -trimpath -ldflags "-s -w -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o /out/importer ./cmd/importer


FROM alpine:3.20 AS service
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata curl socat

COPY --from=builder /out/service /usr/local/bin/service
COPY web/ /app/web/

RUN printf '%s\n' \
  '#!/bin/sh' \
  'set -e' \
  '' \
  '# DB proxy: localhost:<DSN_PORT> -> postgres:5432' \
  'DB_PORT="${DSN_PORT:-5432}"' \
  'socat TCP-LISTEN:${DB_PORT},bind=127.0.0.1,fork,reuseaddr TCP:postgres:5432 & ' \
  '' \
  '# Kafka proxy: localhost:<KAFKA_PORT> -> kafka:29092 (внутренний listener брокера)' \
  'K_PORT="${KAFKA_PORT:-9092}"' \
  'socat TCP-LISTEN:${K_PORT},bind=127.0.0.1,fork,reuseaddr TCP:kafka:29092 & ' \
  '' \
  '# HTTP: если адрес localhost:*, раскрываем на 0.0.0.0:* (только внутри контейнера)' \
  'ADDR="${HTTP_ADDR:-localhost:8081}"' \
  'ADDR="${ADDR#localhost:}"; export HTTP_ADDR="0.0.0.0:${ADDR}"' \
  '' \
  'exec /usr/local/bin/api' \
  > /usr/local/bin/entrypoint-api.sh && chmod +x /usr/local/bin/entrypoint-api.sh

EXPOSE 8081
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=10 \
  CMD curl -fsS "http://127.0.0.1:8081/healthz" || exit 1

ENTRYPOINT ["/usr/local/bin/entrypoint-service.sh"]


FROM alpine:3.20 AS importer
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata socat
COPY --from=builder /out/importer /usr/local/bin/importer

RUN printf '%s\n' \
  '#!/bin/sh' \
  'set -e' \
  'DB_PORT="${DSN_PORT:-5432}"' \
  'socat TCP-LISTEN:${DB_PORT},bind=127.0.0.1,fork,reuseaddr TCP:postgres:5432 & ' \
  'K_PORT="${KAFKA_PORT:-9092}"' \
  'socat TCP-LISTEN:${K_PORT},bind=127.0.0.1,fork,reuseaddr TCP:kafka:29092 & ' \
  'exec /usr/local/bin/importer "$@"' \
  > /usr/local/bin/entrypoint-importer.sh && chmod +x /usr/local/bin/entrypoint-importer.sh

ENTRYPOINT ["/usr/local/bin/entrypoint-importer.sh"]


FROM alpine:3.20 AS migrator
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata socat

COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations/ /app/migrations/

RUN printf '%s\n' \
  '#!/bin/sh' \
  'set -e' \
  '' \
  '# 1) Local TCP proxy so localhost in DSN keeps working' \
  'DB_PORT="${DSN_PORT:-5432}"' \
  'socat TCP-LISTEN:${DB_PORT},bind=127.0.0.1,fork,reuseaddr TCP:postgres:5432 & ' \
  '' \
  '# 2) Build DSN if not provided' \
  'if [ -n "${DATABASE_URL}" ]; then' \
  '  dsn="${DATABASE_URL}"' \
  'else' \
  '  host="${DSN_HOST:-localhost}"' \
  '  port="${DSN_PORT:-5432}"' \
  '  user="${DSN_USER:-postgres}"' \
  '  pass="${DSN_PASSWORD:-}"' \
  '  db="${DSN_NAME:-postgres}"' \
  '  if [ -n "$pass" ]; then' \
  '    dsn="postgres://${user}:${pass}@${host}:${port}/${db}?sslmode=disable"' \
  '  else' \
  '    dsn="postgres://${user}@${host}:${port}/${db}?sslmode=disable"' \
  '  fi' \
  'fi' \
  '' \
  'GOOSE_DIR="${GOOSE_DIR:-/app/migrations}"' \
  'GOOSE_TABLE="${GOOSE_TABLE:-goose_db_version}"' \
  '' \
  '# 3) Allow overriding command via args (e.g. status, down, redo)' \
  'if [ "$#" -eq 0 ]; then' \
  '  set -- up' \
  'fi' \
  '' \
  'echo "goose -dir ${GOOSE_DIR} postgres \"${dsn}\" $@"' \
  'exec goose -dir "${GOOSE_DIR}" -table "${GOOSE_TABLE}" postgres "${dsn}" "$@"' \
  > /usr/local/bin/entrypoint-migrate.sh && chmod +x /usr/local/bin/entrypoint-migrate.sh

ENTRYPOINT ["/usr/local/bin/entrypoint-migrate.sh"]