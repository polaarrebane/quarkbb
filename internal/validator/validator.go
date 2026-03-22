// Package validator wraps go-playground/validator
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/de"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/fr"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	de_translations "github.com/go-playground/validator/v10/translations/de"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	fr_translations "github.com/go-playground/validator/v10/translations/fr"
)

// Validator is a combination of a validator and a translator
type Validator interface {
	ValidateCommand(cmd any, locale string) error
}

// New creates a new Validator instance with support for multiple locales.
// It initializes the underlying validator and registers translators for
// English, French, and German languages.
func New() Validator {
	validate := validator.New(validator.WithRequiredStructEnabled())
	uni := registerTranslator(validate)
	validate.RegisterTagNameFunc(tagNameFunc)

	return &validatorImpl{
		Uni:      uni,
		Validate: validate,
	}
}

// ValidationError represents a single field validation error.
// It contains the field name and the localized error message.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents a collection of validation errors.
// It implements the error interface and provides a formatted error string.
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

type validatorImpl struct {
	Uni      *ut.UniversalTranslator
	Validate *validator.Validate
}

// ValidateCommand validates any API command struct using the configured validator.
// The cmd parameter must be a struct with validation tags.
func (v validatorImpl) ValidateCommand(cmd any, locale string) error {
	err := v.Validate.Struct(cmd)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return fmt.Errorf("validator internal: %w", err)
	}

	return createValidationErrors(ve, getTranslator(v.Uni, locale))
}

// Error returns a formatted string representation of all validation errors.
// Each error is formatted as "field: message" and separated by semicolons.
func (ve *ValidationErrors) Error() string {
	msgs := make([]string, 0, len(ve.Errors))
	for _, e := range ve.Errors {
		msgs = append(msgs, e.Field+": "+e.Message)
	}
	return strings.Join(msgs, "; ")
}

func tagNameFunc(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func registerTranslator(validate *validator.Validate) *ut.UniversalTranslator {
	locEN := en.New()
	locFR := fr.New()
	locDE := de.New()
	uni := ut.New(locEN, locEN, locFR, locDE)

	trans, _ := uni.GetTranslator("en")
	if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		_ = err // todo: log error
	}

	trans, _ = uni.GetTranslator("fr")
	if err := fr_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		_ = err // todo: log error
	}

	trans, _ = uni.GetTranslator("de")
	if err := de_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		_ = err // todo: log error
	}

	return uni
}

func createValidationErrors(ve validator.ValidationErrors, t ut.Translator) *ValidationErrors {
	errs := make([]ValidationError, 0, len(ve))
	for _, fe := range ve {
		errs = append(errs, ValidationError{
			Field:   fe.Field(),
			Message: fe.Translate(t),
		})
	}
	return &ValidationErrors{Errors: errs}
}

func getTranslator(ut *ut.UniversalTranslator, locale string) ut.Translator {
	if trans, ok := ut.GetTranslator(locale); ok {
		return trans
	}
	return ut.GetFallback()
}
