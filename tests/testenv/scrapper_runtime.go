package testenv

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	apiclients "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/api_clients"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
	cache "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/cache"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
	scrapper "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/app"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

type TestEnv struct {
	DB              *sql.DB
	DSN             string
	ChatService     *service.ChatService
	TrackingService *service.LinkTrackingService
	Repos           *out.Repositories
	Trackers        []service.LinkUpdater
}

type KafkaSchemaRegistryEnv struct {
	NetworkName       string
	KafkaContainer    *tckafka.KafkaContainer
	SchemaRegistry    testcontainers.Container
	Brokers           []string
	SchemaRegistryURL string
}

func SetUpTestEnv(t *testing.T, accessType string, client *http.Client) *TestEnv {
	t.Helper()

	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:latest",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "link-tracker",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/link-tracker?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	require.Eventually(t, func() bool {
		return db.PingContext(ctx) == nil
	}, 10*time.Second, 200*time.Millisecond) //nolint:mnd //its ok

	err = scrapper.RunMigrations(dsn)
	require.NoError(t, err)

	repos, err := scrapper.BuildRepos(accessType, db)
	require.NoError(t, err)

	trackers := []service.LinkUpdater{
		apiclients.NewGithubClient("mock-token", client),
		apiclients.NewStackOverflowClient("mock-token", client),
	}
	linkResolver := service.NewLinkResolver(trackers)
	txManager := repository.NewTransactionManager(db)

	//TODO: set up real cache
	cacheClient := cache.NoOpCache{}
	chatService := service.NewChatService(*repos, linkResolver, txManager, cacheClient)
	trackingService := service.NewLinkTrackingService(repos, txManager)

	return &TestEnv{
		DB:              db,
		DSN:             dsn,
		ChatService:     chatService,
		Repos:           repos,
		Trackers:        trackers,
		TrackingService: trackingService,
	}
}

func SetupKafkaWithSchemaRegistry(ctx context.Context) (*KafkaSchemaRegistryEnv, func(), error) {
	kafkaNetwork, err := network.New(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create testcontainers network: %w", err)
	}

	cleanup := func() {
		_ = kafkaNetwork.Remove(ctx)
	}

	kafkaContainer, err := tckafka.Run(
		ctx,
		"confluentinc/confluent-local:7.5.0",
		tckafka.WithClusterID("test-cluster"),
		network.WithNetwork([]string{"kafka"}, kafkaNetwork),
		testcontainers.WithEnv(map[string]string{
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE": "true",
		}),
	)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to start kafka container: %w", err)
	}

	cleanup = func() {
		_ = kafkaContainer.Terminate(ctx)
		_ = kafkaNetwork.Remove(ctx)
	}

	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("error getting brokers: %w", err)
	}

	schemaRegistry, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "confluentinc/cp-schema-registry:7.5.0",
			ExposedPorts: []string{"8081/tcp"},
			Networks:     []string{kafkaNetwork.Name},
			NetworkAliases: map[string][]string{
				kafkaNetwork.Name: {"schema-registry"},
			},
			Env: map[string]string{
				"SCHEMA_REGISTRY_HOST_NAME":                    "schema-registry",
				"SCHEMA_REGISTRY_LISTENERS":                    "http://0.0.0.0:8081",
				"SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS": "PLAINTEXT://kafka:9092",
			},
			WaitingFor: wait.ForHTTP("/subjects").
				WithPort("8081/tcp").
				WithStartupTimeout(60 * time.Second), //nolint:mnd // its ok
		},
		Started: true,
	})
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to start kafka schema registry container: %w", err)
	}

	oldCleanup := cleanup
	cleanup = func() {
		_ = schemaRegistry.Terminate(ctx)
		oldCleanup()
	}

	host, err := schemaRegistry.Host(ctx)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("could not get schema registry container host: %w", err)
	}

	port, err := schemaRegistry.MappedPort(ctx, "8081/tcp")
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("error mapping testContainers tcp port to host: %w", err)
	}

	return &KafkaSchemaRegistryEnv{
		NetworkName:       kafkaNetwork.Name,
		KafkaContainer:    kafkaContainer,
		SchemaRegistry:    schemaRegistry,
		Brokers:           brokers,
		SchemaRegistryURL: fmt.Sprintf("http://%s:%s", host, port.Port()),
	}, cleanup, nil
}

// func BuildTestScrapper(env *TestEnv, t *testing.T) *scrapper.App {
//	cfg := scrapper.ScrapperAppConfig{
//		GithubToken:    "mock-token",
//		StackOFToken:   "mock-token",
//		BotURL:         "http://localhost:8080",
//		Port:           "8081",
//		GatewayPort:    "8082",
//		DatabaseConfig: scrapper.DatabaseConfig{},
//		LinkTrackingConfig: scrapper.LinkTrackingConfig{
//			BatchSize:         2,
//			NThreads:          2,
//			SchedulerInterval: 10,
//			LinksOldAge:       30,
//		},
//	}
//	linkRuntime, err := scrapper.BuildLinkRuntime(context.Background(), cfg, env.Trackers, env.TrackingService)
//	require.NoError(t, err)
//
//	grpcServer := grpc.NewScrapperServer(env.ChatService, cfg.Port, cfg.GatewayPort)
//
//	repos, err := scrapper.BuildRepos("sql", env.DB)
//	require.NoError(t, err)
//
//	app := scrapper.NewApp(cfg, cfg.DatabaseConfig, linkRuntime, nil, grpcServer, *repos)
//	return app
//}
