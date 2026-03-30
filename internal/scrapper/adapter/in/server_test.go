package in

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	grpcin "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/grpc"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

const baseURL = "http://127.0.0.1:8082"

var chatSeq int64 = 1000

func nextChatID() int64 {
	return atomic.AddInt64(&chatSeq, 1)
}

func TestMain(m *testing.M) {
	repo := out.NewInMemoryRepo()

	trackers := []service.LinkUpdater{
		service.NewGithubClient("mock-token"),
		service.NewStackOverflowClient("mock-token"),
	}
	linkResolver := service.NewLinkResolver(trackers)
	chatService := service.NewChatService(repo, linkResolver)

	server := grpcin.NewScrapperServer(chatService)

	go func() {
		_ = server.RunServer("8085", "8082")
	}()

	time.Sleep(2 * time.Second)

	code := m.Run()
	os.Exit(code)
}

func TestAddAndGetLink(t *testing.T) {
	chatID := nextChatID()

	resp, err := postJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	addReq := dto.AddLinkRequest{
		Link: "https://github.com/golang/go",
		Tags: []string{"go", "lang"},
	}

	resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), addReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(fmt.Sprintf("%s/links/%d", baseURL, chatID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var linksResp dto.ListLinksResponse
	err = decodeJSON(resp, &linksResp)
	require.NoError(t, err)
	require.Len(t, linksResp.Links, 1)
	require.Equal(t, "https://github.com/golang/go", linksResp.Links[0].Link)
}

func TestAddAndDeleteLink(t *testing.T) {
	chatID := nextChatID()

	resp, err := postJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	addReq := dto.AddLinkRequest{
		Link: "https://github.com/golang/go",
		Tags: nil,
	}
	resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), addReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	delReq := dto.DeleteLinkRequest{
		Link: "https://github.com/golang/go",
	}
	resp, err = deleteJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), delReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(fmt.Sprintf("%s/links/%d", baseURL, chatID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var linksResp dto.ListLinksResponse
	err = decodeJSON(resp, &linksResp)
	require.NoError(t, err)
	require.Len(t, linksResp.Links, 0)
}

func TestDeleteLinkFromNonExistingChat(t *testing.T) {
	chatID := nextChatID()

	delReq := dto.DeleteLinkRequest{
		Link: "https://github.com/golang/go",
	}

	resp, err := deleteJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), delReq)
	require.NoError(t, err)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestAddLinkToNonExistingChat(t *testing.T) {
	chatID := nextChatID()

	addReq := dto.AddLinkRequest{
		Link: "https://github.com/golang/go",
		Tags: nil,
	}

	resp, err := postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), addReq)
	require.NoError(t, err)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestDeletedChatBehaviour(t *testing.T) {
	chatID := nextChatID()

	resp, err := postJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = deleteJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	addReq := dto.AddLinkRequest{
		Link: "https://github.com/golang/go",
		Tags: nil,
	}

	resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), addReq)
	require.NoError(t, err)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestDeleteNonExistingChat(t *testing.T) {
	chatID := 100

	resp, err := deleteJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAddInvalidLink(t *testing.T) {
	chatID := nextChatID()

	resp, err := postJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	addReq := dto.AddLinkRequest{
		Link: "bubblegum.com",
		Tags: nil,
	}

	resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), addReq)
	require.NoError(t, err)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestGetLinksByTag(t *testing.T) {
	chatID := nextChatID()

	resp, err := postJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), dto.AddLinkRequest{
		Link: "https://github.com/golang/go",
		Tags: []string{"go"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), dto.AddLinkRequest{
		Link: "https://stackoverflow.com/questions/11227809/why-is-processing-a-sorted-array-faster-than-an-unsorted-array",
		Tags: []string{"so"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(fmt.Sprintf("%s/links/%d?tag=go", baseURL, chatID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var linksResp dto.ListLinksResponse
	err = decodeJSON(resp, &linksResp)
	require.NoError(t, err)
	require.Len(t, linksResp.Links, 1)
	require.Equal(t, "https://github.com/golang/go", linksResp.Links[0].Link)
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

func decodeJSON(resp *http.Response, dst any) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, dst)
}
