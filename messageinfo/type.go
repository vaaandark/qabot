package messageinfo

import "qabot/onebot"

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
		m.Text = *text
	}
	return m
}

func (m MessageInfo) IsInGroup() bool {
	return m.GroupId != nil
}
