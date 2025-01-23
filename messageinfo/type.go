package messageinfo

import (
	"qabot/onebot"
	"strings"
)

type MessageInfo struct {
	Nickname  string
	UserId    int64
	GroupId   *int64
	Text      string
	MessageId int32
}

func FromEvent(event onebot.Event, text *string) MessageInfo {
	m := MessageInfo{
		Nickname:  event.Sender.Nickname,
		UserId:    event.UserId,
		GroupId:   event.GroupId,
		Text:      event.RawMessage,
		MessageId: event.MessageId,
	}
	if text != nil {
		m.Text = strings.TrimSpace(*text)
	}
	return m
}

func (m MessageInfo) IsCmd() bool {
	return strings.HasPrefix(m.Text, "/")
}

func (m MessageInfo) IsInGroup() bool {
	return m.GroupId != nil
}
