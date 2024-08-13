package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"github.com/moonlags/sherstjanka/internal/photo"
)

type server struct {
	client *genai.Client
	bot    *tgbotapi.BotAPI
	model  *genai.GenerativeModel
	chats  map[int64]*genai.ChatSession
	photos chan *photo.Photo
}

func (server *server) run() {
	go server.imageHandler()

	updates := server.bot.GetUpdatesChan(tgbotapi.NewUpdate(1))

	for update := range updates {
		if update.Message == nil {
			continue
		}
		slog.Info("Message recieved", "text", update.Message.Text, "firstname", update.Message.From.FirstName)

		go server.getTextResponse(update)
	}
}

func (server *server) getTextResponse(update tgbotapi.Update) {
	id := update.FromChat().ID
	if server.chats[id] == nil {
		server.chats[id] = server.model.StartChat()
		go server.chatDestruct(id, time.Hour*24)
	}

	prompt := server.populatePrompt(update.Message)
	if len(prompt) < 1 {
		return
	}

	data, err := server.chats[id].SendMessage(context.Background(), prompt...)
	if err != nil {
		slog.Error("Can not get model response", "err", err)

		msg := tgbotapi.NewMessage(update.FromChat().ID, "Мне не нравится твое сообщение")
		if _, err := server.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
			os.Exit(1)
		}
		return
	}

	response, err := parseReponse(fmt.Sprint(data.Candidates[0].Content.Parts[0]))
	if err != nil {
		slog.Error("Can not parse model response", "err", err)
		os.Exit(1)
	}

	msg, err := response.telegramMessage(update, server.chats[id], server.photos)
	if err != nil {
		slog.Error("Can not construct telegram message", "err", err)
		os.Exit(1)
	}

	if _, err := server.bot.Send(msg); err != nil {
		slog.Error("Can not send message", "err", err)
		os.Exit(1)
	}
}

func (server *server) chatDestruct(id int64, duration time.Duration) {
	time.Sleep(duration)
	delete(server.chats, id)
}

func (server *server) populatePrompt(message *tgbotapi.Message) []genai.Part {
	prompt := make([]genai.Part, 0)
	if message.Text != "" {
		prompt = append(prompt, genai.Text(message.Text))
	} else if message.Video != nil {
		prompt = append(prompt, genai.FileData{URI: server.uploadMedia(message.Video.FileID)})
	} else if message.VideoNote != nil {
		prompt = append(prompt, genai.FileData{URI: server.uploadMedia(message.VideoNote.FileID)})
	} else if len(message.Photo) > 0 {
		prompt = append(prompt, genai.FileData{URI: server.uploadMedia(message.Photo[0].FileID)})
	} else if message.Sticker != nil {
		prompt = append(prompt, genai.FileData{URI: server.uploadMedia(message.Sticker.Thumbnail.FileID)})
	} else if message.Audio != nil {
		prompt = append(prompt, genai.FileData{URI: server.uploadAudio(message.Audio.FileID, message.Audio.MimeType)})
	} else if message.Voice != nil {
		prompt = append(prompt, genai.FileData{URI: server.uploadAudio(message.Voice.FileID, message.Voice.MimeType)})
	}

	if message.Caption != "" {
		prompt = append(prompt, genai.Text(message.Caption))
	}
	return prompt
}

func (server *server) uploadAudio(fileID string, mimeType string) string {
	url, err := server.bot.GetFileDirectURL(fileID)
	if err != nil {
		slog.Error("Can not get audio url", "err", err)
		os.Exit(1)
	}

	response, err := http.Get(url)
	if err != nil {
		slog.Error("Can not download audio", "err", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	file, err := server.client.UploadFile(context.Background(), "", response.Body, &genai.UploadFileOptions{MIMEType: mimeType})
	if err != nil {
		slog.Error("Can not upload audio", "err", err)
		os.Exit(1)
	}
	return file.URI
}

func (server *server) uploadMedia(fileID string) string {
	url, err := server.bot.GetFileDirectURL(fileID)
	if err != nil {
		slog.Error("Can not get media url", "err", err)
		os.Exit(1)
	}

	response, err := http.Get(url)
	if err != nil {
		slog.Error("Can not download media", "err", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	file, err := server.client.UploadFile(context.Background(), "", response.Body, nil)
	if err != nil {
		slog.Error("Can not upload media", "err", err)
		os.Exit(1)
	}
	return file.URI
}
