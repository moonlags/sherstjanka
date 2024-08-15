package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"github.com/moonlags/sherstjanka/internal/flux"
)

type server struct {
	client    *genai.Client
	bot       *tgbotapi.BotAPI
	model     *genai.GenerativeModel
	chats     map[int64]*genai.ChatSession
	image     *flux.Config
	whitelist int64
}

func (server *server) run() {
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
	if !server.checkWhitelist(update) {
		slog.Warn("user is not in whitelist", "firstname", update.Message.From.FirstName)
		return
	}

	id := update.FromChat().ID
	if server.chats[id] == nil {
		server.chats[id] = server.model.StartChat()
		go server.chatDestruct(id, time.Hour*24)
	}

	prompt, err := server.populatePrompt(update.Message)
	if err != nil {
		slog.Error("Can not populate prompt", "err", err)
		return
	}

	data, err := server.chats[id].SendMessage(context.Background(), prompt...)
	if err != nil {
		slog.Error("Can not get model response", "err", err)

		msg := tgbotapi.NewMessage(update.FromChat().ID, "Мне не нравится твое сообщение")
		if _, err := server.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
		}
		return
	}

	for _, part := range data.Candidates[0].Content.Parts {
		msg, err := server.parseReponse(update, part)
		if err != nil {
			slog.Error("Can not parse model response", "err", err)
			return
		}

		if _, err := server.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
		}
	}
}

func (server *server) chatDestruct(id int64, duration time.Duration) {
	time.Sleep(duration)
	delete(server.chats, id)
}

func (server *server) populatePrompt(message *tgbotapi.Message) ([]genai.Part, error) {
	prompt := make([]genai.Part, 0)
	if message.Text != "" {
		prompt = append(prompt, genai.Text(message.Text))
	} else if message.Video != nil {
		url, err := server.uploadMedia(message.Video.FileID)
		if err != nil {
			return nil, err
		}

		prompt = append(prompt, genai.FileData{URI: url})
	} else if message.VideoNote != nil {
		url, err := server.uploadMedia(message.VideoNote.FileID)
		if err != nil {
			return nil, err
		}

		prompt = append(prompt, genai.FileData{URI: url})
	} else if len(message.Photo) > 0 {
		url, err := server.uploadMedia(message.Photo[0].FileID)
		if err != nil {
			return nil, err
		}

		prompt = append(prompt, genai.FileData{URI: url})
	} else if message.Sticker != nil {
		url, err := server.uploadMedia(message.Sticker.Thumbnail.FileID)
		if err != nil {
			return nil, err
		}

		prompt = append(prompt, genai.FileData{URI: url})
	} else if message.Audio != nil {
		url, err := server.uploadAudio(message.Audio.FileID, message.Audio.MimeType)
		if err != nil {
			return nil, err
		}

		prompt = append(prompt, genai.FileData{URI: url})
	} else if message.Voice != nil {
		url, err := server.uploadAudio(message.Voice.FileID, message.Voice.MimeType)
		if err != nil {
			return nil, err
		}

		prompt = append(prompt, genai.FileData{URI: url})
	}

	if message.Caption != "" {
		prompt = append(prompt, genai.Text(message.Caption))
	}
	return prompt, nil
}

func (server *server) uploadAudio(fileID string, mimeType string) (string, error) {
	url, err := server.bot.GetFileDirectURL(fileID)
	if err != nil {
		return "", err
	}

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	file, err := server.client.UploadFile(context.Background(), "", response.Body, &genai.UploadFileOptions{MIMEType: mimeType})
	if err != nil {
		return "", err
	}
	return file.URI, nil
}

func (server *server) uploadMedia(fileID string) (string, error) {
	url, err := server.bot.GetFileDirectURL(fileID)
	if err != nil {
		return "", err
	}

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	file, err := server.client.UploadFile(context.Background(), "", response.Body, nil)
	if err != nil {
		return "", err
	}
	return file.URI, nil
}
