package chatter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/vaaandark/qabot/pkg/chatcontext"
	"github.com/vaaandark/qabot/pkg/chatter/cmd"
	"github.com/vaaandark/qabot/pkg/chatter/whitelist"
	"github.com/vaaandark/qabot/pkg/messageenvelope"
	"github.com/vaaandark/qabot/pkg/providerconfig"
	"golang.org/x/sync/semaphore"
)

type Chatter struct {
	ctx               context.Context
	ReceivedMessageCh chan messageenvelope.MessageEnvelope
	ToSendMessageCh   chan messageenvelope.MessageEnvelope
	WhitelistAdaptor  whitelist.Whitelist
	CmdAdaptor        cmd.Cmd
	ChatContext       *chatcontext.ChatContext
	Providers         []providerconfig.ProviderConfig
	MaxConcurrent     *semaphore.Weighted
}

func NewChatter(ctx context.Context, receiveMessageCh, toSendMessageCh chan messageenvelope.MessageEnvelope, whitelistFilePath string, chatContext *chatcontext.ChatContext, providers []providerconfig.ProviderConfig, maxConcurrentNum int64) (*Chatter, error) {
	wa, err := whitelist.NewWhitelist(whitelistFilePath)
	if err != nil {
		return nil, err
	}

	ca := cmd.NewCmd(*wa)

	return &Chatter{
		ctx:               ctx,
		ReceivedMessageCh: receiveMessageCh,
		ToSendMessageCh:   toSendMessageCh,
		WhitelistAdaptor:  *wa,
		CmdAdaptor:        ca,
		ChatContext:       chatContext,
		Providers:         providers,
		MaxConcurrent:     semaphore.NewWeighted(maxConcurrentNum),
	}, nil
}

func (c Chatter) Run(stopCh <-chan struct{}) {
	for {
		select {
		case m, ok := <-c.ReceivedMessageCh:
			if !ok {
				return
			}

			if m.IsInGroup() {
				if !c.WhitelistAdaptor.HasGroup(*m.GroupId) {
					continue
				}
			} else {
				if !c.WhitelistAdaptor.HasUser(m.UserId) {
					continue
				}
			}

			go c.doChat(m)
		case <-stopCh:
			return
		}
	}
}

func (c Chatter) doPost(messages []chatcontext.Message, provider *providerconfig.ProviderConfig) (*chatcontext.Message, error) {
	if provider == nil {
		return nil, fmt.Errorf("empty provider")
	}

	apiUrl := provider.Url
	apiModel := provider.Model
	apiKey := provider.NextKey()

	if provider.Reasoning && len(messages) > 0 {
		thinkLabel := "<think>"
		content := messages[len(messages)-1].Content
		messages[len(messages)-1].Content = content + thinkLabel
	}

	request := CompletionRequestFromContext(apiModel, messages)

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiUrl, bytes.NewReader(requestBytes))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	responseBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	response := CompletionResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	message := response.GetMessage()
	return message, nil
}

func (c Chatter) chatWithLlm(p providerconfig.ProviderConfig, m messageenvelope.MessageEnvelope) error {
	ctx, cancel := context.WithTimeout(c.ctx, time.Second*10)
	defer cancel()

	if err := c.MaxConcurrent.Acquire(ctx, 1); err != nil {
		return err
	}
	defer c.MaxConcurrent.Release(1)

	if c.ChatContext == nil {
		return nil
	}

	if m.ReplyTo != nil {
		// 回复的不是 bot 且没有 at bot 的情况，不关心！
		if !c.ChatContext.IsBotReply(&m.UserId, m.GroupId, *m.ReplyTo) && !m.IsAt {
			return nil
		}
	}

	err := c.ChatContext.AddContextNode(&m.UserId, m.GroupId, m.MessageId, m.ReplyTo, chatcontext.Message{
		Role:    "user",
		Content: m.Text,
	}, m.Timestamp)
	if err != nil {
		log.Printf("Failed to add user context: %v", err)
		return nil
	}

	messages, err := c.ChatContext.LoadContextMessages(&m.UserId, m.GroupId, m.MessageId)
	if err != nil {
		log.Printf("Failed to load context: %v", err)
		return nil
	}

	message, err := c.doPost(messages, &p)
	if err != nil {
		return err
	} else if message == nil {
		return fmt.Errorf("empty message")
	}

	content := strings.TrimSpace(message.Content)
	if len(content) == 0 {
		return fmt.Errorf("empty message")
	}

	m.Text = content
	m.ModelName = p.Name
	c.ToSendMessageCh <- m

	return nil
}

func (c *Chatter) execCmd(m messageenvelope.MessageEnvelope) {
	output := c.CmdAdaptor.Exec(m.UserId, m.Text)
	m.Text = output
	c.ToSendMessageCh <- m
}

func (c *Chatter) doChat(m messageenvelope.MessageEnvelope) {
	if m.IsCmd {
		c.execCmd(m)
	} else {
		for _, p := range c.Providers {
			if err := c.chatWithLlm(p, m); err != nil {
				log.Printf("Failed to chat with LLM: %v", err)
			} else {
				break
			}
		}
	}
}
