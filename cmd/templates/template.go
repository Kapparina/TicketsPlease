package templates

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/pkg/errors"
)

//go:embed ticket.gomd
var TicketTemplate string

//go:embed help.gomd
var HelpTemplate string

type TicketData struct {
	Category      string
	Username      string
	Subject       string
	Content       string
	Moderators    []string
	AttachmentURL string
}

type HelpData struct {
	CommandName string
	Version     string
}

func PopulateTicketData(data TicketData) (string, error) {
	var buf bytes.Buffer
	t := template.Must(template.New("ticket").Parse(TicketTemplate))
	if err := t.Execute(&buf, data); err != nil {
		return "", errors.WithMessage(err, "failed to execute ticket template")
	}
	return buf.String(), nil
}

func PopulateHelpData(data HelpData) (string, error) {
	var buf bytes.Buffer
	t := template.Must(template.New("help").Parse(HelpTemplate))
	if err := t.Execute(&buf, data); err != nil {
		return "", errors.WithMessage(err, "failed to execute help template")
	}
	return buf.String(), nil
}
