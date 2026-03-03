package app

import (
	"errors"
	"os"

	"github.com/byrnedo/typesafe-config/parse"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/app"
)

// TODO: build app
type ScrapperAppConfig struct {
	GithubToken  string `config:"github_token"`
	StackOFToken string `config:"stack_of_token"`
	BotUrl       string `config:"bot_url"`
	Port         string `config:"port"`
}

// TODO: type-safety and proper config loading
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
	return &cfg, nil
}
