package flux

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func promptToJson(prompt string) *bytes.Reader {
	var data struct {
		Prompt string `json:"prompt"`
	}
	data.Prompt = prompt

	js, _ := json.Marshal(data)

	return bytes.NewReader(js)
}

func parseImageURL(r io.ReadCloser) (string, error) {
	var body struct {
		Images []struct {
			URL string `json:"url"`
		} `json:"Images"`
	}

	if err := json.NewDecoder(r).Decode(&body); err != nil {
		return "", err
	}

	if len(body.Images) < 1 {
		return "", fmt.Errorf("bad request")
	}

	return body.Images[0].URL, nil
}

func getImageBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
