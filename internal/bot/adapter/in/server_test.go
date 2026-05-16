package in

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	in "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in/grpc"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

func TestServerCorrectAndIncorrectTrack(t *testing.T) {
	publisher := service.NewUpdatePublisher(100, 100)
	server := in.NewBotServiceServer(publisher, "8079", "8080")
	go func() {
		err := server.Run(context.Background())
		if err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	update := model.LinkUpdate{
		ID:   uuid.New().String(),
		Link: "bubblegum.com",
		Update: model.Update{
			Title:       "Issue",
			Description: "Test issue",
		},
		TgChatIDs: make([]int64, 0),
	}
	body, err := json.Marshal(&update)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post("http://127.0.0.1:8080/updates", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	updateWrong := "buba"
	body, err = json.Marshal(&updateWrong)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = http.Post("http://127.0.0.1:8080/updates", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
