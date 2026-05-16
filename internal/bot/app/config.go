//nolint:revive // keep AppConfig naming consistent with the rest of the app package API
package app

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/byrnedo/typesafe-config/parse"
	"github.com/joho/godotenv"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in/kafka"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/adapter/httpclient"
)

type TelegramConfig struct {
	Token string `config:"token"`
	Debug bool   `config:"debug,default=false"`
}

type AppConfig struct {
	Port             int               `config:"port, default=8080"`
	GatewayPort      int               `config:"gateway_port, default=8088"`
	HTTPReqConfig    httpclient.Config `config:"http_req_config"`
	ScrapperURL      string            `config:"scrapper_url"`
	Telegram         TelegramConfig    `config:"telegram"`
	NotificationType string            `config:"notification_type"`
	NThreads         int               `config:"n_threads" default:"1"`
	KafkaConfig      kafka.Config      `config:"kafka"`
	DatabaseConfig   DatabaseConfig    `config:"database"`
}

type DatabaseConfig struct {
	AccessType string `config:"access_type"`
	Name       string `config:"name"`
	User       string `config:"user"`
	Password   string `config:"password"`
	Host       string `config:"host"`
	Port       string `config:"port"`
}

func (c *DatabaseConfig) DSN() string {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Name)
	return dsn
}

func LoadEnv() error {
	err := godotenv.Load(".env.bot")
	if err != nil {
		return fmt.Errorf("loading .env: %w", err)
	}
	return nil
}

func InitLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func LoadConfig() (*AppConfig, error) {
	err := LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("error loading env: %w", err)
	}

	cfg := &AppConfig{}
	tree, err := parse.ParseFile("./bot.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read bot.conf: %w", err)
	}

	parse.Populate(cfg, tree.GetConfig(), "root")

	tgToken := os.Getenv("TG_API_TOKEN")
	cfg.Telegram.Token = tgToken

	cfg.DatabaseConfig.Name = os.Getenv("POSTGRES_DB")
	cfg.DatabaseConfig.User = os.Getenv("POSTGRES_USER")
	cfg.DatabaseConfig.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.DatabaseConfig.Host = os.Getenv("DATABASE_HOST")
	cfg.DatabaseConfig.Port = os.Getenv("DATABASE_PORT")

	return cfg, nil
}
