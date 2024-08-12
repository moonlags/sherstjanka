package main

import (
	"encoding/json"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/moonlags/sherstjanka/internal/flux"
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

func (resp *modelResponse) telegramMessage(update tgbotapi.Update) (tgbotapi.Chattable, error) {
	if resp.ImagePrompt != "" {
		data, err := flux.GenerateImage(resp.ImagePrompt, true, nil)
		if err != nil {
			return nil, err
		}

		msg := tgbotapi.NewPhoto(update.FromChat().ID, tgbotapi.FileBytes{Bytes: data})
		msg.Caption = resp.Response

		return msg, nil
	}

	return tgbotapi.NewMessage(update.FromChat().ID, resp.Response), nil
}
