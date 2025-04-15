package messageenvelope

import (
	"fmt"
	"strings"
	"time"

	"github.com/vaaandark/qabot/pkg/onebot"
)

type MessageEnvelope struct {
	Nickname   string
	UserId     int64
	TargetId   *int64
	GroupId    *int64
	Text       string
	MessageId  int32
	ReplyTo    *int32
	IsFromSelf bool
	Category   onebot.MessageCategory
	IsAt       bool
	Timestamp  time.Time
	ModelName  string
}

func (m MessageEnvelope) GetGroupOrUserID() int64 {
	if m.GroupId != nil {
		return *m.GroupId
	}
	return m.UserId
}
func (m MessageEnvelope) GetNamespacedGroupOrUserID() string {
	if m.GroupId != nil {
		return fmt.Sprintf("group/%d", *m.GroupId)
	} else {
		return fmt.Sprintf("user/%d", m.UserId)
	}
}

func FromEvent(event onebot.Event, text *string, replyTo *int32, category onebot.MessageCategory, isAt bool) MessageEnvelope {
	m := MessageEnvelope{
		Nickname:   event.Sender.Nickname,
		UserId:     event.UserId,
		TargetId:   event.TargetId,
		GroupId:    event.GroupId,
		Text:       event.RawMessage,
		MessageId:  event.MessageId,
		ReplyTo:    replyTo,
		IsFromSelf: event.IsFromSelf(),
		Category:   category,
		IsAt:       isAt,
		Timestamp:  time.Now(),
	}
	if text != nil {
		m.Text = strings.TrimSpace(*text)
	}
	return m
}

func (m MessageEnvelope) IsInGroup() bool {
	return m.GroupId != nil
}
