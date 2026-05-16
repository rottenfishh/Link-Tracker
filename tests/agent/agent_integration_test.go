package agent_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	inkafka "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/adapter/in/kafka"
	outkafka "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/adapter/out/kafka"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/tests/testenv"
)

const (
	rawUpdatesTopic       = "link.raw-updates"
	processedUpdatesTopic = "link.processed-updates"
)

var kafkaEnv *testenv.KafkaSchemaRegistryEnv

func TestMain(m *testing.M) {
	var cleanup func()
	var err error

	kafkaEnv, cleanup, err = testenv.SetupKafkaWithSchemaRegistry(context.Background())
	if err != nil {
		fmt.Printf("failed to set up kafka env: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	cleanup()
	os.Exit(code)
}

func buildTestPublisher(t *testing.T, schemaURL string, brokers []string) *outkafka.Producer {
	t.Helper()

	cfg := outkafka.Config{
		SchemaURL:         schemaURL,
		Brokers:           brokers,
		UpdatesTopic:      rawUpdatesTopic, // публикуем в raw-updates
		BatchSize:         1,
		BatchBytes:        1e6,
		BatchTimeoutMs:    0,
		RequiredAcks:      "all",
		MaxRetries:        3,
		ReplicationFactor: 1,
	}

	p, err := outkafka.NewProducer(cfg)
	require.NoError(t, err, "failed to create test publisher")
	return p
}

func buildAgentFilter(producer service.Producer) *service.Filter {
	cfg := service.FilterConfig{
		StopWords:     []string{"spam"},
		BannedAuthors: []string{"bot"},
		MinLength:     5,
	}
	cfg.Summarization.Threshold = 500

	summarizer := &service.StupidSummarizer{}
	return service.NewFilter(cfg, summarizer, producer)
}

func buildAgentConsumer(
	t *testing.T,
	schemaURL string,
	brokers []string,
	groupID string,
	filter *service.Filter,
) *inkafka.Consumer {
	t.Helper()

	cfg := inkafka.Config{
		SchemaURL:    schemaURL,
		Topics:       []string{rawUpdatesTopic},
		Brokers:      brokers,
		GroupID:      groupID,
		Offset:       "earliest",
		MinBytes:     1,
		MaxBytes:     10e6,
		MaxWaitMsecs: 100,
		MaxRetries:   3,
	}

	c, err := inkafka.NewConsumer(cfg, filter)
	require.NoError(t, err, "failed to create agent consumer")
	return c
}

func buildRawConsumer(
	t *testing.T,
	brokers []string,
	schemaURL string,
	groupID string,
) (*kafka.Consumer, *avrov2.Deserializer) {
	t.Helper()

	brokersStr := strings.Join(brokers, ",")
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokersStr,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	require.NoError(t, err)

	err = c.SubscribeTopics([]string{processedUpdatesTopic}, nil)
	require.NoError(t, err)

	srClient, err := schemaregistry.NewClient(schemaregistry.NewConfig(schemaURL))
	require.NoError(t, err)

	deser, err := avrov2.NewDeserializer(srClient, serde.ValueSerde, avrov2.NewDeserializerConfig())
	require.NoError(t, err)

	return c, deser
}

func consumeProcessedUpdate(
	t *testing.T,
	consumer *kafka.Consumer,
	deser *avrov2.Deserializer,
	timeout time.Duration,
) *model.LinkUpdate {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ev := consumer.Poll(200)
		if ev == nil {
			continue
		}
		msg, ok := ev.(*kafka.Message)
		if !ok {
			continue
		}
		if msg.TopicPartition.Error != nil {
			t.Logf("partition error: %v", msg.TopicPartition.Error)
			continue
		}

		upd := &model.LinkUpdate{}
		err := deser.DeserializeInto(*msg.TopicPartition.Topic, msg.Value, upd)
		require.NoError(t, err, "failed to deserialize processed update")
		return upd
	}
	return nil
}

// ─── TC-1.1: Получение и обработка корректного сообщения ─────────────────────

// TestAgentConsumer_ValidMessage_ProcessedAndForwarded проверяет, что:
//   - агент получает корректное Avro-сообщение из link.raw-updates
//   - сообщение передаётся через фильтр (проходит все условия)
//   - итоговый LinkUpdate появляется в link.processed-updates
func TestAgentConsumer_ValidMessage_ProcessedAndForwarded(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Продюсер в processed-updates (выход агента)
	agentProducerCfg := outkafka.Config{
		SchemaURL:         kafkaEnv.SchemaRegistryURL,
		Brokers:           kafkaEnv.Brokers,
		UpdatesTopic:      processedUpdatesTopic,
		BatchSize:         1,
		BatchBytes:        1e6,
		BatchTimeoutMs:    0,
		RequiredAcks:      "all",
		MaxRetries:        3,
		ReplicationFactor: 1,
	}
	agentProducer, err := outkafka.NewProducer(agentProducerCfg)
	require.NoError(t, err)

	filter := buildAgentFilter(agentProducer)

	groupID := "agent-test-valid-" + uuid.NewString()
	consumer := buildAgentConsumer(t, kafkaEnv.SchemaRegistryURL, kafkaEnv.Brokers, groupID, filter)

	// Запускаем агент
	go func() {
		if runErr := consumer.Run(ctx); runErr != nil {
			t.Logf("agent consumer stopped: %v", runErr)
		}
	}()

	// Публикуем корректное сообщение в link.raw-updates
	publisher := buildTestPublisher(t, kafkaEnv.SchemaRegistryURL, kafkaEnv.Brokers)
	update := model.LinkUpdate{
		ID:   uuid.NewString(),
		Link: "https://github.com/test/repo",
		Update: model.Update{
			Title:       "New PR",
			Author:      "alice",
			Description: "This is a valid update description without any banned words",
		},
		TgChatIDs: []int64{42},
	}

	err = publisher.SendUpdate(ctx, update)
	require.NoError(t, err)

	// Читаем из link.processed-updates
	verifyGroupID := "verify-" + uuid.NewString()
	rawConsumer, deser := buildRawConsumer(t, kafkaEnv.Brokers, kafkaEnv.SchemaRegistryURL, verifyGroupID)
	defer rawConsumer.Close()

	result := consumeProcessedUpdate(t, rawConsumer, deser, 30*time.Second)

	require.NotNil(t, result, "expected a message in link.processed-updates, got none")
	require.Equal(t, update.Link, result.Link)
	require.Equal(t, update.Update.Author, result.Update.Author)
}

// ─── TC-1.2: Обработка некорректного сообщения без падения ───────────────────

// TestAgentConsumer_InvalidMessage_DoesNotCrash проверяет, что:
//   - агент не падает при получении сообщения с некорректной структурой
//   - ошибка десериализации обрабатывается (логируется, не паникует)
//   - сервис продолжает работу после ошибки
func TestAgentConsumer_InvalidMessage_DoesNotCrash(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	agentProducerCfg := outkafka.Config{
		SchemaURL:         kafkaEnv.SchemaRegistryURL,
		Brokers:           kafkaEnv.Brokers,
		UpdatesTopic:      processedUpdatesTopic,
		BatchSize:         1,
		BatchBytes:        1e6,
		BatchTimeoutMs:    0,
		RequiredAcks:      "all",
		MaxRetries:        3,
		ReplicationFactor: 1,
	}
	agentProducer, err := outkafka.NewProducer(agentProducerCfg)
	require.NoError(t, err)

	filter := buildAgentFilter(agentProducer)

	groupID := "agent-test-invalid-" + uuid.NewString()
	consumer := buildAgentConsumer(t, kafkaEnv.SchemaRegistryURL, kafkaEnv.Brokers, groupID, filter)

	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		_ = consumer.Run(ctx)
	}()

	// Публикуем сырые (невалидные) байты напрямую через низкоуровневый продюсер
	brokersStr := strings.Join(kafkaEnv.Brokers, ",")
	rawProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokersStr,
	})
	require.NoError(t, err)
	defer rawProducer.Close()

	deliveryChan := make(chan kafka.Event, 1)
	topic := rawUpdatesTopic
	err = rawProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		// Некорректные байты — не Avro-формат
		Value: []byte(`{"this": "is not avro", "broken": true}`),
		Headers: []kafka.Header{
			{Key: "message-type", Value: []byte("link-update")},
		},
	}, deliveryChan)
	require.NoError(t, err)

	select {
	case e := <-deliveryChan:
		m, ok := e.(*kafka.Message)
		require.True(t, ok)
		if m.TopicPartition.Error != nil {
			t.Logf("delivery error (expected): %v", m.TopicPartition.Error)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for delivery confirmation")
	}

	// Даём агенту время обработать некорректное сообщение.
	// Главная проверка: горутина с Run не завершилась с паникой.
	time.Sleep(3 * time.Second)

	// Отменяем контекст — консьюмер должен штатно завершиться
	cancel()

	select {
	case <-consumerDone:
		// успех: консьюмер завершился без паники
	case <-time.After(10 * time.Second):
		t.Fatal("agent consumer did not stop after context cancellation")
	}
}

// ─── TC-1.2 (доп): Сообщение с неизвестным message-type не вызывает падения ──

// TestAgentConsumer_UnknownMessageType_DoesNotCrash проверяет, что агент
// корректно игнорирует сообщения с неизвестным заголовком message-type.
func TestAgentConsumer_UnknownMessageType_DoesNotCrash(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	agentProducerCfg := outkafka.Config{
		SchemaURL:         kafkaEnv.SchemaRegistryURL,
		Brokers:           kafkaEnv.Brokers,
		UpdatesTopic:      processedUpdatesTopic,
		BatchSize:         1,
		BatchBytes:        1e6,
		BatchTimeoutMs:    0,
		RequiredAcks:      "all",
		MaxRetries:        3,
		ReplicationFactor: 1,
	}
	agentProducer, err := outkafka.NewProducer(agentProducerCfg)
	require.NoError(t, err)

	filter := buildAgentFilter(agentProducer)

	groupID := "agent-test-unknown-type-" + uuid.NewString()
	consumer := buildAgentConsumer(t, kafkaEnv.SchemaRegistryURL, kafkaEnv.Brokers, groupID, filter)

	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		_ = consumer.Run(ctx)
	}()

	brokersStr := strings.Join(kafkaEnv.Brokers, ",")
	rawProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokersStr,
	})
	require.NoError(t, err)
	defer rawProducer.Close()

	topic := rawUpdatesTopic
	deliveryChan := make(chan kafka.Event, 1)
	err = rawProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(`some payload`),
		Headers: []kafka.Header{
			{Key: "message-type", Value: []byte("unknown-type")},
		},
	}, deliveryChan)
	require.NoError(t, err)
	<-deliveryChan

	time.Sleep(3 * time.Second)
	cancel()

	select {
	case <-consumerDone:
		// успех
	case <-time.After(10 * time.Second):
		t.Fatal("agent consumer did not stop after context cancellation")
	}
}
