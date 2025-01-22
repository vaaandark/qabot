package receiver

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"qabot/event"
	"qabot/messageinfo"
)

type Receiver struct {
	ReceivedMessageCh chan messageinfo.MessageInfo
}

func NewReceiver(receivedMessageCh chan messageinfo.MessageInfo) Receiver {
	return Receiver{
		ReceivedMessageCh: receivedMessageCh,
	}
}

func (receiver Receiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read http body", http.StatusInternalServerError)
		log.Printf("Failed to read http body: %v", err)
		return
	}
	defer r.Body.Close()

	event := event.Event{}
	err = json.Unmarshal(bodyBytes, &event)
	if err != nil {
		http.Error(w, "failed to unmarshal event", http.StatusInternalServerError)
		log.Printf("Failed to unmarshal event: %v", err)
		return
	}

	log.Printf("%v", event)

	if !event.IsFromSelf() && event.IsMessage() && !event.ShouldBeIgnore() {
		receiver.ReceivedMessageCh <- messageinfo.FromEvent(event.TrimPrefix())
	}
}
