# syntax=docker/dockerfile:1
FROM node:22-bookworm-slim AS frontend-build
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@9.15.0 --activate
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY packages/core/package.json packages/core/package.json
COPY packages/frontend/package.json packages/frontend/package.json
RUN pnpm install --frozen-lockfile
COPY tsconfig*.json ./
COPY packages/core packages/core
COPY packages/frontend packages/frontend
ARG TARO_APP_WS_URL
ENV NODE_ENV=production TARO_APP_WS_URL=${TARO_APP_WS_URL}
RUN pnpm --filter @clocktower/core build && pnpm --filter @clocktower/frontend build:h5

FROM golang:1.25-alpine AS backend-build
WORKDIR /app
COPY packages/backend/go.mod packages/backend/go.sum ./
RUN go mod download
COPY packages/backend ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /clocktower ./cmd/server

FROM alpine:3.22 AS backend
RUN apk add --no-cache ca-certificates && addgroup -S clocktower && adduser -S -G clocktower clocktower
COPY --from=backend-build /clocktower /usr/local/bin/clocktower
USER clocktower
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 CMD wget -q -O /dev/null http://127.0.0.1:8080/health || exit 1
ENTRYPOINT ["clocktower"]

FROM caddy:2-alpine AS web
COPY --from=frontend-build /app/packages/frontend/dist /srv
COPY deploy/Caddyfile /etc/caddy/Caddyfile
