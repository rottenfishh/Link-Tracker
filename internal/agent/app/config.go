package app

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/byrnedo/typesafe-config/parse"
	"github.com/joho/godotenv"
	inkafka "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/adapter/in/kafka"
	outkafka "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/adapter/out/kafka"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/service"
)

type Config struct {
	ConsumerConfig inkafka.Config       `config:"consumer"`
	ProducerConfig outkafka.Config      `config:"producer"`
	FilterConfig   service.FilterConfig `config:"filter"`
	HuggingToken   string               `config:"hf_token"`
}

func LoadEnv() error {
	err := godotenv.Load(".env.agent")
	if err != nil {
		return fmt.Errorf("loading .env: %w", err)
	}
	return nil
}

func LoadConfig() (*Config, error) {
	err := LoadEnv()
	if err != nil {
		slog.Error("error loading .env file", err)
		return nil, err
	}

	cfg := &Config{}
	tree, err := parse.ParseFile("./agent.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read agent.conf: %w", err)
	}
	cfg.HuggingToken = os.Getenv("hf_token")
	parse.Populate(cfg, tree.GetConfig(), "root")
	slog.Info("loaded config", "config", cfg)
	return cfg, nil
}
