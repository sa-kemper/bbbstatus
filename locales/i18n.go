package locales

import (
	"context"
	"embed"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

//go:embed active.*.toml
var LocaleFiles embed.FS

func TranslateFromCTX(ctx context.Context, messageID string) string {
	translatorFromCtx := ctx.Value("Translator")
	translator := translatorFromCtx.(i18n.Localizer)
	translatedMessage, err := translator.LocalizeMessage(&i18n.Message{ID: messageID})
	if err != nil {
		println("Could not translate message: " + messageID)
	}
	return translatedMessage
}
