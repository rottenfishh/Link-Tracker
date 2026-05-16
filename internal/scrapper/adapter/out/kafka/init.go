package kafka

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
)

type Config struct {
	SchemaURL         string   `config:"schema_url"`
	Brokers           []string `config:"brokers"`
	UpdatesTopic      string   `config:"updates_topic"`
	ReportsTopic      string   `config:"reports_topic"`
	BatchSize         int      `config:"batch_size"`
	BatchBytes        int      `config:"batch_bytes"`
	BatchTimeoutMs    int      `config:"batch_timeout_ms"`
	RequiredAcks      string   `config:"required_acks"`
	MaxRetries        int      `config:"max_retries"`
	ReplicationFactor int      `config:"replication_factor"`
}

func BuildKafkaProducer(cfg Config) *Producer {
	topics := Topics{
		UpdatesTopic: cfg.UpdatesTopic,
		ReportsTopic: cfg.ReportsTopic,
	}

	brokers := strings.Join(cfg.Brokers, ",")

	producerCfg := &kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"enable.idempotence": true,
		"acks":               cfg.RequiredAcks,
		"retries":            cfg.MaxRetries,

		// Maintain ordering during retries
		"max.in.flight.requests.per.connection": 1,

		// Optional throughput tuning
		"linger.ms":          cfg.BatchTimeoutMs,
		"batch.num.messages": cfg.BatchSize,
		"batch.size":         cfg.BatchBytes,
		"compression.type":   "snappy",
	}

	producer, err := kafka.NewProducer(producerCfg)
	if err != nil {
		slog.Error("error creating kafka producer", "err", err)
		return nil
	}

	ser, err := BuildAvroSerializer(cfg)
	if err != nil {
		slog.Error("error creating kafka serializer", "err", err)
		return nil
	}

	return NewProducer(producer, ser, topics)
}

func BuildAvroSerializer(cfg Config) (*avrov2.Serializer, error) {
	client, err := schemaregistry.NewClient(schemaregistry.NewConfig(cfg.SchemaURL))

	if err != nil {
		slog.Error("Failed to create schema registry client", "error", err)
		os.Exit(1)
	}

	ser, err := avrov2.NewSerializer(client, serde.ValueSerde, avrov2.NewSerializerConfig())

	if err != nil {
		slog.Error("Failed to create serializer", "error", err)
		return nil, fmt.Errorf("error creating avro serializer: %w", err)
	}

	return ser, nil
}
