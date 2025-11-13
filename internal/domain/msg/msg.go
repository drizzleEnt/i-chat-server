package msg

type Message struct {
	Content  string `json:"content"`
	SenderID string `json:"sender"`
	ChatID   string `json:"chat_id"`
}
