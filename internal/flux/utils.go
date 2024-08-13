package flux

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func promptToJson(prompt string) *bytes.Reader {
	data := fmt.Sprintf("{\"prompt\":\"%s\"}", prompt)

	return bytes.NewReader([]byte(data))
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
		return "", fmt.Errorf("malformed data")
	}

	return body.Images[0].URL, nil
}
