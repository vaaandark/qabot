package onebot

import (
	"fmt"
	"strconv"
	"strings"
)

type Event struct {
	Time        int64          `json:"time"`
	SelfId      int64          `json:"self_id"`
	PostType    string         `json:"post_type"`
	MessageType string         `json:"message_type"`
	SubType     string         `json:"sub_type"`
	MessageId   int32          `json:"message_id"`
	UserId      int64          `json:"user_id"`
	TargetId    *int64         `json:"target_id,omitempty"`
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

func (e Event) IsAtSelf() bool {
	for _, m := range e.Message {
		if m.Type == "at" && m.Data.Qq == strconv.FormatInt(e.SelfId, 10) {
			return true
		}
	}
	return false
}

func (e Event) ReplyTo() *int32 {
	for _, m := range e.Message {
		if m.Type == "reply" {
			n, err := strconv.ParseInt(m.Data.Id, 10, 32)
			if err != nil {
				continue
			}
			replyTo := int32(n)
			return &replyTo
		}
	}
	return nil
}

func (e Event) IsInGroup() bool {
	return e.GroupId != nil
}

func (e Event) CatText() string {
	text := ""
	for _, m := range e.Message {
		if m.Type == "text" {
			text += m.Data.Text
		}
	}
	return text
}

// 处理消息文本并决定消息是否应该传给 chatter
// 传递给 chatter 不代表一定会被回复，chatter 内部还有处理
// 应该传给 chatter 的情况：
//   - 私聊
//   - 群聊并是一个回复
//   - 群聊并 at 了 bot
func (e Event) ProcessText() (text string, replyTo *int32, shouldBeIgnored, isCmd, isAt bool) {
	shouldBeIgnored = true

	replyTo = e.ReplyTo()

	if !e.IsInGroup() {
		shouldBeIgnored = false
	} else {
		if e.IsAtSelf() {
			shouldBeIgnored = false
			isAt = false
		}

		if replyTo != nil {
			shouldBeIgnored = false
		}
	}

	text = e.CatText()

	if shouldBeIgnored {
		text = strings.TrimSpace(text)
		if strings.HasPrefix(text, "v ") {
			shouldBeIgnored = false
			text = strings.TrimSpace(strings.TrimPrefix(text, "v"))
		}
	}

	if strings.HasPrefix(text, "/") { // cmd
		shouldBeIgnored = false
		isCmd = true
		text = strings.TrimSpace(strings.TrimPrefix(text, "/"))
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

func NewPrivateMessage(dialogBaseUrl string, userId int64, modelName string, messageText string, replyTo *string) PrivateMessage {
	message := []TypedMessage{}

	if replyTo != nil {
		message = append(message, TypedMessage{
			Type: "reply",
			Data: Data{
				Id: *replyTo,
			},
		})
	}

	if len(modelName) != 0 {
		message = append(message, TypedMessage{
			Type: "text",
			Data: Data{
				Text: fmt.Sprintf("[%s]\n\n查看最新上下文：%s/user/%d/latest\n查看上下文树：%s/user/%d/all\n\n-----\n\n", modelName, dialogBaseUrl, userId, dialogBaseUrl, userId),
			},
		})
	}

	message = append(message, TypedMessage{
		Type: "text",
		Data: Data{
			Text: messageText,
		},
	})

	return PrivateMessage{
		UserId:  userId,
		Message: message,
	}
}

func NewGroupMessage(dialogBaseUrl string, groupId int64, modelName string, messageText string, at *string, replyTo *string) GroupMessage {
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

	if len(modelName) != 0 {
		message = append(message, TypedMessage{
			Type: "text",
			Data: Data{
				Text: fmt.Sprintf("\n[%s]\n\n查看最新上下文：%s/group/%d/latest\n查看上下文树：%s/group/%d/all\n\n-----\n\n", modelName, dialogBaseUrl, groupId, dialogBaseUrl, groupId),
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

type SendResponse struct {
	Status  string           `json:"status"`
	RetCode int32            `json:"retcode"`
	Data    SendResponseData `json:"data"`
}

type SendResponseData struct {
	MessageId int32 `json:"message_id"`
}
