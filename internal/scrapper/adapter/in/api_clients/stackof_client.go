package apiclients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/adapter/httpclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

const stackOverflowPreviewMaxLength = 200

type StackOverflowClient struct {
	client httpclient.HTTPClient
	token  string
}

func NewStackOverflowClient(token string, client httpclient.HTTPClient) *StackOverflowClient {
	return &StackOverflowClient{client: client, token: token}
}

func (c *StackOverflowClient) GetTitle(apiURL string) (string, error) {
	parts := strings.Split(apiURL, "/")
	return parts[len(parts)-1], nil
}

func (c *StackOverflowClient) GetDomain() string {
	return "stackof"
}

func (c *StackOverflowClient) FormatLink(link string) (string, error) {
	if !strings.Contains(link, "stackoverflow") {
		return "", errors.New("invalid link")
	}

	u, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("parsing stackoverflow link: %w", err)
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) < 3 || parts[1] != "questions" {
		return "", errors.New("invalid stackoverflow question url")
	}

	questionID := parts[2]

	apiURL := fmt.Sprintf(
		`https://api.stackexchange.com/2.3/questions/%s`,
		questionID,
	)
	return apiURL, nil
}

func (c *StackOverflowClient) GetUpdates(link *model.Link) ([]model.Update, error) {
	timeModified := time.Now()
	var updates []model.Update

	answers, err := c.GetAnswerUpdates(link.FormattedLink, link.LastUpdated)
	if err != nil {
		slog.Error("Error getting answers", "error", err)
	}

	for _, answer := range answers.Items {
		upd, formatErr := c.FormatAnswer(timeModified, answer)
		if formatErr != nil {
			return nil, fmt.Errorf("error parsing answer: %w", formatErr)
		}

		updates = append(updates, upd)
		slog.Debug("stack overflow message update", "updated", upd)
	}

	comments, err := c.GetCommentUpdates(link.FormattedLink, link.LastUpdated)
	if err != nil {
		slog.Error("Error getting comments", "error", err)
	}

	for _, comment := range comments.Items {
		upd, formatErr := c.FormatComment(timeModified, link.Title, comment)
		if formatErr != nil {
			return nil, fmt.Errorf("error parsing comment: %w", formatErr)
		}
		updates = append(updates, upd)
		slog.Debug("stack overflow message comment", "updated", upd)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	return updates, nil
}

func (c *StackOverflowClient) GetCommentUpdates(link string, timeLastUpdated time.Time) (*model.StackOfCommentUpdate, error) {
	var stackOfComments model.StackOfCommentUpdate
	commentsLink := fmt.Sprintf("%s/comments?order=desc&sort=creation&fromdate=%s&site=stackoverflow&filter=!nNPvSN_ZTx", link, strconv.Itoa(int(timeLastUpdated.Unix())))
	if err := c.fetchUpdates(commentsLink, "comments", &stackOfComments); err != nil {
		return nil, err
	}
	return &stackOfComments, nil
}

func (c *StackOverflowClient) GetAnswerUpdates(link string, timeLastUpdated time.Time) (*model.StackOfAnswerUpdate, error) {
	var stackOfAnswers model.StackOfAnswerUpdate
	answersLink := fmt.Sprintf("%s/answers?order=desc&sort=creation&fromdate=%s&site=stackoverflow&filter=!*Mg4Pjg.Veejm5gm", link, strconv.Itoa(int(timeLastUpdated.Unix())))
	if err := c.fetchUpdates(answersLink, "answers", &stackOfAnswers); err != nil {
		return nil, err
	}
	return &stackOfAnswers, nil
}

func (c *StackOverflowClient) FormatAnswer(lastModified time.Time, answer model.StackOfAnswer) (model.Update, error) {
	creationTime, err := parseUnixTime(strconv.Itoa(answer.CreationDate))
	if err != nil {
		return model.Update{}, fmt.Errorf("error parsing creation date in stackoverflow update: %w", err)
	}

	var update model.Update
	update.Title = "Вопрос: " + answer.Title
	update.Author = answer.Owner.DisplayName
	update.TimeCreated = creationTime
	update.LastModified = lastModified
	preview := answer.Body
	if len(preview) > stackOverflowPreviewMaxLength {
		preview = preview[:stackOverflowPreviewMaxLength]
	}
	update.Description = "Ответ: " + preview

	return update, nil
}

func (c *StackOverflowClient) FormatComment(lastModified time.Time, title string,
	comment model.StackOfComment) (model.Update, error) {
	creationTime, err := parseUnixTime(strconv.Itoa(comment.CreationDate))
	if err != nil {
		return model.Update{}, fmt.Errorf("error parsing creation date in stackoverflow comment: %w", err)
	}

	var update model.Update
	update.Title = "Вопрос: " + title
	update.Author = comment.Owner.DisplayName
	update.TimeCreated = creationTime
	update.LastModified = lastModified

	preview := comment.Body
	if len(preview) > stackOverflowPreviewMaxLength {
		preview = preview[:stackOverflowPreviewMaxLength]
	}
	update.Description = "Комментарий: " + preview

	return update, nil
}
func parseUnixTime(tm string) (time.Time, error) {
	i, err := strconv.ParseInt(tm, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing last activity date: %w", err)
	}
	timeModified := time.Unix(i, 0)
	return timeModified, nil
}

func (c *StackOverflowClient) fetchUpdates(requestURL string, kind string, target any) error {
	slog.Debug("link to stackoverflow updates", "kind", kind, "link", requestURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("creating stackoverflow %s request: %w", kind, err)
	}

	req.Header.Add("Authorization", "token "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending stackoverflow %s request: %w", kind, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("closing stackoverflow response body", "kind", kind, "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Error getting updates", "kind", kind, "unexpected status code", resp.Status)
		return fmt.Errorf("unexpected status code for stackoverflow %s request: %d", kind, resp.StatusCode)
	}

	if decodeErr := json.NewDecoder(resp.Body).Decode(target); decodeErr != nil {
		return fmt.Errorf("decoding stackoverflow %s response: %w", kind, decodeErr)
	}

	return nil
}
