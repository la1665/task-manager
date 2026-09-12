# ---- Build stage ----
FROM golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN go build -o /app/bin/task-manager ./cmd/api

# ---- Final stage ----
FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/bin/task-manager .

EXPOSE 8080

ENTRYPOINT ["./task-manager"]
