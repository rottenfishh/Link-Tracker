package model

type Message struct {
	Text string
	/// smth later
}

func NewMessage(text string) *Message {
	return &Message{text}
}
