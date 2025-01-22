package event

import "strings"

// TODO: array message type
type Event struct {
	Time        int64  `json:"time"`
	SelfId      int64  `json:"self_id"`
	PostType    string `json:"post_type"`
	MessageType string `json:"message_type"`
	SubType     string `json:"sub_type"`
	MessageId   int32  `json:"message_id"`
	UserId      int64  `json:"user_id"`
	RawMessage  string `json:"raw_message"`
	Sender      Sender
	GroupId     *int64 `json:"group_id,omitempty"`
	// TODO: message
}

func (e Event) IsFromSelf() bool {
	return e.SelfId == e.UserId
}

func (e Event) IsMessage() bool {
	return e.PostType == "message"
}

func (e Event) ShouldBeIgnore() bool {
	return !strings.HasPrefix(e.RawMessage, "v ")
}

func (e Event) TrimPrefix() Event {
	e.RawMessage = strings.TrimPrefix(e.RawMessage, "v ")
	return e
}

type Sender struct {
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	GroupId  *int64 `json:"group_id,omitempty"`
}
