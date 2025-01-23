# qabot

qabot is for QQ AI Bot.

可以接入 OpenAI 和 Deepseek 的大语言模型 API 服务。

## 部署

```console
Usage of qabot:
  -api-key string
        API key
  -api-url string
        大模型 API 服务的 url (default "https://api.deepseek.com/chat/completions")
  -db string
        持久化存储上下文 (default "context.db")
  -endpoint string
        请求地址 (default "http://127.0.0.1:3000")
  -event-endpoint string
        onebot 上报事件地址 (default "127.0.0.1:8080")
  -group-prompt string
        群聊中给大语言模型的提示词
  -model string
        大语言模型 (default "deepseek-chat")
  -private-prompt string
        私聊中给大语言模型的提示词
  -whitelist string
        白名单文件路径（白名单文件可热更新） (default "whitelist.json")
```

## 使用

qabot 使用方式：

- 新建上下文：
    - 群聊中：@bot 发送消息且该消息不是一条回复；
    - 私聊中：发送消息。
- 继续聊天：回复 bot 的消息（无论是否 at），则从这条消息开始向上直到新建上下文的那条根消息都作为上下文。

假设现在有对话（q 开头表示用户提问，a 开头表示 bot 回答）：`q1 -> a1 -> q2 -> a2 -> q3 -> a3`

- 如果回复 a3，则 `q1 -> a1 -> q2 -> a2 -> q3 -> a3` 作为上文；
- 如果回复 a2，则 `q1 -> a1 -> q2 -> a2` 作为上文。

好处：

1. 可以使用更多的上下文；
2. 可以忽略不想要的上文
