package main

import (
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

func modelTools() *genai.Tool {
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
		}, {
			Name:        "getWeather",
			Description: "Get weather information in the specified city",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"city": {
						Type:        genai.TypeString,
						Description: "The city you want to know the weather for. Default should be Jurmala",
					},
				},
				Required: []string{"city"},
			},
		}},
	}
}

func getImageGenerationPrompt(funcall genai.FunctionCall) (string, error) {
	promptraw, ok := funcall.Args["prompt"]
	if !ok {
		return "", fmt.Errorf("argument prompt not found")
	}

	prompt, ok := promptraw.(string)
	if !ok {
		return "", fmt.Errorf("expected prompt type string got %T", promptraw)
	}

	return prompt, nil
}

func getWeatherCity(funcall genai.FunctionCall) (string, error) {
	cityraw, ok := funcall.Args["city"]
	if !ok {
		return "", fmt.Errorf("argument city not found")
	}

	city, ok := cityraw.(string)
	if !ok {
		return "", fmt.Errorf("expected city type string got %T", cityraw)
	}

	return city, nil
}
