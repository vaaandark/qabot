package chatcontext

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ContextNodeKey struct {
	UserId    *int64
	GroupId   *int64
	MessageId int32
}

type ContextNodeValue struct {
	ReplyTo *int32  `json:"reply_to,omitempty"`
	Message Message `json:"message"`
}

func NewContextNodeValue(replyTo *int32, message Message) ContextNodeValue {
	return ContextNodeValue{
		ReplyTo: replyTo,
		Message: message,
	}
}

func (cv ContextNodeValue) Value() ([]byte, error) {
	return json.Marshal(cv)
}

func (cv ContextNodeValue) IsRoot() bool {
	return cv.ReplyTo == nil
}

func NewContextNodeKey(userId, groupId *int64, messageId int32) ContextNodeKey {
	return ContextNodeKey{
		UserId:    userId,
		GroupId:   groupId,
		MessageId: messageId,
	}
}

func (ck ContextNodeKey) Key() []byte {
	var s string
	if ck.GroupId != nil {
		s = fmt.Sprintf("group/%d/%d", *ck.GroupId, ck.MessageId)
	} else if ck.UserId != nil {
		s = fmt.Sprintf("user/%d/%d", *ck.UserId, ck.MessageId)
	}
	return []byte(s)
}

type ChatContext struct {
	db            *leveldb.DB
	privatePrompt string
	groupPrompt   string
}

func NewChatContext(db *leveldb.DB, privatePrompt, groupPrompt string) ChatContext {
	return ChatContext{
		db:            db,
		privatePrompt: privatePrompt,
		groupPrompt:   groupPrompt,
	}
}

func (cc ChatContext) IsBotReply(userId, groupId *int64, messageId int32) bool {
	val, err := cc.lookupContextNode(userId, groupId, messageId)
	if err != nil || val == nil {
		return false
	}
	return val.Message.Role == "assistant"
}

func (cc ChatContext) AddContextNode(userId, groupId *int64, messageId int32, replyTo *int32, message Message) error {
	key := NewContextNodeKey(userId, groupId, messageId).Key()
	val, err := NewContextNodeValue(replyTo, message).Value()
	if err != nil {
		return err
	}
	return cc.db.Put(key, val, nil)
}

func (cc ChatContext) lookupContextNode(userId, groupId *int64, messageId int32) (*ContextNodeValue, error) {
	key := NewContextNodeKey(userId, groupId, messageId).Key()
	b, err := cc.db.Get(key, nil)
	if err != nil {
		return nil, err
	}

	val := &ContextNodeValue{}
	if err := json.Unmarshal(b, val); err != nil {
		return nil, err
	}
	return val, nil
}

func (cc ChatContext) LoadContextMessages(userId, groupId *int64, messageId int32) ([]Message, error) {
	reversedMessages := []Message{}
	for {
		val, err := cc.lookupContextNode(userId, groupId, messageId)
		if err != nil {
			return nil, err
		}
		reversedMessages = append(reversedMessages, val.Message)
		if val.IsRoot() {
			break
		}
		messageId = *val.ReplyTo
	}

	messages := make([]Message, 0, len(reversedMessages)+1)

	if groupId != nil {
		if len(cc.groupPrompt) != 0 {
			messages = append(messages, Message{
				Role:    "system",
				Content: cc.groupPrompt,
			})
		}
	} else {
		if len(cc.privatePrompt) != 0 {
			messages = append(messages, Message{
				Role:    "system",
				Content: cc.privatePrompt,
			})
		}
	}

	for i := len(reversedMessages) - 1; i >= 0; i-- {
		messages = append(messages, reversedMessages[i])
	}
	return messages, nil
}
