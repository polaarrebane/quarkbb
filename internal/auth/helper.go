package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	v "codeberg.org/ronia/quarkbb/internal/validator"
	"golang.org/x/text/language"
)

func decodeCommand(r *http.Request, cmd any) error {
	if err := json.NewDecoder(r.Body).Decode(cmd); err != nil {
		return errors.New("decoding error")
	}
	return nil
}

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

func decodeAndValidateCommand(r *http.Request, cmd any) (*response, error) {
	if err := decodeCommand(r, cmd); err != nil {
		return newMalformedBodyResponse(), err
	}

	locale := parseLocale(r.Header.Get("Accept-Language"))
	if err := v.ValidateCommand(cmd, locale); err != nil {
		var ve *v.ValidationErrors
		if errors.As(err, &ve) {
			return newValidationErrorResponse(ve), err
		}
		return newInternalErrorResponse(), err
	}

	return nil, nil
}
