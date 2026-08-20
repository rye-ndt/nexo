package helpers

import (
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var configValidator = newConfigValidator()

const structTag = "mapstructure"

func newConfigValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get(structTag), ",")[0]

		if name == "" || name == "-" {
			return field.Name
		}

		return name
	})

	v.RegisterValidation("model_name", func(fl validator.FieldLevel) bool {
		return enums.ModelName(fl.Field().String()).Valid()
	})

	v.RegisterValidation("effort", func(fl validator.FieldLevel) bool {
		return enums.Effort(fl.Field().String()).Valid()
	})

	v.RegisterValidation("input_type", func(fl validator.FieldLevel) bool {
		return enums.InputType(fl.Field().String()).Valid()
	})

	v.RegisterValidation("thinking_level", func(fl validator.FieldLevel) bool {
		return enums.ThinkingLevel(fl.Field().String()).Valid()
	})

	v.RegisterValidation("mcp_auth_flow", func(fl validator.FieldLevel) bool {
		return enums.MCPAuthFlow(fl.Field().String()).Valid()
	})

	return v
}

func ValidateStruct(target any) error {
	if target == nil {
		return custom_error.Critical("config nil")
	}

	return configValidator.Struct(target)
}
