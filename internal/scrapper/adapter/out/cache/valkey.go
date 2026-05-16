package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Config struct {
	Addrs              []string      `config:"addr"`
	Password           string        `config:"password"`
	User               string        `config:"user"`
	Timeout            time.Duration `config:"timeout"`
	Expire             time.Duration `config:"expire"`
	UseClientSideCache bool          `config:"use_client_side_cache"`
}

type ValkeyClient struct {
	client     valkey.Client
	expiration time.Duration
	clientSide bool
}

func NewValkeyClient(cfg Config) (*ValkeyClient, error) {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress:      cfg.Addrs,
		Password:         cfg.Password,
		ConnWriteTimeout: cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("error building valkey client: %w", err)
	}

	return &ValkeyClient{
		client:     client,
		expiration: cfg.Expire,
		clientSide: cfg.UseClientSideCache,
	}, nil
}

func (v *ValkeyClient) Set(ctx context.Context, chatID int64, links []model.Link) error {
	slog.Debug("set cache", "chatID", chatID, "links", links)
	payload, err := json.Marshal(links)
	if err != nil {
		return model.ErrInvalidRequest
	}

	err = v.client.Do(ctx, v.client.B().Set().Key(strconv.FormatInt(chatID, 10)).
		Value(string(payload)).Ex(v.expiration).Build()).Error()
	if err != nil {
		return fmt.Errorf("error setting links in cache: %w", err)
	}

	return nil
}

func (v *ValkeyClient) Get(ctx context.Context, chatID int64) ([]model.Link, error) {
	if v.clientSide {
		return v.GetSideCache(ctx, chatID)
	}
	return v.GetSimple(ctx, chatID)
}

func (v *ValkeyClient) GetSimple(ctx context.Context, chatID int64) ([]model.Link, error) {
	slog.Debug("get cache", "chatID", chatID)
	resp := v.client.Do(ctx, v.client.B().Get().Key(strconv.FormatInt(chatID, 10)).Build())
	val, err := resp.ToString()
	if err != nil {
		if errors.Is(err, valkey.Nil) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("error getting links in cache: %w", err)
	}

	var links []model.Link
	err = json.Unmarshal([]byte(val), &links)
	if err != nil {
		return nil, model.ErrInvalidRequest
	}

	return links, nil
}

func (v *ValkeyClient) GetSideCache(ctx context.Context, chatID int64) ([]model.Link, error) {
	slog.Debug("get cache", "chatID", chatID)
	resp := v.client.DoCache(ctx, v.client.B().Get().Key(strconv.FormatInt(chatID, 10)).Cache(), v.expiration)
	val, err := resp.ToString()
	if err != nil {
		if errors.Is(err, valkey.Nil) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("error getting links in cache: %w", err)
	}

	var links []model.Link
	err = json.Unmarshal([]byte(val), &links)
	if err != nil {
		return nil, model.ErrInvalidRequest
	}

	return links, nil
}

func (v *ValkeyClient) Invalidate(ctx context.Context, chatID int64) error {
	err := v.client.Do(ctx, v.client.B().Del().Key(strconv.FormatInt(chatID, 10)).Build()).Error()
	if err != nil {
		return fmt.Errorf("error invalidating links in cache: %w", err)
	}
	return nil
}
