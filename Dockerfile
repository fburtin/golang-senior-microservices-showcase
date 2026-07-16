# ---------- Build Stage ----------
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/customer-consumer ./cmd/customer-consumer

# ---------- Runtime Stage ----------
FROM alpine:latest

WORKDIR /app

COPY --from=builder /out/api ./api
COPY --from=builder /out/customer-consumer ./customer-consumer

EXPOSE 8080

CMD ["./api"]