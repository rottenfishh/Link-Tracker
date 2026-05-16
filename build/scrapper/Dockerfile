FROM golang:1.25-alpine AS builder

WORKDIR /scrapper

RUN apk add --no-cache gcc musl-dev pkgconfig

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -tags musl -o  scrapper ./cmd/scrapper

FROM alpine:3.20

WORKDIR /scrapper

COPY --from=builder /scrapper/scrapper .
COPY --from=builder /scrapper/database/scrapper ./database/scrapper
COPY --from=builder /scrapper/scrapper.conf ./scrapper.conf
COPY --from=builder /scrapper/.env.scrapper ./.env.scrapper

EXPOSE 8082

CMD ["./scrapper"]