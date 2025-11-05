package locales

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed active.*.toml
var LocaleFiles embed.FS

var Localizer *i18n.Localizer
var Bundle *i18n.Bundle

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
		currentTranslator := i18n.NewLocalizer(Bundle, c.Request().Header.Get("Accept-Language"), language.English.String())
		c.Set("Translator", currentTranslator)
		return next(c)
	}
}

func InitI18n(frontendTextMessages []i18n.Message) {
	Bundle = i18n.NewBundle(language.English)
	var err error
	for _, m := range frontendTextMessages {
		err = Bundle.AddMessages(language.English, &m)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to add message: %v\n", err)
		}
	}

	Bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	overwriteFolder := "overwrite/locales"

	if _, err = os.Stat(overwriteFolder); err == nil {
		// Use files from the overwrite folder if it exists
		files, err := os.ReadDir(overwriteFolder)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to read overwrite folder: %v\n", err)
		} else {
			for _, file := range files {
				if !file.IsDir() {
					path := overwriteFolder + "/" + file.Name()
					_, err = Bundle.LoadMessageFile(path)
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "Unable to load message file: %v\n", err)
					}
				}
			}
		}
	} else {
		// Walk the embedded locales and load each of them.
		files, err := LocaleFiles.ReadDir("locales")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to read locales: %v\n", err)
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			_, err = Bundle.LoadMessageFileFS(LocaleFiles, "locales/"+file.Name())
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Unable to load embedded message file: %v\n", err)
			}
		}
	}
}
