package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Topics struct {
	UpdatesTopic string `json:"updates_topic"`
	ReportsTopic string `json:"reports_topic"`
}

type Producer struct {
	writer     *kafka.Producer
	serializer *avrov2.Serializer
	topics     Topics
}

func NewProducer(writer *kafka.Producer, ser *avrov2.Serializer, topics Topics) *Producer {
	return &Producer{writer: writer, serializer: ser, topics: topics}
}

func (p *Producer) SendUpdate(ctx context.Context, update model.LinkUpdate) error {
	msg, err := p.BuildUpdateMessage(update)
	if err != nil {
		slog.Error("Failed to build update message:", "error", err)
		return err
	}

	slog.Info("sending update to kafka consumer", "message", update)

	if err = p.SendMessage(ctx, msg); err != nil {
		slog.Error("failed to write message in kafka producer", "link", update.Link, "error", err)
		return err
	}

	slog.Info("finished sending update")
	return nil
}

func (p *Producer) SendReport(ctx context.Context, report model.Report) error {
	msg, err := p.BuildReportMessage(report)
	if err != nil {
		slog.Error("Failed to build report message:", "error", err)
		return err
	}

	slog.Info("sending report to kafka consumer", "message", msg)
	if err = p.SendMessage(ctx, msg); err != nil {
		slog.Error("failed to write message in kafka producer", "chatID", report.ChatID, "error", err)
		return err
	}

	slog.Info("Finished sending message")
	return nil
}

func (p *Producer) BuildUpdateMessage(update model.LinkUpdate) (*kafka.Message, error) {
	payload, err := p.serializer.Serialize(p.topics.UpdatesTopic, &update)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload: %w", err)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topics.UpdatesTopic,
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

func (p *Producer) BuildReportMessage(report model.Report) (*kafka.Message, error) {
	payload, err := p.serializer.Serialize(p.topics.ReportsTopic, &report)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload: %w", err)
	}

	msg := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topics.ReportsTopic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(strconv.FormatInt(report.ChatID, 10)),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "source", Value: []byte("scrapper-producer")},
			{Key: "version", Value: []byte("1.0")},
			{Key: "message-type", Value: []byte("report")},
			{Key: "id", Value: []byte(report.ID)},
		},
	}

	return &msg, nil
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
