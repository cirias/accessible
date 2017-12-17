// https://core.telegram.org/bots/api
// Bot API 3.5
package telbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const urlPattern = "https://api.telegram.org/bot%s/%s"

type Bot struct {
	token string
	Httpc *http.Client
}

func NewBot(token string) *Bot {
	return &Bot{
		token: token,
		Httpc: http.DefaultClient,
	}
}

type ChatId = int64

type Update struct {
	Id            int64          `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	Id       int64             `json:"message_id"`
	Chat     *Chat             `json:"chat"`
	Text     string            `json:"text"`
	Entities []*MessageEnitity `json:"entities"`
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *Message) Command() (string, string) {
	for _, entry := range m.Entities {
		if entry.Type == "bot_command" {
			start := min(entry.Offset+entry.Length+1, len(m.Text))
			return m.Text[entry.Offset : entry.Offset+entry.Length], m.Text[start:len(m.Text)]
		}
	}
	return "", ""
}

type Chat struct {
	Id   ChatId `json:"id"`
	Type string `json:"type"`
}

type MessageEnitity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

type CallbackQuery struct {
	Id      string   `json:"id"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

type GetUpdatesParams struct {
	Offset         int64    `json:"offset,omitempty"`
	Limit          int      `json:"limit,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

type MessageParams struct {
	ChatId      ChatId                `json:"chat_id"`
	Text        string                `json:"text"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]*InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type ResponseBody struct {
	Ok          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	ErrorCode   int             `json:"error_code"`
	Description string          `json:"description"`
}

func (b *Bot) request(u string, params interface{}) (json.RawMessage, error) {
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	if err := enc.Encode(params); err != nil {
		return nil, fmt.Errorf("could not encode params: %s", err)
	}

	req, err := http.NewRequest("GET", u, &buf)
	if err != nil {
		return nil, fmt.Errorf("could not new request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.Httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not do request: %s", err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var rb ResponseBody
	if err := dec.Decode(&rb); err != nil {
		return nil, fmt.Errorf("could not decode response: %s", err)
	}

	if !rb.Ok {
		return nil, fmt.Errorf("could not get result: %s", rb.Description)
	}

	return rb.Result, nil
}

func (b *Bot) GetUpdates(params *GetUpdatesParams) ([]*Update, error) {
	u := fmt.Sprintf(urlPattern, b.token, "getUpdates")

	result, err := b.request(u, params)
	if err != nil {
		return nil, err
	}

	var updates []*Update
	if err := json.Unmarshal(result, &updates); err != nil {
		return nil, fmt.Errorf("could not unmarshal updates: %s", err)
	}

	return updates, nil
}

func (b *Bot) SendMessage(params *MessageParams) (*Message, error) {
	u := fmt.Sprintf(urlPattern, b.token, "sendMessage")

	result, err := b.request(u, params)
	if err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(result, &msg); err != nil {
		return nil, fmt.Errorf("could not unmarshal message: %s", err)
	}

	return &msg, nil
}
