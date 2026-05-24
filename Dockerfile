FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS go
WORKDIR /src
RUN apk add --no-cache build-base git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/siphongear ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata wget \
 && adduser -D -u 1000 siphon
COPY --from=go /out/siphongear /app/siphongear
COPY config.yaml.example /app/config.yaml.example
RUN mkdir -p /app/data && chown -R siphon:siphon /app
USER siphon
ENV SIPHON_SERVER__HOST=0.0.0.0 \
    SIPHON_SERVER__PORT=7080 \
    SIPHON_DATABASE__DSN=/app/data/siphongear.db \
    SIPHON_LOG__LEVEL=info
VOLUME ["/app/data"]
EXPOSE 7080
ENTRYPOINT ["/app/siphongear"]
