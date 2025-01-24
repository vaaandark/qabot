package chatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"qabot/chatcontext"
	"qabot/chatter/cmdadaptor"
	"qabot/chatter/whitelistadaptor"
	"qabot/messageenvelope"
)

type Chatter struct {
	ReceivedMessageCh chan messageenvelope.MessageEnvelope
	ToSendMessageCh   chan messageenvelope.MessageEnvelope
	WhitelistAdaptor  whitelistadaptor.WhitelistAdaptor
	CmdAdaptor        cmdadaptor.CmdAdaptor
	ChatContext       *chatcontext.ChatContext
	ApiUrl            string
	ApiKey            string
	Model             string
}

func NewChatter(receiveMessageCh, toSendMessageCh chan messageenvelope.MessageEnvelope, whitelistFilePath string, chatContext *chatcontext.ChatContext, apiUrl, apiKey, model string) (*Chatter, error) {
	wa, err := whitelistadaptor.NewWhitelistAdaptor(whitelistFilePath)
	if err != nil {
		return nil, err
	}

	ca := cmdadaptor.NewCmdAdaptor(*wa)

	return &Chatter{
		ReceivedMessageCh: receiveMessageCh,
		ToSendMessageCh:   toSendMessageCh,
		WhitelistAdaptor:  *wa,
		CmdAdaptor:        ca,
		ChatContext:       chatContext,
		ApiUrl:            apiUrl,
		ApiKey:            apiKey,
		Model:             model,
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

func (c Chatter) doPost(messages []chatcontext.Message) (*chatcontext.Message, error) {
	request := CompletionRequestFromContext(c.Model, messages)

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", c.ApiUrl, bytes.NewReader(requestBytes))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiKey))
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

func (c Chatter) chatWithLlm(m messageenvelope.MessageEnvelope) {
	if c.ChatContext == nil {
		return
	}

	if m.ReplyTo != nil {
		// 回复的不是 bot 且没有 at bot 的情况，不关心！
		if !c.ChatContext.IsBotReply(&m.UserId, m.GroupId, *m.ReplyTo) && !m.IsAt {
			return
		}
	}

	err := c.ChatContext.AddContextNode(&m.UserId, m.GroupId, m.MessageId, m.ReplyTo, chatcontext.Message{
		Role:    "user",
		Content: m.Text,
	})
	if err != nil {
		log.Printf("Failed to add user context: %v", err)
		return
	}

	messages, err := c.ChatContext.LoadContextMessages(&m.UserId, m.GroupId, m.MessageId)
	if err != nil {
		log.Printf("Failed to load context: %v", err)
		return
	}

	message, err := c.doPost(messages)
	if err != nil {
		log.Printf("Failed to do post request: %v", err)
		return
	} else if message == nil {
		log.Print("Empty message")
		return
	}

	m.Text = message.Content
	c.ToSendMessageCh <- m
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
		c.chatWithLlm(m)
	}
}
