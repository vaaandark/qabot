package main

import (
	"flag"
	"log"
	"net/http"
	"qabot/chatcontext"
	"qabot/chatter"
	"qabot/messageinfo"
	"qabot/nix"
	"qabot/receiver"
	"qabot/sender"
	"strings"
)

func addHttpUrlPrefix(url string) string {
	if !strings.HasPrefix(url, "http://") {
		url = "http://" + url
	}
	return url
}

func main() {
	eventEndpoint := flag.String("event-endpoint", "127.0.0.1:8080", "onebot 上报事件地址")
	endpoint := flag.String("endpoint", "http://127.0.0.1:3000", "请求地址")
	whitelist := flag.String("whitelist", "whitelist.json", "白名单文件路径（白名单文件可热更新）")
	apiUrl := flag.String("api-url", "https://api.deepseek.com/chat/completions", "大模型 API 服务的 url")
	apiKey := flag.String("api-key", "", "API key")
	model := flag.String("model", "deepseek-chat", "大语言模型")
	prompt := flag.String("prompt", "", "给大语言模型的提示词")

	flag.Parse()

	*endpoint = addHttpUrlPrefix(*endpoint)

	receivedMessageCh := make(chan messageinfo.MessageInfo, 100)
	toSendMessageCh := make(chan messageinfo.MessageInfo, 100)

	chatContext := chatcontext.NewChatContext(*prompt)

	c, err := chatter.NewChatter(receivedMessageCh, toSendMessageCh, *whitelist, &chatContext, *apiUrl, *apiKey, *model)
	if err != nil {
		log.Panicf("Failed to init chatter: %v", err)
	}
	s := sender.NewSender(toSendMessageCh, *endpoint)

	stopCh := nix.SetupSignalHandler()

	go c.Run(stopCh)
	go s.Run(stopCh)

	handler := receiver.NewReceiver(receivedMessageCh)

	log.Fatal(http.ListenAndServe(*eventEndpoint, handler))
}
