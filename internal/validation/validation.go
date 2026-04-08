package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type StructValidator struct {
	validate *validator.Validate
}

// New создаёт экземпляр валидатора приложения.
func New() *StructValidator {
	v := validator.New()

	// Подменяем имя поля в ошибках валидации.
	// Вместо имени Go-поля (например, ChatID) будет использоваться имя из json-тега (например, chatId).
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			return field.Name
		}

		name := strings.Split(tag, ",")[0]
		if name == "" {
			return field.Name
		}

		return name
	})

	return &StructValidator{
		validate: v,
	}
}

// Validate проверяет структуру по тегам validate.
func (v *StructValidator) Validate(out any) error {
	return v.validate.Struct(out)
}
