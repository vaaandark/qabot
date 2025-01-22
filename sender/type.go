package sender

type Data struct {
	Text string `json:"text,omitempty"`
	Qq   string `json:"qq,omitempty"`
	Id   string `json:"id,omitempty"`
}

type TypedMessage struct {
	Type string `json:"type,omitempty"`
	Data Data   `json:"data"`
}

type PrivateMessage struct {
	UserId  int64          `json:"user_id"`
	Message []TypedMessage `json:"message"`
}

type GroupMessage struct {
	GroupId int64          `json:"group_id"`
	Message []TypedMessage `json:"message"`
}

func NewPrivateMessage(userId int64, messageText string) PrivateMessage {
	return PrivateMessage{
		UserId: userId,
		Message: []TypedMessage{
			{
				Type: "text",
				Data: Data{
					Text: messageText,
				},
			},
		},
	}
}

func NewGroupMessage(groupId int64, messageText string, at *string, replyTo *string) GroupMessage {
	message := []TypedMessage{}

	if replyTo != nil {
		message = append(message, TypedMessage{
			Type: "reply",
			Data: Data{
				Id: *replyTo,
			},
		})
	}

	if at != nil {
		message = append(message, TypedMessage{
			Type: "at",
			Data: Data{
				Qq: *at,
			},
		})
	}

	message = append(message, TypedMessage{
		Type: "text",
		Data: Data{
			Text: messageText,
		},
	})

	return GroupMessage{
		GroupId: groupId,
		Message: message,
	}
}
