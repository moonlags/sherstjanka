package chats

import (
	"context"
	"sync"
	"time"

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
	go c.chatDestruct(id, 24*time.Hour)
}

func (c *Chats) Exists(id int64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.chats[id]
	return ok
}

func (c *Chats) chatDestruct(id int64, dur time.Duration) {
	time.Sleep(dur)

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.chats, id)
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
