package chats

import (
	"context"
	"sync"

	"github.com/google/generative-ai-go/genai"
)

type Chats struct {
	chats map[int64]*Chat
	mu    sync.RWMutex
}

func New() *Chats {
	return &Chats{
		chats: make(map[int64]*Chat),
	}
}

func (c *Chats) NewChat(id int64, model *genai.GenerativeModel) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chats[id] = NewChat(model)
}

func (c *Chats) Remove(id int64) {
	delete(c.chats, id)
}

func (c *Chats) Exists(id int64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.chats[id]
	return ok
}

func (c *Chats) Send(ctx context.Context, id int64, parts ...genai.Part) ([]genai.Part, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resp, err := c.chats[id].SendAsync(ctx, parts...)
	if err != nil {
		return nil, err
	}

	return resp.Candidates[0].Content.Parts, nil
}
