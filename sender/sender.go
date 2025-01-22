package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"qabot/messageinfo"
	"strconv"
)

type Sender struct {
	ToSendMessageCh chan messageinfo.MessageInfo
	Endpoint        string
}

func NewSender(toSendMessageCh chan messageinfo.MessageInfo, endpoint string) Sender {
	return Sender{
		ToSendMessageCh: toSendMessageCh,
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

func (s Sender) post(path string, body interface{}) error {
	url := fmt.Sprintf("%s/%s", s.Endpoint, path)

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	log.Printf("Send message: %s", string(b))

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))

	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	if err != nil {
		return nil
	}

	return nil
}

func (s Sender) doSend(m messageinfo.MessageInfo) {
	if m.IsInGroup() {
		// at := strconv.FormatInt(m.UserId, 10)
		replyTo := strconv.Itoa(int(m.MessageId))
		groupMessage := NewGroupMessage(*m.GroupId, m.Text, nil, &replyTo)
		if err := s.post("send_group_msg", groupMessage); err != nil {
			log.Printf("Failed to send group message: group=%d, id=%d: %v", *m.GroupId, m.UserId, err)
		}
	} else {
		privateMessage := NewPrivateMessage(m.UserId, m.Text)
		if err := s.post("send_private_msg", privateMessage); err != nil {
			log.Printf("Failed to send private message: id=%d: %v", m.UserId, err)
		}
	}
}
