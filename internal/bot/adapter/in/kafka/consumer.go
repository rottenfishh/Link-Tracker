package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Consumer struct {
	reader       *kafka.Consumer
	deserializer *avrov2.Deserializer
	dlqWriter    *kafka.Producer
	dlqTopic     string
	publisher    *service.UpdatePublisher
	idemptRepo   service.ProcessedEventsRepository
	retriesCount int
}

func NewConsumer(reader *kafka.Consumer, deser *avrov2.Deserializer, dlqWriter *kafka.Producer, dlqTopic string, publisher *service.UpdatePublisher,
	retries int, idemptRepo service.ProcessedEventsRepository) *Consumer {
	return &Consumer{
		reader:       reader,
		deserializer: deser,
		dlqWriter:    dlqWriter,
		dlqTopic:     dlqTopic,
		publisher:    publisher,
		retriesCount: retries,
		idemptRepo:   idemptRepo,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	slog.Info("kafka consumer started")
	pollTimeout := 1000

	for {
		select {
		case <-ctx.Done():
			slog.Info("kafka consumer stopped")
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
				err = c.sendToDLQ(ctx, *e, err)
				if err == nil {
					commitReady = true
				}
			default:
				err = c.retryHandleWithDLQ(ctx, *e, c.retriesCount)
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

func (c *Consumer) GetUpdates() chan service.UpdateJob {
	return c.publisher.GetUpdatesChan()
}

func (c *Consumer) GetReports() chan service.ReportJob {
	return c.publisher.GetReportsChan()
}

// попытка в идемпотентность...
func (c *Consumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	ID := getHeader(msg, "id")
	ok, err := c.idemptRepo.Register(ctx, ID)

	if err != nil {
		slog.Error("error registering message ID", "id", ID, "error", err)
		return fmt.Errorf("error registering message %s: %w", ID, err)
	}
	if !ok {
		return nil
	}

	msgType := getHeader(msg, "message-type")

	switch msgType {
	case "link-update":
		err = c.processUpdate(msg)
		if err != nil {
			slog.Error(" kafka consumer: error processing update", "error", err)
		}
	case "report":
		err = c.processReport(msg)
		if err != nil {
			slog.Error("kafka consumer: error processing report", "error", err)
		}
	default:
		slog.Error("kafka consumer: unrecognized message type", "type", msgType)
		err = model.ErrInvalidRequest
	}

	if err != nil {
		delErr := c.idemptRepo.Delete(ctx, ID)
		if delErr != nil {
			slog.Error(" kafka consumer: error deleting processed event entry", "error", err)
		}
		return err
	}
	return nil
}

func (c *Consumer) retryHandleWithDLQ(ctx context.Context, msg kafka.Message, retries int) error {
	var err error

	for i := range retries {
		slog.Info("retrying to send message...", "retry number", i+1)
		err = c.handleMessage(ctx, msg)
		if err == nil {
			return nil
		}
		if errors.Is(err, model.ErrInvalidRequest) {
			break
		}
	}
	if err == nil {
		return nil
	}

	if err = c.sendToDLQ(ctx, msg, err); err != nil {
		slog.Error("error sending message to DLQ", "error", err)
		return err
	}

	return nil
}

func (c *Consumer) processUpdate(msg kafka.Message) error {
	update := model.LinkUpdate{}
	err := c.deserializer.DeserializeInto(*msg.TopicPartition.Topic, msg.Value, &update)
	if err != nil {
		return model.ErrInvalidRequest
	}

	err = c.publisher.PublishUpdate(update)
	if err != nil {
		slog.Error("kafka consumer: error publishing update", "update", update, "error", err)
		return fmt.Errorf("kafka consumer: error publishing update: %w", err)
	}

	return nil
}

func (c *Consumer) processReport(msg kafka.Message) error {
	report := model.Report{}

	err := c.deserializer.DeserializeInto(*msg.TopicPartition.Topic, msg.Value, &report)
	if err != nil {
		return model.ErrInvalidRequest
	}

	err = c.publisher.PublishReport(report)
	if err != nil {
		slog.Error(" kafka consumer: error processing report", "report", report, "error", err)
		return fmt.Errorf("kafka consumer: error processing report: %w", err)
	}

	return nil
}

func (c *Consumer) sendToDLQ(ctx context.Context, msg kafka.Message, err error) error {
	dlqMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &c.dlqTopic,
			Partition: kafka.PartitionAny,
		},
		Key:   msg.Key,
		Value: msg.Value,
		Headers: []kafka.Header{
			{Key: "source", Value: []byte("scrapper-producer")},
			{Key: "version", Value: []byte("1.0")},
			{Key: "message-type", Value: []byte(getHeader(msg, "message-type"))},
			{Key: "original-topic", Value: []byte(*msg.TopicPartition.Topic)},
			{Key: "original-partition", Value: []byte(strconv.Itoa(int(msg.TopicPartition.Partition)))},
			{Key: "original-offset", Value: []byte(msg.TopicPartition.Offset.String())},
			{Key: "error", Value: []byte(err.Error())},
		},
	}
	slog.Info("sending failed message to DLQ", "message", dlqMsg.Value)

	deliveryChan := make(chan kafka.Event, 1)

	if err = c.dlqWriter.Produce(dlqMsg, deliveryChan); err != nil {
		slog.Error("failed to write message to DLQ", "error", err)
		return fmt.Errorf("failed to write message to DLQ: %w", err)
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

func getHeader(msg kafka.Message, key string) string {
	for _, h := range msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}
