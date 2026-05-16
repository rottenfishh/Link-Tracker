package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

const (
	model_url = "https://router.huggingface.co/hf-inference/models/facebook/bart-large-cnn"
)

type HuggingSummarizer struct {
	httpclient *http.Client
	token      string
}

func NewHuggingSummarizer(token string) *HuggingSummarizer {
	return &HuggingSummarizer{httpclient: &http.Client{}, token: token}
}

type HuggingRequest struct {
	Inputs string `json:"inputs"`
}

type HuggingResponse struct {
	SummaryText string `json:"summary_text"`
}

func (s *HuggingSummarizer) Summarize(ctx context.Context, description string, _ int) (string, error) {
	request := HuggingRequest{
		Inputs: description,
	}

	body, err := json.Marshal(request)
	if err != nil {
		slog.Error("failed to marshal request")
	}
	slog.Info("summarizing request: %s", body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, model_url, bytes.NewBuffer(body))

	if err != nil {
		return description, fmt.Errorf("hugging face client: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpclient.Do(req)
	if err != nil {
		return description, fmt.Errorf("hugging face client: failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Error("hugging face client: failed to execute request", "status", resp.Status)
		return description, fmt.Errorf("hugging face client: failed to execute request: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	var response []HuggingResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return description, fmt.Errorf("hugging face client: failed to decode response: %w", err)
	}
	slog.Info("hugging face response", "resp", response[0].SummaryText)
	return response[0].SummaryText, nil
}
