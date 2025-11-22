package mailsender

import (
	"bytes"
	"fmt"
	"html/template"
	texttemplate "text/template"
)

// RenderHTMLTemplate renders an HTML template with the given data.
// The templateStr should be a valid Go html/template string.
func RenderHTMLTemplate(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return buf.String(), nil
}

// RenderTextTemplate renders a plain text template with the given data.
// The templateStr should be a valid Go text/template string.
func RenderTextTemplate(templateStr string, data interface{}) (string, error) {
	tmpl, err := texttemplate.New("email").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse text template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute text template: %w", err)
	}

	return buf.String(), nil
}
