FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o backend ./cmd/dem

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/backend .

COPY config ./config
COPY frontend-dist ./frontend-dist

COPY *.json ./

EXPOSE 8080

CMD ["./backend"]