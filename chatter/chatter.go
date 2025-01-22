package chatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"qabot/chatcontext"
	"qabot/chatter/whitelistadaptor"
	"qabot/messageinfo"
)

type Chatter struct {
	ReceivedMessageCh chan messageinfo.MessageInfo
	ToSendMessageCh   chan messageinfo.MessageInfo
	WhitelistAdaptor  whitelistadaptor.WhitelistAdaptor
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

	return &Chatter{
		ReceivedMessageCh: receiveMessageCh,
		ToSendMessageCh:   toSendMessageCh,
		WhitelistAdaptor:  *wa,
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

func (c Chatter) doChat(m messageinfo.MessageInfo) {
	if c.ChatContext != nil {
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
	} else {
		messageText := m.Text
		messageText = fmt.Sprintf("【复读】%s", messageText)
		m.Text = messageText
		c.ToSendMessageCh <- m
	}
}
