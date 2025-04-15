package receiver

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/vaaandark/qabot/pkg/messageenvelope"
	"github.com/vaaandark/qabot/pkg/onebot"
	"github.com/vaaandark/qabot/pkg/util"
)

type Receiver struct {
	ReceivedMessageCh chan messageenvelope.MessageEnvelope
}

func NewReceiver(receivedMessageCh chan messageenvelope.MessageEnvelope) Receiver {
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

	event := onebot.Event{}
	err = json.Unmarshal(bodyBytes, &event)
	if err != nil {
		http.Error(w, "failed to unmarshal event", http.StatusInternalServerError)
		log.Printf("Failed to unmarshal event: %v", err)
		return
	}

	if event.IsMessage() {
		if text, replyTo, shouldBeIgnored, category, isAt := event.ProcessText(); !shouldBeIgnored {
			me := messageenvelope.FromEvent(event, &text, replyTo, category, isAt)
			log.Printf("Receive message from %s: %s", me.GetNamespacedGroupOrUserID(), util.TruncateLogStr(me.Text))
			receiver.ReceivedMessageCh <- me
		}
	}
}
