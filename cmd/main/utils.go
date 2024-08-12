package main

import (
	"context"
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"github.com/moonlags/sherstjanka/internal/photo"
)

type modelResponse struct {
	Response    string `json:"response"`
	ImagePrompt string `json:"image_prompt"`
}

func parseReponse(response string) (*modelResponse, error) {
	parsed := new(modelResponse)
	if err := json.Unmarshal([]byte(response), parsed); err != nil {
		return nil, err
	}

	return parsed, nil
}

func (resp *modelResponse) telegramMessage(update tgbotapi.Update, chat *genai.ChatSession, photos chan *photo.Photo) (tgbotapi.Chattable, error) {
	fmt.Printf("%#v\n", resp)

	if resp.ImagePrompt != "" {
		if len(photos) >= 5 {
			resp, err := chat.SendMessage(context.Background(), genai.Text("ImageGenerationResponse: queue is full"))
			if err != nil {
				return nil, err
			}

			parsed, err := parseReponse(fmt.Sprint(resp.Candidates[0].Content.Parts[0]))
			if err != nil {
				return nil, err
			}

			msg := tgbotapi.NewMessage(update.FromChat().ID, parsed.Response)
			msg.ReplyToMessageID = update.Message.MessageID

			return msg, nil
		}
		photos <- &photo.Photo{MessageID: update.Message.MessageID, ChatID: update.FromChat().ID, Prompt: resp.ImagePrompt, ChatSession: chat}
	}

	msg := tgbotapi.NewMessage(update.FromChat().ID, resp.Response)
	msg.ReplyToMessageID = update.Message.MessageID

	return msg, nil
}

func generationFailure(photo *photo.Photo) (tgbotapi.Chattable, error) {
	resp, err := photo.ChatSession.SendMessage(context.Background(), genai.Text("ImageGenerationResponse: \""+photo.Prompt+"\" failed"))
	if err != nil {
		return nil, err
	}

	parsed, err := parseReponse(fmt.Sprint(resp.Candidates[0].Content.Parts[0]))
	if err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(photo.ChatID, parsed.Response)
	msg.ReplyToMessageID = photo.MessageID

	return msg, nil
}

func generationSuccess(photo *photo.Photo, data []byte) (tgbotapi.Chattable, error) {
	resp, err := photo.ChatSession.SendMessage(context.Background(), genai.Text("ImageGenerationResponse: \""+photo.Prompt+"\" is ready"))
	if err != nil {
		return nil, err
	}

	parsed, err := parseReponse(fmt.Sprint(resp.Candidates[0].Content.Parts[0]))
	if err != nil {
		return nil, err
	}

	msg := tgbotapi.NewPhoto(photo.ChatID, tgbotapi.FileBytes{Bytes: data})
	msg.ReplyToMessageID = photo.MessageID
	msg.Caption = parsed.Response

	return msg, nil
}
