package main

import (
	"bbbstatus/locales"
	"html/template"
	"io"
	"net"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// Render is a function hook to provide function calls inside html templates, such as Translate
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	currentLocalizer := i18n.NewLocalizer(locales.Bundle, c.Request().Header.Get("Accept-Language"))
	t.templates.Funcs(template.FuncMap{
		"t": func(text string) string {
			msg, err := currentLocalizer.LocalizeMessage(&i18n.Message{ID: text})
			if err != nil {
				println("[ERROR] Cannot translate text:", text)
			}
			return msg

		},
		"reverse": c.Echo().Reverse,
	})
	return t.templates.ExecuteTemplate(w, name, data)
}

func getIpFromContext(c echo.Context) net.IP {
	contextIp := c.RealIP()
	if strings.Contains(contextIp, "[") {
		endIndex := strings.Index(contextIp, "]:")
		return net.ParseIP(contextIp[1:endIndex])
	}
	if strings.Contains(contextIp, ":") {
		endIndex := strings.Index(contextIp, ":")
		return net.ParseIP(contextIp[:endIndex])
	}
	return net.ParseIP(contextIp)
}

// TranslateAdvanced is a simple locales function that supports text templating
func TranslateAdvanced(text string, data map[string]string) string {
	return locales.Localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: text, TemplateData: data})
}
