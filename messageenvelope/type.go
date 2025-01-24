package messageenvelope

import (
	"qabot/onebot"
	"strings"
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
	IsCmd      bool
	IsAt       bool
}

func FromEvent(event onebot.Event, text *string, replyTo *int32, isCmd, isAt bool) MessageEnvelope {
	m := MessageEnvelope{
		Nickname:   event.Sender.Nickname,
		UserId:     event.UserId,
		TargetId:   event.TargetId,
		GroupId:    event.GroupId,
		Text:       event.RawMessage,
		MessageId:  event.MessageId,
		ReplyTo:    replyTo,
		IsFromSelf: event.IsFromSelf(),
		IsCmd:      isCmd,
		IsAt:       isAt,
	}
	if text != nil {
		m.Text = strings.TrimSpace(*text)
	}
	return m
}

func (m MessageEnvelope) IsInGroup() bool {
	return m.GroupId != nil
}
