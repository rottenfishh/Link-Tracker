package app

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/byrnedo/typesafe-config/parse"
	"github.com/joho/godotenv"
)

type TelegramConfig struct {
	Token string `config:"token"`
	Debug bool   `config:"debug,default=false"`
}

type AppConfig struct {
	Port        int            `config:"port, default=8080"`
	ScrapperUrl string         `config:"scrapper_url"`
	Telegram    TelegramConfig `config:"telegram"`
}

func LoadEnv() error {
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}
	return nil
}

func InitLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func LoadConfig() (*AppConfig, error) {
	err := LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("error loading env: %v", err)
	}

	cfg := &AppConfig{}
	tree, err := parse.ParseFile("./app.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read /app.conf:" + err.Error())
	}

	parse.Populate(cfg, tree.GetConfig(), "root")

	tgToken := os.Getenv("TG_API_TOKEN")
	cfg.Telegram.Token = tgToken
	return cfg, nil
}
