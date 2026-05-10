FROM golang:1.25.0-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.19

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

COPY --from=builder /out/server /app/server
COPY --from=builder /src/migrations /app/migrations

ENV HTTP_ADDR=:8080

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD sh -ec 'port="${HTTP_ADDR##*:}"; wget -q -O /dev/null "http://127.0.0.1:${port}/healthz" || exit 1'

CMD ["/app/server"]
