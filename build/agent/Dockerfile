FROM golang:1.25-alpine AS builder

WORKDIR /agent


RUN apk update && apk add --no-cache gcc musl-dev pkgconfig

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -tags musl -o agent ./cmd/agent

FROM alpine:3.20

WORKDIR /agent

COPY --from=builder /agent/agent .
COPY --from=builder /agent/agent.conf ./agent.conf
COPY --from=builder /agent/.env.agent ./.env.agent
CMD ["./agent"]