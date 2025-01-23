# qabot

qabot is for QQ AI Bot.

可以接入 OpenAI 和 Deepseek 的大语言模型 API 服务。

```console
Usage of ./qabot:
  -api-key string
        API key
  -api-url string
        大模型 API 服务的 url (default "https://api.deepseek.com/chat/completions")
  -endpoint string
        请求地址 (default "http://127.0.0.1:3000")
  -event-endpoint string
        onebot 上报事件地址 (default "127.0.0.1:8080")
  -group-prompt string
        群聊中给大语言模型的提示词 (default "你可能同时与多个用户（或者多个人格）聊天，你注意我说的每句话开头 [] 内的是不同的用户（人格）")
  -model string
        大语言模型 (default "deepseek-chat")
  -private-prompt string
        私聊中给大语言模型的提示词
  -whitelist string
        白名单文件路径（白名单文件可热更新） (default "whitelist.json")
```

