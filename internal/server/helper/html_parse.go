package helper

import (
	"bytes"
	"embed"
	"html/template"
)

func ParseEmailHTMLTemplate(tpl embed.FS, data interface{}) (buffBytes []byte, err error) {
	var body bytes.Buffer

	tmpl := template.Must(template.ParseFS(tpl, "template/*.html"))

	if err = tmpl.Execute(&body, data); err != nil {
		return
	}

	buffBytes = body.Bytes()
	return
}
