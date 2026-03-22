package auth

import (
	"net/http"

	"golang.org/x/text/language"
)

func limitBodySize(w http.ResponseWriter, r *http.Request) {
	const maxBodyBytes = 4096
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
}

func parseLocale(header string) string {
	if header == "" {
		return "en"
	}
	// Берём первый тег из "ru,en;q=0.9,de;q=0.8"
	tag, _ := language.MatchStrings(
		language.NewMatcher([]language.Tag{
			language.English, language.German, language.French,
		}),
		header,
	)
	base, _ := tag.Base()
	return base.String()
}
