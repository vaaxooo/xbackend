package events

import (
	"bytes"
	"embed"
	htmpl "html/template"
	ttmpl "text/template"

	userevents "github.com/vaaxooo/xbackend/internal/modules/users/application/events"
)

const emailTemplateDateFormat = "02.01.2006 15:04"

//go:embed templates/*.txt templates/*.html
var emailTemplateFS embed.FS

type emailTemplates struct {
	confirmationText *ttmpl.Template
	confirmationHTML *htmpl.Template
	resetText        *ttmpl.Template
	resetHTML        *htmpl.Template
}

type templateData struct {
	Code    string
	Expires string
}

func mustLoadEmailTemplates() emailTemplates {
	return emailTemplates{
		confirmationText: ttmpl.Must(ttmpl.ParseFS(emailTemplateFS, "templates/confirm_email.txt")),
		confirmationHTML: htmpl.Must(htmpl.ParseFS(emailTemplateFS, "templates/confirm_email.html")),
		resetText:        ttmpl.Must(ttmpl.ParseFS(emailTemplateFS, "templates/reset_password.txt")),
		resetHTML:        htmpl.Must(htmpl.ParseFS(emailTemplateFS, "templates/reset_password.html")),
	}
}

func (t emailTemplates) renderConfirmation(evt userevents.EmailConfirmationRequested) (string, string, error) {
	data := templateData{Code: evt.Code, Expires: evt.ExpiresAt.Format(emailTemplateDateFormat)}
	return renderTemplates(t.confirmationText, t.confirmationHTML, data)
}

func (t emailTemplates) renderPasswordReset(evt userevents.PasswordResetRequested) (string, string, error) {
	data := templateData{Code: evt.Code, Expires: evt.ExpiresAt.Format(emailTemplateDateFormat)}
	return renderTemplates(t.resetText, t.resetHTML, data)
}

func renderTemplates(textTpl *ttmpl.Template, htmlTpl *htmpl.Template, data templateData) (string, string, error) {
	var textBuf bytes.Buffer
	if err := textTpl.Execute(&textBuf, data); err != nil {
		return "", "", err
	}

	var htmlBuf bytes.Buffer
	if err := htmlTpl.Execute(&htmlBuf, data); err != nil {
		return "", "", err
	}

	return textBuf.String(), htmlBuf.String(), nil
}

var defaultEmailTemplates = mustLoadEmailTemplates()
