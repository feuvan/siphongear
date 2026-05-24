FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS go
WORKDIR /src
RUN apk add --no-cache build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/siphongear ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go /out/siphongear /app/siphongear
COPY config.yaml.example /app/config.yaml.example
ENV SIPHON_DATABASE__DSN=/app/data/siphongear.db
EXPOSE 7080
ENTRYPOINT ["/app/siphongear"]
