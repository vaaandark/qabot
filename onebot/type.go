package onebot

import (
	"strconv"
	"strings"
)

// TODO: array message type
type Event struct {
	Time        int64          `json:"time"`
	SelfId      int64          `json:"self_id"`
	PostType    string         `json:"post_type"`
	MessageType string         `json:"message_type"`
	SubType     string         `json:"sub_type"`
	MessageId   int32          `json:"message_id"`
	UserId      int64          `json:"user_id"`
	RawMessage  string         `json:"raw_message"`
	Sender      Sender         `json:"sender"`
	GroupId     *int64         `json:"group_id,omitempty"`
	Message     []TypedMessage `json:"message"`
}

func (e Event) IsFromSelf() bool {
	return e.SelfId == e.UserId
}

func (e Event) IsMessage() bool {
	return e.PostType == "message"
}

func (e Event) IsInGroup() bool {
	return e.GroupId != nil
}

func (e Event) ProcessText() (text string, shouldBeIgnored bool) {
	shouldBeIgnored = true
	if !e.IsInGroup() {
		shouldBeIgnored = false
	}

	for _, m := range e.Message {
		if m.Type == "text" {
			text += m.Data.Text
		} else if m.Type == "at" {
			if strconv.FormatInt(e.SelfId, 10) == m.Data.Qq {
				shouldBeIgnored = false
			}
		}
	}

	if shouldBeIgnored {
		if strings.HasPrefix(text, "v ") {
			shouldBeIgnored = false
			text = strings.TrimPrefix(text, "v ")
		} else if strings.HasPrefix(text, "/") { // cmd
			shouldBeIgnored = false
		}
	}
	return
}

type Sender struct {
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	GroupId  *int64 `json:"group_id,omitempty"`
}
type Data struct {
	Text string `json:"text,omitempty"`
	Qq   string `json:"qq,omitempty"`
	Id   string `json:"id,omitempty"`
}

type TypedMessage struct {
	Type string `json:"type,omitempty"`
	Data Data   `json:"data"`
}

type PrivateMessage struct {
	UserId  int64          `json:"user_id"`
	Message []TypedMessage `json:"message"`
}

type GroupMessage struct {
	GroupId int64          `json:"group_id"`
	Message []TypedMessage `json:"message"`
}

func NewPrivateMessage(userId int64, messageText string) PrivateMessage {
	return PrivateMessage{
		UserId: userId,
		Message: []TypedMessage{
			{
				Type: "text",
				Data: Data{
					Text: messageText,
				},
			},
		},
	}
}

func NewGroupMessage(groupId int64, messageText string, at *string, replyTo *string) GroupMessage {
	message := []TypedMessage{}

	if replyTo != nil {
		message = append(message, TypedMessage{
			Type: "reply",
			Data: Data{
				Id: *replyTo,
			},
		})
	}

	if at != nil {
		message = append(message, TypedMessage{
			Type: "at",
			Data: Data{
				Qq: *at,
			},
		})
	}

	message = append(message, TypedMessage{
		Type: "text",
		Data: Data{
			Text: messageText,
		},
	})

	return GroupMessage{
		GroupId: groupId,
		Message: message,
	}
}
