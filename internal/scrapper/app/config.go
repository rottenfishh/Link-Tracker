package app

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/byrnedo/typesafe-config/parse"
	"github.com/joho/godotenv"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/adapter/httpclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/middleware"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/cache"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/kafka"
)

// TODO: build app
type ScrapperAppConfig struct {
	GithubToken        string                     `config:"github_token"`
	StackOFToken       string                     `config:"stack_of_token"`
	BotURL             string                     `config:"bot_url"`
	HTTPReqConfig      httpclient.Config          `config:"http_req_config"`
	RateLimitConfig    middleware.RateLimitConfig `config:"rate_limit"`
	Port               string                     `config:"port"`
	GatewayPort        string                     `config:"gateway_port"`
	DatabaseConfig     DatabaseConfig             `config:"database"`
	LinkTrackingConfig LinkTrackingConfig         `config:"link_tracking"`
	NotificationType   string                     `config:"notification_type"`
	KafkaConfig        kafka.Config               `config:"kafka"`
	CacheConfig        cache.Config
}

type DatabaseConfig struct {
	AccessType   string `config:"access_type"`
	DatabaseName string `config:"postgres_db"`
	Username     string `config:"postgres_user"`
	Password     string `config:"postgres_password"`
	Host         string `config:"host"`
	Port         string `config:"port"`
}

type LinkTrackingConfig struct {
	BatchSize         int `config:"batch_size"`
	NThreads          int `config:"n_threads"`
	SchedulerInterval int `config:"scheduler_interval"`
	LinksOldAge       int `config:"links_old_age"`
}

func (c *DatabaseConfig) DSN() string {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		c.Username, c.Password, c.Host, c.Port, c.DatabaseName)
	return dsn
}

func LoadEnv() error {
	err := godotenv.Load(".env.scrapper")
	if err != nil {
		return fmt.Errorf("loading .env: %w", err)
	}
	return nil
}

func LoadConfig() (*ScrapperAppConfig, error) {
	err := LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("loading environment: %w", err)
	}
	var cfg ScrapperAppConfig
	tree, err := parse.ParseFile("scrapper.conf")
	if err != nil {
		return nil, fmt.Errorf("parsing scrapper config: %w", err)
	}
	parse.Populate(&cfg, tree.GetConfig(), "root")

	ghToken := os.Getenv("GITHUB_API_TOKEN")
	if ghToken == "" {
		return nil, errors.New("GITHUB_API_TOKEN environment variable not set")
	}

	stackOFToken := os.Getenv("STACK_OF_TOKEN")
	if stackOFToken == "" {
		return nil, errors.New("STACK_OF_TOKEN environment variable not set")
	}
	cfg.StackOFToken = stackOFToken
	cfg.GithubToken = ghToken

	cfg.DatabaseConfig.DatabaseName = os.Getenv("POSTGRES_DB")
	cfg.DatabaseConfig.Username = os.Getenv("POSTGRES_USER")
	cfg.DatabaseConfig.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.DatabaseConfig.Host = os.Getenv("DATABASE_HOST")
	cfg.DatabaseConfig.Port = os.Getenv("DATABASE_PORT")

	slog.Info(cfg.DatabaseConfig.DSN())
	return &cfg, nil
}
