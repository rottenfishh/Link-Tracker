package kafka

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
)

type Config struct {
	SchemaURL    string   `config:"schema_url"`
	Topics       []string `config:"topics"`
	Brokers      []string `config:"brokers"`
	GroupID      string   `config:"group_id"`
	Offset       string   `config:"offset"`
	MinBytes     int      `config:"min_bytes"`
	MaxBytes     int      `config:"max_bytes"`
	MaxWaitMsecs int      `config:"max_wait_ms"`
	MaxRetries   int      `config:"max_retries"`
}

func buildKafkaConsumer(cfg Config) (*kafka.Consumer, error) {
	brokers := strings.Join(cfg.Brokers, ",")

	slog.Info("kafka config", "config", cfg)
	consumerCfg := &kafka.ConfigMap{
		"bootstrap.servers":             brokers,
		"group.id":                      cfg.GroupID,
		"auto.offset.reset":             cfg.Offset,
		"enable.auto.commit":            false, // Manual offset control
		"enable.auto.offset.store":      false, // Store after success only
		"fetch.min.bytes":               cfg.MinBytes,
		"fetch.max.bytes":               cfg.MaxBytes,
		"fetch.wait.max.ms":             cfg.MaxWaitMsecs,
		"partition.assignment.strategy": "cooperative-sticky",
	}
	reader, err := kafka.NewConsumer(consumerCfg)
	if err != nil {
		slog.Error("failed to create consumer:", "error", err)
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	err = reader.SubscribeTopics(cfg.Topics, nil)
	if err != nil {
		slog.Error("Consumer failed to subscribe to topic", "error", err)
		return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	return reader, nil
}

func buildAvroDeserializer(cfg Config) (*avrov2.Deserializer, error) {
	client, err := schemaregistry.NewClient(schemaregistry.NewConfig(cfg.SchemaURL))
	if err != nil {
		slog.Error("failed to create schema registry:", "error", err)
		return nil, fmt.Errorf("failed to create schema registry: %w", err)
	}

	deser, err := avrov2.NewDeserializer(client, serde.ValueSerde, avrov2.NewDeserializerConfig())

	if err != nil {
		slog.Error("failed to create deserializer", "error", err)
		return nil, fmt.Errorf("failed to create deserializer: %w", err)
	}

	return deser, nil
}
