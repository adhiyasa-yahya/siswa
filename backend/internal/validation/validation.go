// Package validation configures the validator used by Gin's binding.
package validation

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Setup makes validation errors report the JSON field name instead of the Go
// struct field name, so a failing `SKU string \`json:"sku"\`` is reported as
// "sku" (not "SKU" or "s_k_u") in the errors array.
func Setup() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		switch name {
		case "-":
			return ""
		case "":
			return strings.ToLower(field.Name)
		default:
			return name
		}
	})
}
