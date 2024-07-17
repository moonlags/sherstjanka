package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
)

type Server struct {
	client *genai.Client
	bot    *tgbotapi.BotAPI
	model  *genai.GenerativeModel
	chats  map[int64]*genai.ChatSession
}

func NewServer(client *genai.Client, bot *tgbotapi.BotAPI, model *genai.GenerativeModel) *Server {
	server := &Server{
		client: client,
		bot:    bot,
		model:  model,
		chats:  make(map[int64]*genai.ChatSession),
	}
	return server
}

func (server *Server) Run(config tgbotapi.UpdateConfig) {
	updates := server.bot.GetUpdatesChan(config)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		go server.getTextResponse(update)
	}
}

func (server *Server) getTextResponse(update tgbotapi.Update) {
	id := update.FromChat().ID
	if server.chats[id] == nil {
		server.chats[id] = server.model.StartChat()
		go server.chatDestruct(id, time.Hour*24)
	}
	prompt := server.populatePrompt(update.Message)
	if len(prompt) < 1 {
		return
	}
	response, err := server.chats[id].SendMessage(context.Background(), prompt...)
	if err != nil {
		msg := tgbotapi.NewMessage(update.FromChat().ID, "Мне не нравится твое сообщение")
		if _, err := server.bot.Send(msg); err != nil {
			log.Fatal("Failed to send message:", err)
		}
		fmt.Println("Failed to get model response:", err)
		return
	}
	msg := tgbotapi.NewMessage(update.FromChat().ID, fmt.Sprint(response.Candidates[0].Content.Parts[0]))
	if _, err := server.bot.Send(msg); err != nil {
		log.Fatal("Failed to send message:", err)
	}
}

func (server *Server) chatDestruct(id int64, duration time.Duration) {
	time.Sleep(duration)
	delete(server.chats, id)
}

func (server *Server) populatePrompt(message *tgbotapi.Message) []genai.Part {
	prompt := make([]genai.Part, 0)
	if message.Text != "" {
		prompt = append(prompt, genai.Text(message.Text))
	} else if message.Video != nil {
		prompt = append(prompt, genai.FileData{URI: server.uploadMedia(message.Video.FileID)})
		if message.Caption != "" {
			prompt = append(prompt, genai.Text(message.Caption))
		}
	} else if message.VideoNote != nil {
		prompt = append(prompt, genai.FileData{URI: server.uploadMedia(message.VideoNote.FileID)})
	} else if len(message.Photo) > 0 {
		prompt = append(prompt, genai.FileData{URI: server.uploadMedia(message.Photo[0].FileID)})
		if message.Caption != "" {
			prompt = append(prompt, genai.Text(message.Caption))
		}
	} else if message.Sticker != nil {
		prompt = append(prompt, genai.FileData{URI: server.uploadMedia(message.Sticker.Thumbnail.FileID)})
	} else if message.Audio != nil {
		prompt = append(prompt, genai.FileData{URI: server.uploadAudio(message.Audio.FileID, message.Audio.MimeType)})
	} else if message.Voice != nil {
		prompt = append(prompt, genai.FileData{URI: server.uploadAudio(message.Voice.FileID, message.Voice.MimeType)})
	}
	return prompt
}

func (server *Server) uploadAudio(fileID string, mimeType string) string {
	url, err := server.bot.GetFileDirectURL(fileID)
	if err != nil {
		log.Fatal("Failed to get video url:", err)
	}
	response, err := http.Get(url)
	if err != nil {
		log.Fatal("Failed to download media:", err)
	}
	defer response.Body.Close()
	file, err := server.client.UploadFile(context.Background(), "", response.Body, &genai.UploadFileOptions{MIMEType: mimeType})
	if err != nil {
		log.Fatal("Failed to upload media:", err)
	}
	return file.URI
}

func (server *Server) uploadMedia(fileID string) string {
	url, err := server.bot.GetFileDirectURL(fileID)
	if err != nil {
		log.Fatal("Failed to get video url:", err)
	}
	response, err := http.Get(url)
	if err != nil {
		log.Fatal("Failed to download media:", err)
	}
	defer response.Body.Close()
	file, err := server.client.UploadFile(context.Background(), "", response.Body, nil)
	if err != nil {
		log.Fatal("Failed to upload media:", err)
	}
	return file.URI
}
