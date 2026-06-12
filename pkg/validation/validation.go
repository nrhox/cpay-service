package validation

import (
	"errors"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	id_translations "github.com/go-playground/validator/v10/translations/id"
)

const (
	alphaUpperNumericSpecialRegexString = "^[A-Z0-9_.]+$"
)

func ValidationAlpaUpperNumberSpecial(text string) bool {
	return regexp.MustCompile(alphaUpperNumericSpecialRegexString).MatchString(text)
}

func alpaUpperNumericSpecialChar(fl validator.FieldLevel) bool {
	return ValidationAlpaUpperNumberSpecial(fl.Field().String())
}

func registrationFunc(tag string, translation string, override bool) validator.RegisterTranslationsFunc {
	return func(ut ut.Translator) (err error) {
		if err = ut.Add(tag, translation, override); err != nil {
			return
		}

		return
	}
}

func translateFunc(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T(fe.Tag(), fe.Field())
	if err != nil {
		log.Printf("warning: error translating FieldError: %#v", fe)
		return fe.(error).Error()
	}

	return t
}

type translation struct {
	tag             string
	translation     string
	override        bool
	customRegisFunc validator.RegisterTranslationsFunc
	customTransFunc validator.TranslationFunc
}

func registerTranslation(v *validator.Validate, trans ut.Translator, translations []translation) (err error) {
	for _, t := range translations {
		if t.customTransFunc != nil && t.customRegisFunc != nil {
			err = v.RegisterTranslation(t.tag, trans, t.customRegisFunc, t.customTransFunc)
		} else if t.customTransFunc != nil && t.customRegisFunc == nil {
			err = v.RegisterTranslation(t.tag, trans, registrationFunc(t.tag, t.translation, t.override), t.customTransFunc)
		} else if t.customTransFunc == nil && t.customRegisFunc != nil {
			err = v.RegisterTranslation(t.tag, trans, t.customRegisFunc, translateFunc)
		} else {
			err = v.RegisterTranslation(t.tag, trans, registrationFunc(t.tag, t.translation, t.override), translateFunc)
		}

		if err != nil {
			return
		}
	}
	return
}

func registerTransCustomID(v *validator.Validate, trans ut.Translator) (err error) {
	translations := []translation{
		{
			tag:         "alpauppernum",
			translation: "{0} harus berupa huruf besar, angka, garis bawah, dan titik",
			override:    false,
		},
	}

	return registerTranslation(v, trans, translations)
}

type Validate struct {
	validator *validator.Validate
}

var ErrNotSupportLanguage = errors.New("not support the language")

func ParseMessage(s string) string {
	regexSnakeCase := regexp.MustCompile(`([a-zA-Z]+(?:_[a-zA-Z]+)*)`)

	s = regexSnakeCase.ReplaceAllStringFunc(s, func(match string) string {
		return strings.ReplaceAll(match, "_", " ")
	})

	regexCamelCase := regexp.MustCompile(`([a-z])(A-Z)`)
	s = regexCamelCase.ReplaceAllString(s, `$1 $2`)
	return strings.ToLower(s)
}

func New() *Validate {
	validate := validator.New()
	validate.RegisterValidation("alpauppernum", alpaUpperNumericSpecialChar)

	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		return field.Tag.Get("json")
	})

	return &Validate{
		validator: validate,
	}
}

type ErrorField struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (v *Validate) Struct(s any) []ErrorField {
	id := id.New()
	uni := ut.New(id, id)

	trans, _ := uni.GetTranslator("id")

	id_translations.RegisterDefaultTranslations(v.validator, trans)
	registerTransCustomID(v.validator, trans)

	err := v.validator.Struct(s)
	var errors []ErrorField

	if err != nil {
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			firstDot := strings.Index(e.Namespace(), ".")
			errors = append(errors, ErrorField{
				Field:   e.Namespace()[firstDot+1:],
				Message: ParseMessage(e.Translate(trans)),
			})
		}
	}

	if len(errors) > 0 && err != nil {
		return errors
	}
	return nil
}
