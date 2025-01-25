package chatcontext

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"sort"
	"strings"
	"time"

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
	ReplyTo   *int32    `json:"reply_to,omitempty"`
	Message   Message   `json:"message"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

func NewContextNodeValue(replyTo *int32, message Message, timestamp time.Time) ContextNodeValue {
	return ContextNodeValue{
		ReplyTo:   replyTo,
		Message:   message,
		Timestamp: timestamp,
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

func (ck ContextNodeKey) Id() (id string) {
	if ck.GroupId != nil {
		id = fmt.Sprintf("group/%d", *ck.GroupId)
	} else if ck.UserId != nil {
		id = fmt.Sprintf("user/%d", *ck.UserId)
	}
	return
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

type DialogNode struct {
	Id        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	ReplyTo   *int32
	Timestamp time.Time
	Children  []*DialogNode `json:"children"`
}

func NewDialogNode(id, role, text string, replyTo *int32, timestamp time.Time, children []*DialogNode) *DialogNode {
	return &DialogNode{
		Id:        id,
		Role:      role,
		Content:   text,
		ReplyTo:   replyTo,
		Timestamp: timestamp,
		Children:  children,
	}
}

func maskLastFour(input string) string {
	runes := []rune(input)
	length := len(runes)
	switch {
	case length == 0:
		return ""
	case length <= 4:
		return strings.Repeat("x", length)
	default:
		return string(runes[:length-4]) + "xxxx"
	}
}

func (cc ChatContext) BuildIndexedDialogTrees(fuzzId bool) (map[string][]*DialogNode, error) {
	roots, err := cc.buildDialogTrees()
	if err != nil {
		return nil, err
	}

	indexedDialogTrees := make(map[string][]*DialogNode)
	for _, root := range roots {
		id := root.Id
		if fuzzId {
			id = maskLastFour(id)
		}
		trees, exist := indexedDialogTrees[id]
		if exist {
			trees = append(trees, root)
		} else {
			trees = []*DialogNode{root}
		}
		indexedDialogTrees[id] = trees
	}

	for key, nodes := range indexedDialogTrees {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Timestamp.After(nodes[j].Timestamp)
		})
		indexedDialogTrees[key] = nodes
	}

	return indexedDialogTrees, nil
}

func visitNode(key string, roots *[]*DialogNode, nodeMap map[string]*DialogNode) {
	node := nodeMap[key]
	if node.ReplyTo == nil {
		*roots = append(*roots, node)
	} else {
		parentKey := fmt.Sprintf("%s/%d", node.Id, *node.ReplyTo)
		if parentNode, exist := nodeMap[parentKey]; exist {
			parentNode.Children = append(parentNode.Children, node)
		}
	}
}

func (cc ChatContext) buildDialogTrees() ([]*DialogNode, error) {
	nodeMap := make(map[string]*DialogNode)
	iter := cc.db.NewIterator(nil, nil)

	for iter.Next() {
		key := string(iter.Key())
		var val ContextNodeValue
		if err := json.Unmarshal(iter.Value(), &val); err != nil {
			log.Printf("Failed to unmarshal: %v", err)
			continue
		}
		nodeMap[key] = NewDialogNode(path.Dir(key), val.Message.Role, val.Message.Content, val.ReplyTo, val.Timestamp, []*DialogNode{})
	}

	roots := []*DialogNode{}
	for key := range nodeMap {
		visitNode(key, &roots, nodeMap)
	}

	return roots, nil
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

func (cc ChatContext) AddContextNode(userId, groupId *int64, messageId int32, replyTo *int32, message Message, timestamp time.Time) error {
	key := NewContextNodeKey(userId, groupId, messageId).Key()
	val, err := NewContextNodeValue(replyTo, message, timestamp).Value()
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
