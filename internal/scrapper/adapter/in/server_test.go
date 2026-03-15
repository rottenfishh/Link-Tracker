package in

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
)

const baseURL = "http://127.0.0.1:8082"

//func TestMain(m *testing.M) {
//	chatService := service.NewChatService()
//	server := grpc.NewScrapperServer(chatService)
//
//	go func() {
//		err := server.RunServer("8085", "8082")
//		if err != nil {
//			slog.Error("Error starting server:", "error", err)
//		}
//	}()
//	time.Sleep(2000 * time.Millisecond)
//	m.Run()
//}

func TestAddAndGetLink(t *testing.T) {
	resp, err := postJSON(baseURL+"/tg-chat/1", nil)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	addLinkRequest := dto.AddLinkRequest{
		Link: "bubblegum.com",
		Tags: nil,
	}
	resp, err = postJSON(baseURL+"/links/1", addLinkRequest)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	resp, err = http.Get(baseURL + "/links/1")
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
	if resp.Body == nil {
		t.Errorf("expected non-nil body")
	}
}

func TestAddAndDeleteLink(t *testing.T) {
	response, err := postJSON(baseURL+"/tg-chat/1", nil)
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", response.StatusCode)
	}

	addReq := dto.AddLinkRequest{
		Link: "bubblegum.com",
		Tags: nil,
	}
	resp, _ := postJSON(baseURL+"/links/1", addReq)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("add link failed")
	}

	delReq := dto.DeleteLinkRequest{
		Link: "bubblegum.com",
	}
	resp, _ = deleteJSON(baseURL+"/links/1", delReq)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("delete link failed")
	}

	resp, _ = http.Get(baseURL + "/links/1")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("get links failed")
	}
}

func TestDeleteLinkFromNonExistingChat(t *testing.T) {
	response, err := postJSON(baseURL+"/tg-chat/1", nil)
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", response.StatusCode)
	}

	addReq := dto.AddLinkRequest{
		Link: "bubblegum.com",
		Tags: nil,
	}
	resp, _ := postJSON(baseURL+"/links/1", addReq)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("add link failed")
	}

	delReq := dto.DeleteLinkRequest{
		Link: "bubblegum.com",
	}

	resp, _ = deleteJSON(baseURL+"/links/999", delReq)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 status")
	}

	resp, _ = http.Get(baseURL + "/links/1")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("get links failed")
	}
	if resp.Body == nil {
		t.Errorf("expected non-nil body")
	}
}

// 3.4 Добавление ссылки в несуществующий чат
func TestAddLinkToNonExistingChat(t *testing.T) {
	response, err := postJSON(baseURL+"/tg-chat/1", nil)
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", response.StatusCode)
	}

	addReq := dto.AddLinkRequest{
		Link: "bubblegum.com",
		Tags: nil,
	}
	resp, _ := postJSON(baseURL+"/links/2", addReq)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 status")
	}
}

// 3.5 Работа с удалённым чатом
func TestDeletedChatBehaviour(t *testing.T) {
	_, err := postJSON(baseURL+"/tg-chat/1", nil)
	if err != nil {
		t.Errorf("creating chat error: %v", err)
	}

	resp, _ := deleteJSON(baseURL+"/tg-chat/1", nil)
	if resp.StatusCode != http.StatusOK {
		t.Error("chat delete failed")
	}

	addReq := dto.AddLinkRequest{
		Link: "bubblegum.com",
		Tags: nil,
	}
	resp, _ = postJSON(baseURL+"/links/1", addReq)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected error adding link to deleted chat")
	}
}

func TestDeleteNonExistingChat(t *testing.T) {
	resp, _ := deleteJSON(baseURL+"/tg-chat/5", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 got %d", resp.StatusCode)
	}
}

func postJSON(url string, body any) (*http.Response, error) {
	var buf *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(data)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	return http.Post(url, "application/json", buf)
}

func deleteJSON(url string, body any) (*http.Response, error) {
	var buf *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(data)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(http.MethodDelete, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}
