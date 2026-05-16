package model

type StackOfAnswerUpdate struct {
	Items          []StackOfAnswer `json:"items"`
	HasMore        bool            `json:"has_more"`
	QuotaMax       int             `json:"quota_max"`
	QuotaRemaining int             `json:"quota_remaining"`
}

type StackOfCommentUpdate struct {
	Items          []StackOfComment `json:"items"`
	HasMore        bool             `json:"has_more"`
	QuotaMax       int              `json:"quota_max"`
	QuotaRemaining int              `json:"quota_remaining"`
}

type StackOfAnswer struct {
	Owner struct {
		UserID      int    `json:"user_id"`
		DisplayName string `json:"display_name"`
	} `json:"owner"`
	Score            int    `json:"score"`
	LastActivityDate int    `json:"last_activity_date"`
	LastEditDate     int    `json:"last_edit_date"`
	CreationDate     int    `json:"creation_date"`
	AnswerID         int    `json:"answer_id"`
	QuestionID       int    `json:"question_id"`
	Link             string `json:"link"`
	Title            string `json:"title"`
	Body             string `json:"body"`
}

type StackOfComment struct {
	Owner struct {
		UserID      int    `json:"user_id"`
		DisplayName string `json:"display_name"`
	} `json:"owner"`
	Edited       bool   `json:"edited"`
	CreationDate int    `json:"creation_date"`
	PostID       int    `json:"post_id"`
	CommentID    int    `json:"comment_id"`
	Link         string `json:"link"`
	Body         string `json:"body"`
}

type StackDefaultAnswer struct {
	Items []struct {
		QuestionID int    `json:"question_id"`
		Title      string `json:"title"`
	} `json:"items"`
}
