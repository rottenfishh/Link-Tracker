package in

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	in "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in/grpc"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

func TestServerCorrectAndIncorrectTrack(t *testing.T) {
	publisher := service.NewUpdatePublisher()
	server := in.NewBotServiceServer(publisher)
	go func() {
		err := server.RunServer("8079", "8080")
		if err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	update := model.LinkUpdate{
		Id:          1,
		Link:        "bubblegum.com",
		Description: "test requests",
		TgChatIds:   make([]int64, 0),
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
