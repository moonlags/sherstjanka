package flux

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func optsToInfo(prompt string, randomiseSeed bool, opts *GenerationOptions) *generationInfo {
	if opts == nil {
		opts = new(GenerationOptions)
		opts.populate()
	}

	return &generationInfo{
		Prompt:            prompt,
		RandomiseSeed:     randomiseSeed,
		Seed:              opts.Seed,
		Width:             opts.Width,
		Height:            opts.Height,
		NumInferenceSteps: opts.NumInferenceSteps,
	}
}

func parseImageURL(r io.ReadCloser) (string, error) {
	js, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	fmt.Println(string(js))

	start := bytes.LastIndex(js, []byte("data")) + 7
	end := len(js) - 14
	if start > end {
		return "", fmt.Errorf("bad request")
	}

	js = bytes.TrimRight(js[start:end], ",")

	fmt.Println(string(js))

	var data struct {
		URL string `json:"url"`
	}

	if err := json.Unmarshal(js, &data); err != nil {
		return "", err
	}

	return data.URL, nil
}

func getImageBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
