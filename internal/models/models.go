package models

// GreenAPIAuthParams представляет параметры авторизации инстанса GREEN-API,
// которые пользователь вводит на странице для каждого запроса.
type GreenAPIAuthParams struct {
	IDInstance       string `json:"idInstance" validate:"required"`       // Идентификатор инстанса GREEN-API
	APITokenInstance string `json:"apiTokenInstance" validate:"required"` // Токен доступа к инстансу GREEN-API
}

// SendMessageRequest представляет входящее тело запроса от фронта
// на отправку текстового сообщения через GREEN-API.
type SendMessageRequest struct {
	GreenAPIAuthParams
	ChatID          string `json:"chatId" validate:"required"`            // Телефон или chatId, например 79991234567 или 79991234567@c.us
	Message         string `json:"message" validate:"required,max=20000"` // Текст сообщения
	QuotedMessageID string `json:"quotedMessageId,omitempty"`             // ID сообщения, на которое нужно ответить
}

// SendFileByURLRequest представляет входящее тело запроса от фронта
// на отправку файла по публичной ссылке через GREEN-API.
type SendFileByURLRequest struct {
	GreenAPIAuthParams
	ChatID          string `json:"chatId" validate:"required"`      // Телефон или chatId, например 79991234567 или 79991234567@c.us
	URLFile         string `json:"urlFile" validate:"required,url"` // Публичная ссылка на файл
	FileName        string `json:"fileName" validate:"required"`    // Имя файла с расширением, например file.pdf
	Caption         string `json:"caption,omitempty"`               // Необязательная подпись к файлу
	QuotedMessageID string `json:"quotedMessageId,omitempty"`       // ID сообщения, на которое нужно ответить
}

// GreenAPISendMessageRequest представляет тело запроса,
// которое отправляется непосредственно в метод sendMessage GREEN-API.
type GreenAPISendMessageRequest struct {
	ChatID          string `json:"chatId"`                    // Идентификатор чата
	Message         string `json:"message"`                   // Текст сообщения
	QuotedMessageID string `json:"quotedMessageId,omitempty"` // ID сообщения, на которое нужно ответить
}

// GreenAPISendFileByURLRequest представляет тело запроса,
// которое отправляется непосредственно в метод sendFileByUrl GREEN-API.
type GreenAPISendFileByURLRequest struct {
	ChatID          string `json:"chatId"`                    // Идентификатор чата
	URLFile         string `json:"urlFile"`                   // Публичная ссылка на файл
	FileName        string `json:"fileName"`                  // Имя файла с расширением
	Caption         string `json:"caption,omitempty"`         // Необязательная подпись к файлу
	QuotedMessageID string `json:"quotedMessageId,omitempty"` // ID сообщения, на которое нужно ответить
}
