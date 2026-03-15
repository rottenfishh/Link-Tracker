package app

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/byrnedo/typesafe-config/parse"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/app"
)

// TODO: build app
type ScrapperAppConfig struct {
	GithubToken    string          `config:"github_token"`
	StackOFToken   string          `config:"stack_of_token"`
	BotUrl         string          `config:"bot_url"`
	Port           string          `config:"port"`
	DatabaseConfig *DatabaseConfig `config:"database"`
}

type DatabaseConfig struct {
	AccessType  string `config:"access_type"`
	DatabaseUrl string `config:"database_url"`
	Name        string `config:"name"`
	Password    string `config:"password"`
	Host        string `config:"host"`
	Port        string `config:"port"`
}

func (c *DatabaseConfig) DSN() string {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		c.Name, c.Password, c.Host, c.Port, c.DatabaseUrl)
	return dsn
}

func LoadConfig() (*ScrapperAppConfig, error) {
	err := app.LoadEnv()
	if err != nil {
		return nil, err
	}
	var cfg ScrapperAppConfig
	tree, err := parse.ParseFile("scrapper.conf")
	if err != nil {
		return nil, err
	}
	parse.Populate(&cfg, tree.GetConfig(), "root")

	ghToken := os.Getenv("GITHUB_API_TOKEN")
	if ghToken == "" {
		return nil, errors.New("GITHUB_ACCESS_TOKEN environment variable not set")
	}

	stackOFToken := os.Getenv("STACK_OF_TOKEN")
	if stackOFToken == "" {
		return nil, errors.New("STACK_OF_TOKEN environment variable not set")
	}
	cfg.StackOFToken = stackOFToken
	cfg.GithubToken = ghToken

	if cfg.DatabaseConfig == nil {
		cfg.DatabaseConfig = &DatabaseConfig{}
	}

	cfg.DatabaseConfig.DatabaseUrl = os.Getenv("DATABASE_URL")
	cfg.DatabaseConfig.Name = os.Getenv("DATABASE_NAME")
	cfg.DatabaseConfig.Password = os.Getenv("DATABASE_PASSWORD")
	cfg.DatabaseConfig.Host = os.Getenv("DATABASE_HOST")
	cfg.DatabaseConfig.Port = os.Getenv("DATABASE_PORT")

	slog.Info(cfg.DatabaseConfig.DSN())
	return &cfg, nil
}
