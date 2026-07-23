FROM golang:1.26-alpine AS builder
LABEL authors="nesanessis"

WORKDIR /app

#Кэшируем зависимости отдельным слоем,
#чтобы пересборка при изменении кода не качала модули заново
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/gateway ./cmd

FROM alpine:latest
LABEL authors="nesanessis"

WORKDIR /app

COPY --from=builder /app/bin/gateway ./gateway
COPY config/config.yml ./config/config.yml
COPY migrations ./migrations

EXPOSE 8080 7070

ENTRYPOINT ["./gateway"]