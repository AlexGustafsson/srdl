package main

import (
	"bytes"
	texttemplate "text/template"
)

type ProgramTemplateValues struct {
	Name string
}

type SubscriptionTemplateValues struct {
	Artist string
	Album  string
}

type TemplateValues struct {
	Subscription SubscriptionTemplateValues
	Program      ProgramTemplateValues
}

// renderOutputPathTemplate will render the templated path as a go template.
func renderOutputPathTemplate(template string, values TemplateValues) (string, error) {
	tmpl, err := texttemplate.New("").Parse(template)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, values); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
