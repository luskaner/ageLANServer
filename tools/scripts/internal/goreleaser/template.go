package goreleaser

import (
	"bytes"
	"text/template"

	"github.com/google/uuid"
)

type Renders[D any] interface {
	Render(data D) string
}
type LiteralString[D any] string

func (t LiteralString[D]) Render(_ D) string {
	return string(t)
}

type Template[D any] struct {
	tmpl *template.Template
	text string
}

func NewTemplate[D any](text string) *Template[D] {
	tmpl := template.New(uuid.NewString())
	return &Template[D]{tmpl: tmpl, text: text}
}

func (t *Template[D]) Render(data D) string {
	if tmpl, err := t.tmpl.Parse(t.text); err != nil {
		return ""
	} else {
		var buf bytes.Buffer
		if err = tmpl.Execute(&buf, data); err != nil {
			buf.WriteString("")
		}
		return buf.String()
	}
}
