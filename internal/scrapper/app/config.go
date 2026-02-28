package app

import (
	"errors"
	"os"

	"github.com/byrnedo/typesafe-config/parse"
)

// TODO: build app
type AppConfig struct {
	GithubToken  string `config:"github_token"`
	StackOFToken string `config:"stack_of_token"`
	BotUrl       string `config:"bot_url"`
	Port         string `config:"port"`
}

// TODO: type-safety and proper config loading
func LoadConfig() (*AppConfig, error) {
	var cfg AppConfig
	tree, err := parse.ParseFile("scrapper.conf")
	if err != nil {
		return nil, err
	}
	parse.Populate(cfg, tree.GetConfig(), "root")

	ghToken := os.Getenv("GITHUB_ACCESS_TOKEN")
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
