# ── Build stage ──
FROM golang:alpine AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /movie-api .

# ── Runtime stage ──
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /movie-api .
COPY frontend/ ./frontend/

EXPOSE 8081

ENV DATABASE_URL=""

CMD ["./movie-api"]
