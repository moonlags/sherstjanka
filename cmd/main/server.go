package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"github.com/moonlags/sherstjanka/internal/chats"
	"github.com/moonlags/sherstjanka/internal/flux"
	"github.com/moonlags/sherstjanka/internal/openweathermap"
)

type server struct {
	client    *genai.Client
	bot       *tgbotapi.BotAPI
	model     *genai.GenerativeModel
	chats     *chats.Chats
	image     *flux.Config
	weather   *openweathermap.Config
	whitelist int64
}

func (server *server) run() {
	u := tgbotapi.NewUpdate(-1)
	u.Timeout = 60

	updates := server.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		slog.Debug("message recieved", "text", update.Message.Text, "user", update.Message.From)

		go server.getTextResponse(update)
	}
}

func (server *server) getTextResponse(update tgbotapi.Update) {
	if !server.checkWhitelist(update) {
		slog.Warn("user is not in whitelist", "user", update.Message.From)

		msg := tgbotapi.NewMessage(update.FromChat().ID, "Обратитесь к администратору (@ridxj) для доступа к Шерстянке")
		if _, err := server.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
		}
		return
	}

	id := update.FromChat().ID
	if !server.chats.Exists(id) {
		server.chats.NewChat(id, server.model)
	}

	if update.Message.Text == "/reset" {
		slog.Info("reseting chat", "chat_id", update.FromChat().ID)

		msg := tgbotapi.NewMessage(update.FromChat().ID, "Ой, я все забыла")
		if _, err := server.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
		}

		server.chats.Remove(update.FromChat().ID)
		return
	}

	prompt, err := server.populatePrompt(update.Message)
	if err != nil {
		slog.Error("Can not populate prompt", "err", err)
		return
	}

	if len(prompt) < 1 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	parts, err := server.chats.Send(ctx, id, prompt...)
	if err != nil {
		slog.Error("Can not get model response", "err", err)

		msg := tgbotapi.NewMessage(update.FromChat().ID, "Мне не нравится твое сообщение")
		if _, err := server.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
		}
		return
	}

	slog.Debug("model response", "parts", parts)

	for _, part := range parts {
		response := server.parseReponse(update, part)

		if response == "" {
			continue
		}

		msg := tgbotapi.NewMessage(update.FromChat().ID, response)
		msg.ReplyToMessageID = update.Message.MessageID

		if _, err := server.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
		}
	}
}

func (server *server) populatePrompt(message *tgbotapi.Message) ([]genai.Part, error) {
	prompt := make([]genai.Part, 0)
	if message.Text != "" {
		text := fmt.Sprintf("%s: %s", message.From.FirstName, message.Text)
		prompt = append(prompt, genai.Text(text))
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
		text := fmt.Sprintf("%s: %s", message.From.FirstName, message.Caption)
		prompt = append(prompt, genai.Text(text))
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	file, err := server.client.UploadFile(ctx, "", response.Body, &genai.UploadFileOptions{MIMEType: mimeType})
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	file, err := server.client.UploadFile(ctx, "", response.Body, nil)
	if err != nil {
		return "", err
	}
	return file.URI, nil
}
