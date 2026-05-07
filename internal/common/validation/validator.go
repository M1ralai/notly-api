package validation

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/tr"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	tr_translations "github.com/go-playground/validator/v10/translations/tr"
)

var (
	validate *validator.Validate
	trans    ut.Translator
	once     sync.Once
)

func Init() {
	once.Do(func() {
		tr := tr.New()
		uni := ut.New(tr, tr)
		trans, _ = uni.GetTranslator("tr")

		validate = validator.New()

		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		tr_translations.RegisterDefaultTranslations(validate, trans)

		validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
			return ut.Add("required", "{0} alanı boş olamaz", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("required", fe.Field())
			return t
		})
	})
}

func Get() *validator.Validate {
	if validate == nil {
		Init()
	}
	return validate
}

func GetTranslator() ut.Translator {
	if validate == nil {
		Init()
	}
	return trans
}

func FormatErr(err error) string {
	if err == nil {
		return ""
	}

	if validate == nil {
		Init()
	}

	validatorErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}

	var errMsgs []string
	for _, e := range validatorErrs {
		errMsgs = append(errMsgs, e.Translate(trans))
	}

	return strings.Join(errMsgs, " | ")
}
