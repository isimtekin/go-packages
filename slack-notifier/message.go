package slacknotifier

// Message represents a Slack message
type Message struct {
	Text        string       `json:"text,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
	ThreadTS    string       `json:"thread_ts,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Blocks      []Block      `json:"blocks,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Fallback   string            `json:"fallback,omitempty"`
	Color      string            `json:"color,omitempty"`
	Pretext    string            `json:"pretext,omitempty"`
	AuthorName string            `json:"author_name,omitempty"`
	AuthorLink string            `json:"author_link,omitempty"`
	AuthorIcon string            `json:"author_icon,omitempty"`
	Title      string            `json:"title,omitempty"`
	TitleLink  string            `json:"title_link,omitempty"`
	Text       string            `json:"text,omitempty"`
	Fields     []AttachmentField `json:"fields,omitempty"`
	ImageURL   string            `json:"image_url,omitempty"`
	ThumbURL   string            `json:"thumb_url,omitempty"`
	Footer     string            `json:"footer,omitempty"`
	FooterIcon string            `json:"footer_icon,omitempty"`
	Timestamp  int64             `json:"ts,omitempty"`
	MarkdownIn []string          `json:"mrkdwn_in,omitempty"`
}

// AttachmentField represents a field in an attachment
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Block represents a Slack Block Kit element
type Block struct {
	Type     string                 `json:"type"`
	Text     *TextObject            `json:"text,omitempty"`
	Elements []interface{}          `json:"elements,omitempty"`
	BlockID  string                 `json:"block_id,omitempty"`
	Fields   []*TextObject          `json:"fields,omitempty"`
	Accessory interface{}           `json:"accessory,omitempty"`
}

// TextObject represents text in Block Kit
type TextObject struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

// MessageBuilder helps build Slack messages fluently
type MessageBuilder struct {
	message *Message
}

// NewMessage creates a new message builder
func NewMessage(text string) *MessageBuilder {
	return &MessageBuilder{
		message: &Message{
			Text: text,
		},
	}
}

// Channel sets the channel
func (mb *MessageBuilder) Channel(channel string) *MessageBuilder {
	mb.message.Channel = channel
	return mb
}

// Username sets the username
func (mb *MessageBuilder) Username(username string) *MessageBuilder {
	mb.message.Username = username
	return mb
}

// IconEmoji sets the icon emoji
func (mb *MessageBuilder) IconEmoji(emoji string) *MessageBuilder {
	mb.message.IconEmoji = emoji
	return mb
}

// IconURL sets the icon URL
func (mb *MessageBuilder) IconURL(url string) *MessageBuilder {
	mb.message.IconURL = url
	return mb
}

// Thread sets the thread timestamp for threading
func (mb *MessageBuilder) Thread(threadTS string) *MessageBuilder {
	mb.message.ThreadTS = threadTS
	return mb
}

// AddAttachment adds an attachment to the message
func (mb *MessageBuilder) AddAttachment(attachment Attachment) *MessageBuilder {
	mb.message.Attachments = append(mb.message.Attachments, attachment)
	return mb
}

// AddBlock adds a block to the message
func (mb *MessageBuilder) AddBlock(block Block) *MessageBuilder {
	mb.message.Blocks = append(mb.message.Blocks, block)
	return mb
}

// Build returns the built message
func (mb *MessageBuilder) Build() *Message {
	return mb.message
}

// NewAttachment creates a new attachment
func NewAttachment(fallback, text, color string) Attachment {
	return Attachment{
		Fallback: fallback,
		Text:     text,
		Color:    color,
	}
}

// AddField adds a field to the attachment
func (a *Attachment) AddField(title, value string, short bool) *Attachment {
	a.Fields = append(a.Fields, AttachmentField{
		Title: title,
		Value: value,
		Short: short,
	})
	return a
}

// NewSectionBlock creates a new section block
func NewSectionBlock(text string) Block {
	return Block{
		Type: "section",
		Text: &TextObject{
			Type: "mrkdwn",
			Text: text,
		},
	}
}

// NewDividerBlock creates a divider block
func NewDividerBlock() Block {
	return Block{
		Type: "divider",
	}
}

// NewHeaderBlock creates a header block
func NewHeaderBlock(text string) Block {
	return Block{
		Type: "header",
		Text: &TextObject{
			Type: "plain_text",
			Text: text,
		},
	}
}

// Color constants for attachments
const (
	ColorGood    = "good"    // Green
	ColorWarning = "warning" // Yellow
	ColorDanger  = "danger"  // Red
	ColorInfo    = "#36a64f" // Blue-green
)

// Predefined message templates

// NewSuccessMessage creates a success message with green color
func NewSuccessMessage(text string) *MessageBuilder {
	return NewMessage("").
		AddAttachment(NewAttachment("Success", text, ColorGood))
}

// NewWarningMessage creates a warning message with yellow color
func NewWarningMessage(text string) *MessageBuilder {
	return NewMessage("").
		AddAttachment(NewAttachment("Warning", text, ColorWarning))
}

// NewErrorMessage creates an error message with red color
func NewErrorMessage(text string) *MessageBuilder {
	return NewMessage("").
		AddAttachment(NewAttachment("Error", text, ColorDanger))
}

// NewInfoMessage creates an info message with blue color
func NewInfoMessage(text string) *MessageBuilder {
	return NewMessage("").
		AddAttachment(NewAttachment("Info", text, ColorInfo))
}
