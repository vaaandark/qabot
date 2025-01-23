package cmdadaptor

import (
	"fmt"
	"log"
	"qabot/chatter/whitelistadaptor"
	"strconv"
	"strings"
)

type CmdAdaptor struct {
	WhitelistAdaptor whitelistadaptor.WhitelistAdaptor
}

func NewCmdAdaptor(whitelistAdaptor whitelistadaptor.WhitelistAdaptor) CmdAdaptor {
	return CmdAdaptor{
		WhitelistAdaptor: whitelistAdaptor,
	}
}

func (ca CmdAdaptor) IsAdmin(userId int64) bool {
	return ca.WhitelistAdaptor.IsAdmin(userId)
}

func (ca *CmdAdaptor) cmdWhitelist(cmds []string) (string, error) {
	if len(cmds) < 2 {
		return fmt.Sprintf("%s: wrong args", cmds[0]), nil
	}

	switch cmds[1] {
	case "show":
		// do nothing
	case "add":
		if len(cmds) != 4 {
			return fmt.Sprintf("%s: wrong args", strings.Join(cmds[:2], " ")), nil
		}
		id, err := strconv.ParseInt(cmds[3], 10, 64)
		if err != nil {
			return fmt.Sprintf("%s: failed to parse id: %s", cmds[0], cmds[3]), nil
		}
		switch cmds[2] {
		case "group":
			if err := ca.WhitelistAdaptor.AddGroup(id); err != nil {
				return fmt.Sprintf("%s: failed to add group id: %s", cmds[0], cmds[3]), nil
			}
		case "user":
			if err := ca.WhitelistAdaptor.AddUser(id); err != nil {
				return fmt.Sprintf("%s: failed to add user id: %s", cmds[0], cmds[3]), nil
			}
		default:
			return fmt.Sprintf("%s: wrong args", strings.Join(cmds[:3], " ")), nil
		}
	default:
		return fmt.Sprintf("%s: unknown subcommand: %s", cmds[0], cmds[1]), nil
	}

	output, err := ca.WhitelistAdaptor.Show()
	log.Printf("inner output:%s", *output)
	if err != nil {
		return fmt.Sprintf("%s: failed to check whitelist: %v", cmds[0], err), err
	}

	return *output, nil
}

func (ca *CmdAdaptor) Exec(userId int64, text string) (output string, ok bool) {
	if !ca.IsAdmin(userId) {
		output = fmt.Sprintf("You(%d) are not administrator.", userId)
		return
	}

	cmds := strings.Split(text, " ")
	if len(cmds) == 0 {
		output = "Empty cmd"
		return
	}

	switch cmds[0] {
	case "wl", "whitelist":
		cmdOutput, err := ca.cmdWhitelist(cmds)
		if err != nil {
			log.Printf("Failed to exec whitelist: %v", err)
		} else {
			ok = true
		}
		output = cmdOutput
	}

	return
}
