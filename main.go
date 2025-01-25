package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"qabot/chatcontext"
	"qabot/chatter"
	"qabot/messageenvelope"
	"qabot/receiver"
	"qabot/sender"
	"qabot/util"
	"strings"

	_ "net/http/pprof"

	"github.com/syndtr/goleveldb/leveldb"
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
	privatePrompt := flag.String("private-prompt", "", "私聊中给大语言模型的提示词")
	groupPrompt := flag.String("group-prompt", "", "群聊中给大语言模型的提示词")
	dbPath := flag.String("db", "context.db", "持久化存储上下文")
	userPprof := flag.Bool("pprof", false, "使用 pprof 性能分析")

	log.Printf("Command line args: %s", strings.Join(os.Args, ", "))
	flag.Parse()

	log.Printf("Whitelist path: %s", *whitelist)

	*endpoint = addHttpUrlPrefix(*endpoint)

	receivedMessageCh := make(chan messageenvelope.MessageEnvelope, 100)
	toSendMessageCh := make(chan messageenvelope.MessageEnvelope, 100)

	db, err := leveldb.OpenFile(*dbPath, nil)
	if err != nil {
		log.Panicf("Failed to open db: %v", err)
	}
	defer db.Close()

	chatContext := chatcontext.NewChatContext(db, *privatePrompt, *groupPrompt)

	c, err := chatter.NewChatter(receivedMessageCh, toSendMessageCh, *whitelist, &chatContext, *apiUrl, *apiKey, *model)
	if err != nil {
		log.Panicf("Failed to init chatter: %v", err)
	}
	s := sender.NewSender(toSendMessageCh, chatContext, *endpoint)

	stopCh := util.SetupSignalHandler()

	go c.Run(stopCh)
	go s.Run(stopCh)

	if *userPprof {
		go func() {
			log.Println("Pprof is now listening :6060")
			if err := http.ListenAndServe(":6060", nil); err != nil {
				log.Fatalf("Failed to start pprof: %v", err)
			}
		}()
	}

	handler := receiver.NewReceiver(receivedMessageCh)

	log.Fatal(http.ListenAndServe(*eventEndpoint, handler))
}
