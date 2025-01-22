package messageinfo

import "qabot/event"

type MessageInfo struct {
	Nickname  string
	UserId    int64
	GroupId   *int64
	Text      string
	MessageId int32
}

func FromEvent(event event.Event) MessageInfo {
	return MessageInfo{
		Nickname:  event.Sender.Nickname,
		UserId:    event.UserId,
		GroupId:   event.GroupId,
		Text:      event.RawMessage,
		MessageId: event.MessageId,
	}
}

func (m MessageInfo) IsInGroup() bool {
	return m.GroupId != nil
}
