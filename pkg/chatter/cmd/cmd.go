package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/vaaandark/qabot/pkg/chatter/whitelist"
)

type Cmd struct {
	WhitelistAdaptor whitelist.Whitelist
}

func NewCmd(whitelistAdaptor whitelist.Whitelist) Cmd {
	return Cmd{
		WhitelistAdaptor: whitelistAdaptor,
	}
}

func (ca Cmd) IsAdmin(userId int64) bool {
	return ca.WhitelistAdaptor.IsAdmin(userId)
}

func (ca Cmd) cmdCheckHealth(_ int64, _ []string) (string, error) {
	return "1", nil
}

func (ca Cmd) cmdHelp(_ int64, _ []string) (string, error) {
	return "github.com/vaaandark/qabot 使用方式：\n\n" +
		"  - 新建上下文：\n" +
		"      - 群聊中：@bot 发送消息且该消息不是一条回复；\n" +
		"      - 私聊中：发送消息。\n\n" +
		"  - 继续聊天：回复 bot 的消息（无论是否 at），则从这条消息开始向上直到新建上下文的那条根消息都作为上下文。\n\n\n" +
		"假设现在有对话（q 开头表示用户提问，a 开头表示 bot 回答）：\n" +
		"  q1 -> a1 -> q2 -> a2 -> q3 -> a3\n\n" +
		"  - 如果回复 a3，则 q1 -> a1 -> q2 -> a2 -> q3 -> a3 作为上文；\n\n" +
		"  - 如果回复 a2，则 q1 -> a1 -> q2 -> a2 作为上文。\n\n" +
		"好处：\n" +
		"  1. 可以使用更多的上下文；\n" +
		"  2. 可以忽略不想要的上文\n\n\n" +
		"你也可以使用命令\n\n" +
		helpStr(), nil
}

func helpStr() string {
	return "Non-admin cmd:\n" +
		"    /help(/h)\n" +
		"    /check-health(/ch)\n" +
		"Admin cmd:\n" +
		"    /whitelist(/wl)"
}

func (ca *Cmd) cmdWhitelist(userId int64, cmds []string) (string, error) {
	if !ca.IsAdmin(userId) {
		return fmt.Sprintf("You(%d) are not administrator.", userId), nil
	}

	if len(cmds) < 2 {
		return fmt.Sprintf("%s: wrong args", cmds[0]), nil
	}

	switch cmds[1] {
	case "show":
		output, err := ca.WhitelistAdaptor.Show()
		if err != nil {
			return fmt.Sprintf("%s: failed to check whitelist: %v", cmds[0], err), err
		}
		return *output, nil
	case "add":
		if len(cmds) < 4 {
			return fmt.Sprintf("%s: wrong args", strings.Join(cmds[:2], " ")), nil
		}

		addedIds := []string{}
		switch cmds[2] {
		case "group":
			for _, idStr := range cmds[3:] {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					continue
				}
				if err := ca.WhitelistAdaptor.AddGroup(id); err == nil {
					addedIds = append(addedIds, strconv.FormatInt(id, 10))
				}
			}
		case "user":
			for _, idStr := range cmds[3:] {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					continue
				}
				if err := ca.WhitelistAdaptor.AddUser(id); err == nil {
					addedIds = append(addedIds, strconv.FormatInt(id, 10))
				}
			}
		default:
			return fmt.Sprintf("%s: wrong args", strings.Join(cmds[:3], " ")), nil
		}
		return fmt.Sprintf("Successfully added %s", strings.Join(addedIds, ", ")), nil
	default:
		return fmt.Sprintf("%s: unknown subcommand: %s", cmds[0], cmds[1]), nil
	}
}

func (ca *Cmd) Exec(userId int64, text string) (output string) {
	cmds := strings.Split(text, " ")
	if len(cmds) == 0 {
		output = "Empty cmd"
		return
	}

	switch cmds[0] {
	case "":
		return helpStr()
	case "wl", "whitelist":
		cmdOutput, err := ca.cmdWhitelist(userId, cmds)
		if err != nil {
			log.Printf("Failed to exec whitelist: %v", err)
		}
		output = cmdOutput
	case "h", "help":
		output, _ = ca.cmdHelp(userId, cmds)
	case "ch", "check-health":
		output, _ = ca.cmdCheckHealth(userId, cmds)
	default:
		output = "Unknown cmd"
	}

	return
}
