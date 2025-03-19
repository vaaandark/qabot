package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vaaandark/qabot/pkg/chatcontext"
	"github.com/vaaandark/qabot/pkg/messageenvelope"
	"github.com/vaaandark/qabot/pkg/onebot"
	"github.com/vaaandark/qabot/pkg/util"
)

type Sender struct {
	ToSendMessageCh chan messageenvelope.MessageEnvelope
	ChatContext     chatcontext.ChatContext
	Endpoint        string
	DialogEndpoint  string
}

func NewSender(toSendMessageCh chan messageenvelope.MessageEnvelope, chatContext chatcontext.ChatContext, endpoint, dialogEndpoint string) Sender {
	return Sender{
		ToSendMessageCh: toSendMessageCh,
		ChatContext:     chatContext,
		Endpoint:        endpoint,
		DialogEndpoint:  dialogEndpoint,
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

func splitThinkAndAnswer(text string) (string, string) {
	thinkLable := "</think>"
	if !strings.Contains(text, thinkLable) {
		return "", text
	}
	// 根据 </think> 分割
	before, after, _ := strings.Cut(text, thinkLable)
	before = strings.TrimSpace(before)
	after = strings.TrimSpace(after)
	return before, after
}

func (s Sender) doSend(m messageenvelope.MessageEnvelope) {
	var messageId int32
	var err error
	replyTo := strconv.Itoa(int(m.MessageId))

	think, answer := splitThinkAndAnswer(m.Text)
	if m.IsInGroup() {
		userIdStr := strconv.FormatInt(m.UserId, 10)
		if len(think) != 0 {
			forwardMessage := onebot.NewGroupForwordMessage(*m.GroupId, think)
			if _, err = s.doPost("send_group_forward_msg", forwardMessage); err != nil {
				log.Printf("Failed to send group forward message: group=%d, id=%d: %v", *m.GroupId, m.UserId, err)
			}
		}
		groupMessage := onebot.NewGroupMessage(s.DialogEndpoint, *m.GroupId, m.ModelName, answer, &userIdStr, &replyTo)
		if messageId, err = s.doPost("send_group_msg", groupMessage); err != nil {
			log.Printf("Failed to send group message: group=%d, id=%d: %v", *m.GroupId, m.UserId, err)
			return
		}
	} else {
		if len(think) != 0 {
			forwardMessage := onebot.NewPrivateForwordMessage(m.UserId, think)
			if _, err = s.doPost("send_private_forward_msg", forwardMessage); err != nil {
				log.Printf("Failed to send private forward message: id=%d: %v", m.UserId, err)
			}
		}
		privateMessage := onebot.NewPrivateMessage(s.DialogEndpoint, m.UserId, m.ModelName, answer, &replyTo)
		if messageId, err = s.doPost("send_private_msg", privateMessage); err != nil {
			log.Printf("Failed to send private message: id=%d: %v", m.UserId, err)
			return
		}
	}

	timestamp := time.Now()
	log.Printf("Cost %s to send message to %s: %s", timestamp.Sub(m.Timestamp), m.GetNamespacedGroupOrUserID(), util.TruncateLogStr(m.Text))
	m.Timestamp = timestamp

	// 不是命令回复才存档
	if !m.IsCmd {
		if messageId == 0 { // 被 QQ 拦截了，手动给它一个不会重复的值
			messageId = -int32(time.Now().UnixMicro())
		}
		if err := s.recordSent(messageId, m); err != nil {
			log.Printf("Failed to add user context: %v", err)
		}
	}
}
