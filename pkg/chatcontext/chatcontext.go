package chatcontext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/vaaandark/qabot/pkg/idmap"
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
	PrivatePrompt []Message
	GroupPrompt   []Message
}

type DialogNode struct {
	Id        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	MessageId int32
	ReplyTo   *int32
	Timestamp time.Time
	Children  []*DialogNode `json:"children"`
}

func NewDialogNode(id, role, text string, messageId int32, replyTo *int32, timestamp time.Time, children []*DialogNode) *DialogNode {
	return &DialogNode{
		Id:        id,
		Role:      role,
		Content:   text,
		MessageId: messageId,
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

type Dialogs struct {
	Welcome               string
	IndexedDialogTreesmap map[string][]*DialogNode
}

func (cc ChatContext) BuildIndexedDialogTrees(fuzzId bool, all bool, allowed []string, welcome string, idMap idmap.IdMap, specificId *string) (*Dialogs, error) {
	var allowedMap map[string]struct{}
	if !all {
		allowedMap = make(map[string]struct{})
		for _, a := range allowed {
			allowedMap[a] = struct{}{}
		}
	}

	roots, err := cc.buildDialogTrees()
	if err != nil {
		return nil, err
	}

	indexedDialogTrees := make(map[string][]*DialogNode)
	for _, root := range roots {
		id := root.Id
		if !all {
			if _, exist := allowedMap[id]; !exist {
				continue
			}
		}
		if specificId != nil {
			if *specificId != id {
				continue
			}
		}
		name := idMap.LookupName(id)
		if fuzzId {
			id = maskLastFour(id)
		}
		if name != nil {
			id = fmt.Sprintf("%s@%s", id, *name)
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

	return &Dialogs{
		Welcome:               welcome,
		IndexedDialogTreesmap: indexedDialogTrees}, nil
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
		messageId, err := strconv.Atoi(path.Base(key))
		if err != nil {
			log.Printf("Failed to unmarshal: %v", err)
			continue
		}
		nodeMap[key] = NewDialogNode(path.Dir(key), val.Message.Role, val.Message.Content, int32(messageId), val.ReplyTo, val.Timestamp, []*DialogNode{})
	}

	roots := []*DialogNode{}
	for key := range nodeMap {
		visitNode(key, &roots, nodeMap)
	}

	return roots, nil
}

func loadSystemPrompt(path string) ([]Message, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var messages []Message
	if err := json.Unmarshal(b, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func NewChatContext(db *leveldb.DB, privatePromptPath, groupPromptPath string) (*ChatContext, error) {
	privatePrompt, err := loadSystemPrompt(privatePromptPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load private prompt: %w", err)
	}
	groupPrompt, err := loadSystemPrompt(groupPromptPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load group prompt: %w", err)
	}
	return &ChatContext{
		db:            db,
		PrivatePrompt: privatePrompt,
		GroupPrompt:   groupPrompt,
	}, nil
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

func (cc ChatContext) lookupLatestMessageId(userId, groupId *int64) *int32 {
	iter := cc.db.NewIterator(nil, nil)
	defer iter.Release()

	start := []byte(path.Dir(string(NewContextNodeKey(userId, groupId, 0).Key())))
	end := bytes.Clone(start)
	end[len(end)-1] += 1

	latestTimestamp := time.Unix(0, 0)
	var messageId *int32
	for iter.Seek(start); iter.Valid() && bytes.Compare(iter.Key(), end) < 0; iter.Next() {
		var val ContextNodeValue
		if err := json.Unmarshal(iter.Value(), &val); err != nil {
			log.Printf("Failed to unmarshal: %v", err)
			continue
		}
		if val.Timestamp.After(latestTimestamp) {
			latestTimestamp = val.Timestamp
			n, err := strconv.ParseInt(path.Base(string(iter.Key())), 10, 32)
			if err != nil {
				log.Printf("Failed to parse number: %v", err)
				continue
			}
			n32 := int32(n)
			messageId = &n32
		}
	}

	return messageId
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

func (cc ChatContext) LoadContextLatestMessages(userId, groupId *int64) ([]Message, error) {
	latestMessageId := cc.lookupLatestMessageId(userId, groupId)
	if latestMessageId == nil {
		return nil, fmt.Errorf("latest message not exist")
	}
	return cc.LoadContextMessages(userId, groupId, *latestMessageId)
}

func BuildNicknamePrompt(nickname string) Message {
	return Message{
		Role:    "system",
		Content: fmt.Sprintf("正在和你对话的用户是 %s", nickname),
	}
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

	messages := make([]Message, 0, len(reversedMessages))
	for i := len(reversedMessages) - 1; i >= 0; i-- {
		messages = append(messages, reversedMessages[i])
	}
	return messages, nil
}
