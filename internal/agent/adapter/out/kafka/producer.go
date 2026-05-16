package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Producer struct {
	writer       *kafka.Producer
	serializer   *avrov2.Serializer
	updatesTopic string
}

func NewProducer(cfg Config) (*Producer, error) {
	writer := buildKafkaProducer(cfg)
	ser, err := buildAvroSerializer(cfg)
	if err != nil {
		slog.Error("error creating kafka serializer", "err", err)
		return nil, fmt.Errorf("error creating kafka serializer: %w", err)
	}

	return &Producer{writer: writer, serializer: ser, updatesTopic: cfg.UpdatesTopic}, nil
}

func (p *Producer) SendUpdate(ctx context.Context, update model.LinkUpdate) error {
	msg, err := p.BuildUpdateMessage(update)
	if err != nil {
		slog.Error("Failed to build update message:", "error", err)
		return fmt.Errorf("building update message: %w", err)
	}

	slog.Info("sending update to kafka consumer", "message", update)

	if err = p.SendMessage(ctx, msg); err != nil {
		slog.Error("failed to write message in kafka producer", "link", update.Link, "error", err)
		return fmt.Errorf("sending message in kafka producer: %w", err)
	}

	slog.Info("finished sending update")
	return nil
}

func (p *Producer) SendMessage(ctx context.Context, msg *kafka.Message) error {
	slog.Info("sending message to kafka consumer", "message", msg.Value)

	deliveryChan := make(chan kafka.Event, 1)

	err := p.writer.Produce(msg, deliveryChan)
	if err != nil {
		return fmt.Errorf("cannot send update in kafka producer: %w", err)
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", ctx.Err())
	case e := <-deliveryChan:
		m, ok := e.(*kafka.Message)
		if !ok {
			return fmt.Errorf("cannot convert message in kafka producer: %v", e)
		}
		if m == nil {
			return errors.New("nil kafka delivery event")
		}

		if m.TopicPartition.Error != nil {
			slog.Error("delivery failed", "error", m.TopicPartition.Error)
			return fmt.Errorf("delivery failed: %w", m.TopicPartition.Error)
		}
		slog.Info("Delivered message to topic", "topic",
			*m.TopicPartition.Topic, "partition", m.TopicPartition.Partition, "offset", m.TopicPartition.Offset)
	}

	return nil
}

func (p *Producer) BuildUpdateMessage(update model.LinkUpdate) (*kafka.Message, error) {
	payload, err := p.serializer.Serialize(p.updatesTopic, &update)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload: %w", err)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.updatesTopic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(update.Link),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "source", Value: []byte("scrapper-producer")},
			{Key: "version", Value: []byte("1.0")},
			{Key: "message-type", Value: []byte("link-update")},
			{Key: "id", Value: []byte(update.ID)},
		},
	}
	return msg, nil
}
