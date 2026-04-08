package validation

import "github.com/go-playground/validator/v10"

// ErrorMessage возвращает человекочитаемый текст ошибки валидации.
func ErrorMessage(e validator.FieldError) string {
	switch e.Field() {
	case "idInstance":
		if e.Tag() == "required" {
			return "idInstance is required"
		}

	case "apiTokenInstance":
		if e.Tag() == "required" {
			return "apiTokenInstance is required"
		}

	case "chatId":
		if e.Tag() == "required" {
			return "chatId is required"
		}

	case "message":
		switch e.Tag() {
		case "required":
			return "message is required"
		case "max":
			return "message must not exceed " + e.Param() + " characters"
		}

	case "urlFile":
		switch e.Tag() {
		case "required":
			return "urlFile is required"
		case "url":
			return "urlFile must be a valid URL"
		}

	case "fileName":
		if e.Tag() == "required" {
			return "fileName is required"
		}
	}

	return e.Error()
}
