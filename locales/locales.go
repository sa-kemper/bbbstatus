package locales

import (
	"context"
	"embed"

	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed active.*.toml
var LocaleFiles embed.FS

// TranslateFromCTX makes translating for individual requests thread safe.
func TranslateFromCTX(ctx context.Context, messageID string) string {
	translatorFromCtx := ctx.Value("Translator")
	translator := translatorFromCtx.(i18n.Localizer)
	translatedMessage, err := translator.LocalizeMessage(&i18n.Message{ID: messageID})
	if err != nil {
		println("Could not translate message: " + messageID)
	}
	return translatedMessage
}

func TranslateFromEchoContext(ctx echo.Context, messageID string) string {
	translatorFromCtx := ctx.Get("Translator")
	translator := translatorFromCtx.(i18n.Localizer)
	translatedMessage, err := translator.LocalizeMessage(&i18n.Message{ID: messageID})
	if err != nil {
		println("Could not translate message: " + messageID)
	}
	return translatedMessage

}

func AddTranslatorToContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("Translator", i18n.NewLocalizer(Bundle, c.Request().Header.Get("Accept-Language"), language.English.String()))
		return next(c)
	}
}
