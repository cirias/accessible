// https://core.telegram.org/bots/api
// Bot API 3.5
package telbot

type Bot struct {
	token string
}

func NewBot(token string) *Bot {
	return &Bot{
		token: token,
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

func (m *Message) Command() string {
	// TODO
	return ""
}

func (m *Message) CommandArguments() string {
	// TODO
	return ""
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

type UpdatesParams struct {
}

type MessageParams struct {
	ChatId      ChatId                `json:"chat_id"`
	Text        string                `json:"text"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]*InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

func (b *Bot) GetUpdates(*UpdatesParams) ([]*Update, error) {
	return nil, nil
}

func (b *Bot) SendMessage(*MessageParams) (*Message, error) {
	return nil, nil
}

type SendMessageFunc = func(*MessageParams) (*Message, error)
