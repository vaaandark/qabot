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
	"qabot/messageinfo"
	"strings"
)

type Chatter struct {
	ReceivedMessageCh chan messageinfo.MessageInfo
	ToSendMessageCh   chan messageinfo.MessageInfo
	WhitelistAdaptor  whitelistadaptor.WhitelistAdaptor
	CmdAdaptor        cmdadaptor.CmdAdaptor
	ChatContext       *chatcontext.ChatContext
	ApiUrl            string
	ApiKey            string
	Model             string
}

func NewChatter(receiveMessageCh, toSendMessageCh chan messageinfo.MessageInfo, whitelistFilePath string, chatContext *chatcontext.ChatContext, apiUrl, apiKey, model string) (*Chatter, error) {
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

func contextKeyFromMessageInfo(m *messageinfo.MessageInfo) string {
	if m.IsInGroup() {
		return fmt.Sprintf("group:%d", *m.GroupId)
	} else {
		return fmt.Sprintf("user:%d", m.UserId)
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
	return &message, nil
}

func (c Chatter) chatWithLlm(m messageinfo.MessageInfo) {
	if c.ChatContext == nil {
		return
	}

	context := c.ChatContext.GetContext(contextKeyFromMessageInfo(&m))
	context.AddUserMessage(m.Text)

	message, err := c.doPost(context.Data)
	if err != nil {
		context.Unlock()
		log.Printf("Failed to do post request: %v", err)
		return
	}

	m.Text = message.Content
	c.ToSendMessageCh <- m
	context.AddAssistantMessage(message.Content)
}

func (c *Chatter) execCmd(m messageinfo.MessageInfo) {
	output, _ := c.CmdAdaptor.Exec(m.UserId, strings.TrimPrefix(m.Text, "/"))
	m.Text = output
	c.ToSendMessageCh <- m
}

func (c *Chatter) doChat(m messageinfo.MessageInfo) {
	if m.IsCmd() {
		c.execCmd(m)
	} else {
		c.chatWithLlm(m)
	}
}
