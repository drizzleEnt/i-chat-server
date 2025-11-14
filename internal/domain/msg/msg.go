package msgdomain

type ActionType string

const (
	ActionSendText   ActionType = "send_text"
	ActionSendBinary ActionType = "send_binary"
	ActionJoinChat   ActionType = "join_chat"
	ActionLeaveChat  ActionType = "leave_chat"
	ActionCreateChat ActionType = "create_chat"
)

type Message struct {
	Action   string `json:"action"`
	Content  string `json:"content"`
	SenderID string `json:"sender"`
	ChatID   string `json:"chat_id"`
}
