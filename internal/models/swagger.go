package models

// ErrorResponse представляет общий ответ с ошибкой.
type ErrorResponse struct {
	Error string `json:"error" example:"idInstance is required"`
}

// ValidationFieldError представляет ошибку валидации по одному полю.
type ValidationFieldError struct {
	Field string `json:"field" example:"chatId"`
	Tag   string `json:"tag" example:"required"`
	Param string `json:"param" example:""`
	Error string `json:"error" example:"chatId is required"`
}

// ValidationErrorResponse представляет ответ с ошибками валидации.
type ValidationErrorResponse struct {
	Error  string                 `json:"error" example:"validation failed"`
	Fields []ValidationFieldError `json:"fields"`
}
