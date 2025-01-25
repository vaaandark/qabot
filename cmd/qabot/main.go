package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/vaaandark/qabot/pkg/chatcontext"
	"github.com/vaaandark/qabot/pkg/chatter"
	"github.com/vaaandark/qabot/pkg/dialog"
	"github.com/vaaandark/qabot/pkg/messageenvelope"
	"github.com/vaaandark/qabot/pkg/receiver"
	"github.com/vaaandark/qabot/pkg/sender"
	"github.com/vaaandark/qabot/pkg/util"
	"golang.org/x/sync/errgroup"

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
	dialogEndpoint := flag.String("dialog-endpoint", "127.0.0.1:6060", "查看对话历史记录的地址")
	dialogAuthConfig := flag.String("dialog-auth-config", "dialog-auth-config.yaml", "查看对话历史记录认证的配置文件")
	dialogFuzzId := flag.Bool("dialog-fuzz-id", true, "查看对话历史记录时隐藏对话的群 ID 或用户 ID")

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

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		log.Printf("Event listening service starting on %s", *eventEndpoint)
		return http.ListenAndServe(*eventEndpoint, receiver.NewReceiver(receivedMessageCh))
	})

	auth, err := dialog.LoadAuthFromFile(*dialogAuthConfig)
	if err != nil {
		log.Printf("Failed to load auth config file: %v", err)
	} else {
		g.Go(func() error {
			log.Printf("Dialog service starting on %s", *dialogEndpoint)
			return http.ListenAndServe(*dialogEndpoint,
				dialog.RateLimiter(
					dialog.BasicAuth(auth,
						dialog.NewDialogHtmlBuilder(chatContext, auth, *dialogFuzzId))))
		})
	}

	if err := g.Wait(); err != nil {
		log.Fatalf("Service error: %v", err)
	}
}
