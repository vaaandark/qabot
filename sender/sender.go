package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"qabot/chatcontext"
	"qabot/messageenvelope"
	"qabot/onebot"
	"strconv"
	"time"
)

type Sender struct {
	ToSendMessageCh chan messageenvelope.MessageEnvelope
	ChatContext     chatcontext.ChatContext
	Endpoint        string
}

func NewSender(toSendMessageCh chan messageenvelope.MessageEnvelope, chatContext chatcontext.ChatContext, endpoint string) Sender {
	return Sender{
		ToSendMessageCh: toSendMessageCh,
		ChatContext:     chatContext,
		Endpoint:        endpoint,
	}
}

func (s Sender) Run(stopCh <-chan struct{}) {
	for {
		select {
		case m, ok := <-s.ToSendMessageCh:
			if !ok {
				return
			}
			s.doSend(m)
		case <-stopCh:
			return
		}
	}
}

func (s Sender) doPost(path string, body interface{}) (int32, error) {
	url := fmt.Sprintf("%s/%s", s.Endpoint, path)

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))

	if err != nil {
		return 0, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	b, err = io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	response := onebot.SendResponse{}
	if err := json.Unmarshal(b, &response); err != nil {
		return 0, err
	}

	return response.Data.MessageId, nil
}

func (s Sender) recordSent(messageId int32, m messageenvelope.MessageEnvelope) error {
	return s.ChatContext.AddContextNode(m.TargetId, m.GroupId, messageId, &m.MessageId, chatcontext.Message{
		Role:    "assistant",
		Content: m.Text,
	}, m.Timestamp)
}

func (s Sender) doSend(m messageenvelope.MessageEnvelope) {
	var messageId int32
	var err error

	if m.IsInGroup() {
		replyTo := strconv.Itoa(int(m.MessageId))
		groupMessage := onebot.NewGroupMessage(*m.GroupId, m.Text, nil, &replyTo)
		if messageId, err = s.doPost("send_group_msg", groupMessage); err != nil {
			log.Printf("Failed to send group message: group=%d, id=%d: %v", *m.GroupId, m.UserId, err)
			return
		}
	} else {
		privateMessage := onebot.NewPrivateMessage(m.UserId, m.Text)
		if messageId, err = s.doPost("send_private_msg", privateMessage); err != nil {
			log.Printf("Failed to send private message: id=%d: %v", m.UserId, err)
			return
		}
	}

	timestamp := time.Now()
	log.Printf("Cost %s to send message: %s", timestamp.Sub(m.Timestamp), string(m.Text))
	m.Timestamp = timestamp

	// 不是命令回复才存档
	if !m.IsCmd {
		if err := s.recordSent(messageId, m); err != nil {
			log.Printf("Failed to add user context: %v", err)
		}
	}
}
