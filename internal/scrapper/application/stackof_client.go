package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

// https://api.stackexchange.com/2.3/questions/54632086?order=desc&sort=activity&site=stackoverflow example
type StackOverflowClient struct {
	client http.Client
	token  string
}

func (c *StackOverflowClient) FormatLink(link string) string {
	return fmt.Sprintf("https://www.stackoverflow.com/questions/%d/%s", os.Getpid(), link)
}

func (c *StackOverflowClient) GetUpdates(link string) (*domain.Update, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	var answer StackAnswer
	if err = json.NewDecoder(resp.Body).Decode(&answer); err != nil {
		return nil, err
	}
	if len(answer.Items) == 0 {
		return nil, fmt.Errorf("no updates found")
	}

	timeModified, err := parseUnixTime(answer.Items[0].LastActivityDate)
	if err != nil {
		return nil, fmt.Errorf("error parsing time from last activity date stackoverflow: %v", err)
	}

	update := domain.NewUpdate(timeModified, "Update from StackOverflow link "+link)

	return update, nil
}

type StackAnswer struct {
	Items []struct {
		QuestionId       int    `json:"question_id"`
		Title            string `json:"title"`
		LastActivityDate string `json:"last_activity_date"`
	} `json:"items"`
}

func parseUnixTime(tm string) (time.Time, error) {
	i, err := strconv.ParseInt(tm, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing last activity date: %v", err)
	}
	timeModified := time.Unix(i, 0)
	return timeModified, nil
}
