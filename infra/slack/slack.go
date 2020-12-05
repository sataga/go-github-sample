package slack

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	apiURL = "https://hooks.slack.com/services/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"
)

// Message is struct of message
type Message struct {
	Channel  string `json:"channel"`
	Username string `json:"username"`
	Text     string `json:"text"`
}

// PostMessage is function which send to slack
func PostMessage(channel, username, text string) (res *http.Response, err error) {
	m := Message{
		Channel:  channel,
		Username: username,
		Text:     text,
	}
	b, _ := json.Marshal(m)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err = client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	return
}
