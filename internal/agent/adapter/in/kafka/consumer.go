package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Consumer struct {
	reader       *kafka.Consumer
	deserializer *avrov2.Deserializer
	service      *service.Filter
	retriesCount int
}

func NewConsumer(cfg Config, service *service.Filter) (*Consumer, error) {
	reader, err := buildKafkaConsumer(cfg)
	if err != nil {
		return nil, err
	}
	deser, err := buildAvroDeserializer(cfg)
	if err != nil {
		slog.Error("failed to build avro deserializer:", "error", err)
		return nil, fmt.Errorf("failed to build avro deserializer: %w", err)
	}

	return &Consumer{
		reader:       reader,
		deserializer: deser,
		service:      service,
		retriesCount: cfg.MaxRetries,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	slog.Info("agent: kafka consumer started")
	pollTimeout := 1000

	for {
		select {
		case <-ctx.Done():
			slog.Info("agent: kafka consumer stopped")
			return nil
		default:
		}
		ev := c.reader.Poll(pollTimeout)
		if ev == nil {
			continue
		}

		commitReady := false
		switch e := ev.(type) {
		case *kafka.Message:
			slog.Info("Received message", "message",
				string(e.Value), "topic", *e.TopicPartition.Topic)

			err := c.handleMessage(ctx, *e)

			switch {
			case err == nil:
				commitReady = true
			case errors.Is(err, model.ErrInvalidRequest):
				slog.Error("Invalid message", "err", err)
			default:
				err = c.retryHandleMessage(ctx, *e, c.retriesCount)
				if err == nil {
					commitReady = true
				}
			}

			if commitReady {
				_, err = c.reader.CommitMessage(e)
				if err != nil {
					slog.Error("error committing message", "error", err)
				}
			}
		case kafka.Error:
			slog.Error("kafka message error", "error", e)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	// ID := getHeader(msg, "id")

	var err error
	msgType := getHeader(msg, "message-type")

	switch msgType {
	case "link-update":
		update := model.LinkUpdate{}
		err = c.deserializer.DeserializeInto(*msg.TopicPartition.Topic, msg.Value, &update)
		if err != nil {
			return model.ErrInvalidRequest
		}

		err = c.service.HandleUpdate(ctx, update)
		if err != nil {
			slog.Error("kafka consumer: error sending update", "update", update, "error", err)
			return fmt.Errorf("handling update error: %w", err)
		}
	default:
		slog.Error("kafka consumer: unrecognized message type", "type", msgType)
		err = model.ErrInvalidRequest
	}

	return err
}

func (c *Consumer) retryHandleMessage(ctx context.Context, msg kafka.Message, retries int) error {
	var err error

	for i := range retries {
		slog.Info("retrying handling message...", "retry number", i+1)
		err = c.handleMessage(ctx, msg)
		if err == nil {
			return nil
		}
		if errors.Is(err, model.ErrInvalidRequest) {
			break
		}
	}

	return err
}

func getHeader(msg kafka.Message, key string) string {
	for _, h := range msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}
