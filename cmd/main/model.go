package main

import (
	"log/slog"
	"os"

	"github.com/google/generative-ai-go/genai"
)

func newModel(client *genai.Client) *genai.GenerativeModel {
	instructions, err := os.ReadFile("instructions.txt")
	if err != nil {
		slog.Error("Can not read instructions from instructions.txt", "err", err)
		os.Exit(1)
	}

	model := client.GenerativeModel("gemini-1.5-flash-latest")

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(string(instructions))},
	}

	model.Tools = []*genai.Tool{imageGenerationTool()}

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

func imageGenerationTool() *genai.Tool {
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "generateImage",
			Description: "Generate image providing an image description",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"prompt": {
						Type:        genai.TypeString,
						Description: "Description of the image to generate",
					},
				},
				Required: []string{"prompt"},
			},
		}},
	}
}
