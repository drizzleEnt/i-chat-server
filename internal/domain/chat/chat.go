package chatdomain

type Chat struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateChatRequest struct {
	Name string `json:"name"`
}
