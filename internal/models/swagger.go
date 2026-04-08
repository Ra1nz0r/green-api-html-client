package models

// BadRequestErrorResponse представляет ответ с ошибкой клиента,
// когда в запросе отсутствуют обязательные параметры или переданы некорректные данные.
type BadRequestErrorResponse struct {
	Error string `json:"error" example:"idInstance is required"`
}

// InternalErrorResponse представляет ответ с внутренней ошибкой сервиса,
// например при сборке запроса или чтении ответа от GREEN-API.
type InternalErrorResponse struct {
	Error string `json:"error" example:"failed to build request"`
}

// BadGatewayErrorResponse представляет ответ, когда сервис
// не смог успешно выполнить запрос к внешнему GREEN-API.
type BadGatewayErrorResponse struct {
	Error string `json:"error" example:"failed to request GREEN-API"`
}

// ValidationFieldError представляет ошибку валидации по одному полю.
type ValidationFieldError struct {
	Field string `json:"field" example:"chatId"`
	Tag   string `json:"tag" example:"required"`
	Param string `json:"param" example:""`
	Error string `json:"error" example:"chatId is required"`
}

// ValidationErrorResponse представляет ответ с ошибками валидации
// входного тела запроса.
type ValidationErrorResponse struct {
	Error  string                 `json:"error" example:"validation failed"`
	Fields []ValidationFieldError `json:"fields"`
}

// GetStateInstanceResponse представляет успешный ответ метода getStateInstance.
type GetStateInstanceResponse struct {
	StateInstance string `json:"stateInstance" example:"authorized"`
}

// MessageIDResponse представляет успешный ответ методов sendMessage
// и sendFileByUrl, содержащий идентификатор созданного сообщения.
type MessageIDResponse struct {
	IDMessage string `json:"idMessage" example:"3EB00BA9F12E0A8028047E"`
}
