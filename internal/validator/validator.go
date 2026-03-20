// Package validator wraps go-playground/validator
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/de"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/fr"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	de_translations "github.com/go-playground/validator/v10/translations/de"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	fr_translations "github.com/go-playground/validator/v10/translations/fr"
)

var (
	_v    *Validator
	_once sync.Once
)

func instance() *Validator {
	_once.Do(func() {
		_v = createValidator()
	})
	return _v
}

// Validator is a combination of a validator and a translator
type Validator struct {
	Uni      *ut.UniversalTranslator
	Validate *validator.Validate
}

// ValidationError is a combination of a field's name and a message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a set of validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// ValidateCommand is a function for validate any api command
func ValidateCommand(cmd any, locale string) error {
	v := instance()
	err := v.Validate.Struct(cmd)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return fmt.Errorf("validator internal: %w", err)
	}

	return createValidationErrors(ve, getTranslator(v, locale))
}

func (ve *ValidationErrors) Error() string {
	msgs := make([]string, 0, len(ve.Errors))
	for _, e := range ve.Errors {
		msgs = append(msgs, e.Field+": "+e.Message)
	}
	return strings.Join(msgs, "; ")
}

func createValidator() *Validator {
	validate := validator.New(validator.WithRequiredStructEnabled())
	uni := registerTranslator(validate)
	validate.RegisterTagNameFunc(tagNameFunc)

	return &Validator{
		Uni:      uni,
		Validate: validate,
	}
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
	translations := ve.Translate(t)
	errs := make([]ValidationError, 0, len(translations))
	for field, msg := range translations {
		errs = append(errs, ValidationError{
			Field:   field,
			Message: msg,
		})
	}
	return &ValidationErrors{Errors: errs}
}

func getTranslator(v *Validator, locale string) ut.Translator {
	if trans, ok := v.Uni.GetTranslator(locale); ok {
		return trans
	}
	return v.Uni.GetFallback()
}
