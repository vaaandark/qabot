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
	FuzzId      bool
}

func NewDialogHtmlBuilder(chatContext chatcontext.ChatContext, fuzzId bool) DialogHtmlBuilder {
	return DialogHtmlBuilder{
		ChatContext: chatContext,
		FuzzId:      fuzzId,
	}
}

func (dhb DialogHtmlBuilder) buildDialogHtml(w http.ResponseWriter) error {
	indexedDialogTrees, err := dhb.ChatContext.BuildIndexedDialogTrees(dhb.FuzzId)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, indexedDialogTrees)
}

func (dhb DialogHtmlBuilder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := dhb.buildDialogHtml(w); err != nil {
		http.Error(w, fmt.Sprintf("Failed to build dialog html: %v", err), http.StatusInternalServerError)
	}
}
