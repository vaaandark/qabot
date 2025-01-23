package chatcontext

import (
	"strings"
	"sync"
)

type ChatContext struct {
	sync.Mutex
	ContextMap    map[string]*Messages
	PrivatePrompt string
	GroupPrompt   string
}

func NewChatContext(privatePrompt, groupPrompt string) ChatContext {
	return ChatContext{
		Mutex:         sync.Mutex{},
		ContextMap:    make(map[string]*Messages),
		PrivatePrompt: privatePrompt,
		GroupPrompt:   groupPrompt,
	}
}

func (cc *ChatContext) GetContext(key string) *Messages {
	cc.Lock()
	defer cc.Unlock()

	context, exist := cc.ContextMap[key]
	if !exist {
		context = &Messages{}
		if strings.HasPrefix(key, "group") {
			if len(cc.GroupPrompt) != 0 {
				context.Data = append(context.Data, Message{
					Role:    "system",
					Content: cc.GroupPrompt,
				})
			}
		} else if strings.HasSuffix(key, "user") {
			if len(cc.PrivatePrompt) != 0 {
				context.Data = append(context.Data, Message{
					Role:    "system",
					Content: cc.PrivatePrompt,
				})
			}
		}
		cc.ContextMap[key] = context
	}

	return context
}

type Messages struct {
	sync.Mutex
	Data []Message
}

func (ms *Messages) AddUserMessage(content string) {
	ms.Lock()
	ms.Data = append(ms.Data, Message{Role: "user", Content: content})
}

func (ms *Messages) AddAssistantMessage(content string) {
	ms.Data = append(ms.Data, Message{Role: "assistant", Content: content})
	ms.Unlock()
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
