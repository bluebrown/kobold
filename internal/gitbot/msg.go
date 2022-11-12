package gitbot

import (
	"bytes"
	"text/template"

	"github.com/bluebrown/kobold/internal/krm"
)

type TemplateCommitMessenger struct {
	titleT *template.Template
	descrT *template.Template
}

func NewTemplateCommitMessenger(title, description string) TemplateCommitMessenger {
	return TemplateCommitMessenger{
		titleT: template.Must(template.New("").Parse(title)),
		descrT: template.Must(template.New("").Parse(description)),
	}
}

func (c TemplateCommitMessenger) Msg(changes []krm.Change) (title, description string, err error) {
	{
		var buf bytes.Buffer
		err := c.titleT.Execute(&buf, changes)
		if err != nil {
			return "", "", err
		}
		title = buf.String()
	}
	{
		var buf bytes.Buffer
		err := c.descrT.Execute(&buf, changes)
		if err != nil {
			return "", "", err
		}
		description = buf.String()
	}
	return title, description, nil
}
