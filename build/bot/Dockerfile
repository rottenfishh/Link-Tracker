FROM golang:1.25-alpine AS builder

WORKDIR /bot


RUN apk add --no-cache gcc musl-dev pkgconfig

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -tags musl -o bot ./cmd/bot

FROM alpine:3.20

WORKDIR /scrapper

COPY --from=builder /bot/bot .
COPY --from=builder /bot/database/bot ./database/bot
COPY --from=builder /bot/bot.conf ./bot.conf
COPY --from=builder /bot/.env.bot ./.env.bot

EXPOSE 8080

CMD ["./bot"]