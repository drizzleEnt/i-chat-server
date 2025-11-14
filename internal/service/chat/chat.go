package chatsrv

import "sync"

type chat struct {
	m       sync.RWMutex
	chatID  string
	clients map[string]*client

	isClosed bool
}

func newChat(chatID string) *chat {
	return &chat{
		chatID:  chatID,
		clients: make(map[string]*client),
	}
}

func (c *chat) addClient(client *client) {
	c.m.Lock()
	defer c.m.Unlock()
	c.clients[client.id] = client
}

func (c *chat) removeClient(id string) {
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.clients, id)
}
