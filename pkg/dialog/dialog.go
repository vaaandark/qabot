package dialog

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/vaaandark/qabot/pkg/chatcontext"
	"github.com/vaaandark/qabot/pkg/idmap"
)

var dialogTreeHtmlTmpl, dialogListHtmlTmpl *template.Template

func init() {
	dialogTreeHtmlTmpl = template.Must(template.New("dialogTree").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			dict := make(map[string]interface{})
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}).Parse(dialogTreeHtmlTemplate))

	dialogListHtmlTmpl = template.Must(template.New("dialogList").Parse(dialogListHtmlTemplate))
}

type DialogHtmlBuilder struct {
	ChatContext chatcontext.ChatContext
	Auth        *Auth
	FuzzId      bool
	IdMap       idmap.IdMap
}

func NewDialogHtmlBuilder(chatContext chatcontext.ChatContext, auth *Auth, fuzzId bool, idMap idmap.IdMap) DialogHtmlBuilder {
	return DialogHtmlBuilder{
		ChatContext: chatContext,
		Auth:        auth,
		FuzzId:      fuzzId,
		IdMap:       idMap,
	}
}

func (dhb DialogHtmlBuilder) buildDialogTreeHtml(w http.ResponseWriter, all bool, user *User) error {
	if user == nil {
		return fmt.Errorf("User is not found")
	}
	indexedDialogTrees, err := dhb.ChatContext.BuildIndexedDialogTrees(dhb.FuzzId, all, user.Allowed, user.Welcome, dhb.IdMap)
	if err != nil {
		return err
	}
	return dialogTreeHtmlTmpl.Execute(w, indexedDialogTrees)
}

func (dhb DialogHtmlBuilder) buildDialogListHtml(w http.ResponseWriter, key string, all bool, user *User) error {
	if user == nil {
		return fmt.Errorf("User is not found")
	}

	splited := strings.Split(key, "/")
	if len(splited) != 3 {
		return fmt.Errorf("bad path")
	}

	userOrGroup := splited[0]
	var userId, groupId *int64
	if userOrGroup == "user" {
		id, err := strconv.ParseInt(splited[1], 10, 64)
		if err != nil {
			return err
		}
		userId = &id
	} else if userOrGroup == "group" {
		id, err := strconv.ParseInt(splited[1], 10, 64)
		if err != nil {
			return err
		}
		groupId = &id
	} else {
		return fmt.Errorf("bad path")
	}

	messageId, err := strconv.Atoi(splited[2])
	if err != nil {
		return err
	}

	shouldAllow := all
	if !all {
		for _, allowed := range user.Allowed {
			if allowed == path.Dir(key) {
				shouldAllow = true
			}
		}
	}

	if !shouldAllow {
		return fmt.Errorf("no permission")
	}

	messages, err := dhb.ChatContext.LoadContextMessages(userId, groupId, int32(messageId))
	if err != nil {
		return err
	}

	return dialogListHtmlTmpl.Execute(w, messages)
}

func (dhb DialogHtmlBuilder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name, password := getNameAndPassword(r)
	if name == nil || password == nil {
		http.Error(w, "Empty user name", http.StatusInternalServerError)
		return
	}

	all := dhb.Auth.isAdmin(*name, *password)
	user := dhb.Auth.getUser(*name)

	startTime := time.Now()
	defer func() {
		log.Printf("Cost %s to build dialog html for %s", time.Since(startTime), r.URL.Path)
	}()

	splited := strings.SplitN(r.URL.Path, "/", 2)
	if len(splited) < 2 || len(splited[1]) == 0 {
		if err := dhb.buildDialogTreeHtml(w, all, user); err != nil {
			http.Error(w, fmt.Sprintf("Failed to build dialog tree html: %v", err), http.StatusInternalServerError)
		}
	} else {
		if err := dhb.buildDialogListHtml(w, splited[1], all, user); err != nil {
			http.Error(w, fmt.Sprintf("Failed to build dialog list html: %v", err), http.StatusInternalServerError)
		}
	}
}
