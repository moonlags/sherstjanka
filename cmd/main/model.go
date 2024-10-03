package main

import (
	"log/slog"
	"os"

	"github.com/google/generative-ai-go/genai"
)

func newModel(client *genai.Client) *genai.GenerativeModel {
	instructions, err := os.ReadFile("instructions.txt")
	if err != nil {
		slog.Warn("Can not read instructions from instructions.txt", "err", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash-latest")

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(string(instructions))},
	}

	model.Tools = []*genai.Tool{modelTools()}

	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockOnlyHigh,
		},
	}

	return model
}
