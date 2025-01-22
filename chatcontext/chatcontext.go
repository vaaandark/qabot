package chatcontext

import "sync"

type ChatContext struct {
	sync.Mutex
	ContextMap map[string]*Messages
	Prompt     string
}

func NewChatContext(prompt string) ChatContext {
	return ChatContext{
		Mutex:      sync.Mutex{},
		ContextMap: make(map[string]*Messages),
		Prompt:     prompt,
	}
}

func (cc *ChatContext) GetContext(key string) *Messages {
	cc.Lock()
	defer cc.Unlock()

	context, exist := cc.ContextMap[key]
	if !exist {
		context = &Messages{}
		if len(cc.Prompt) != 0 {
			context.Data = append(context.Data, Message{
				Role:    "system",
				Content: cc.Prompt,
			})
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
