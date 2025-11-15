package models

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func Validate(s any) error {
	return validate.Struct(s)
}
