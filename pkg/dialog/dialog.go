package dialog

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/vaaandark/qabot/pkg/chatcontext"
)

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.New("dialog").Funcs(template.FuncMap{
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
	}).Parse(htmlTemplate))
}

type DialogHtmlBuilder struct {
	ChatContext chatcontext.ChatContext
	Auth        *Auth
	FuzzId      bool
}

func NewDialogHtmlBuilder(chatContext chatcontext.ChatContext, auth *Auth, fuzzId bool) DialogHtmlBuilder {
	return DialogHtmlBuilder{
		ChatContext: chatContext,
		Auth:        auth,
		FuzzId:      fuzzId,
	}
}

func (dhb DialogHtmlBuilder) buildDialogHtml(w http.ResponseWriter, all bool, user *User) error {
	if user == nil {
		return fmt.Errorf("User is not found")
	}
	indexedDialogTrees, err := dhb.ChatContext.BuildIndexedDialogTrees(dhb.FuzzId, all, user.Allowed, user.Welcome)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, indexedDialogTrees)
}

func (dhb DialogHtmlBuilder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name, password := getNameAndPassword(r)
	if name == nil || password == nil {
		http.Error(w, "Empty user name", http.StatusInternalServerError)
		return
	}

	all := dhb.Auth.isAdmin(*name, *password)
	user := dhb.Auth.getUser(*name)

	if err := dhb.buildDialogHtml(w, all, user); err != nil {
		http.Error(w, fmt.Sprintf("Failed to build dialog html: %v", err), http.StatusInternalServerError)
	}
}
