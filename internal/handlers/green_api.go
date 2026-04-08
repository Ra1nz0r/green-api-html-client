package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"green-api-html-client/internal/logs"
	"green-api-html-client/internal/models"
	"green-api-html-client/internal/validation"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Handler содержит зависимости, необходимые для работы HTTP-обработчиков.
type Handler struct {
	httpClient *http.Client
	baseURL    string
}

// New создаёт новый экземпляр Handler с переданными зависимостями.
func New(httpClient *http.Client, baseURL string) *Handler {
	return &Handler{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// GetSettings получает настройки инстанса из GREEN-API.
func (h *Handler) GetSettings(c fiber.Ctx) error {
	return h.doGreenAPIGet(c, "getSettings")
}

// GetStateInstance получает текущий статус инстанса из GREEN-API.
func (h *Handler) GetStateInstance(c fiber.Ctx) error {
	return h.doGreenAPIGet(c, "getStateInstance")
}

// SendMessage принимает запрос от клиента на отправку текстового сообщения,
// преобразует его в JSON и проксирует в метод sendMessage GREEN-API.
func (h *Handler) SendMessage(c fiber.Ctx) error {
	var req models.SendMessageRequest

	logger := log.With().
		Str("func", "SendMessage").
		Str("method", c.Method()).
		Str("path", c.Path()).
		Logger()

	// Bind декодирует JSON и, при наличии StructValidator,
	// автоматически запускает валидацию структуры.
	if err := c.Bind().JSON(&req); err != nil {
		return buildValidationErrorResponse(c, logger, err)
	}

	// Формируем тело запроса, которое ожидает непосредственно GREEN-API.
	payload := models.GreenAPISendMessageRequest{
		ChatID:          h.normalizeChatID(req.ChatID),
		Message:         req.Message,
		QuotedMessageID: req.QuotedMessageID,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to marshal send message request body")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to marshal request body",
		})
	}

	return h.doGreenAPIPost(
		c,
		requestBody,
		req.IDInstance,
		req.APITokenInstance,
		"sendMessage",
	)
}

// SendFileByUrl принимает запрос от клиента на отправку файла по ссылке,
// преобразует его в JSON и проксирует в метод sendFileByUrl GREEN-API.
func (h *Handler) SendFileByUrl(c fiber.Ctx) error {
	var req models.SendFileByURLRequest

	logger := log.With().
		Str("func", "SendFileByUrl").
		Str("method", c.Method()).
		Str("path", c.Path()).
		Logger()

	// Bind декодирует JSON и, при наличии StructValidator,
	// автоматически запускает валидацию структуры.
	if err := c.Bind().JSON(&req); err != nil {
		return buildValidationErrorResponse(c, logger, err)
	}

	// Формируем тело запроса, которое ожидает непосредственно GREEN-API.
	payload := models.GreenAPISendFileByURLRequest{
		ChatID:          h.normalizeChatID(req.ChatID),
		URLFile:         req.URLFile,
		FileName:        req.FileName,
		Caption:         req.Caption,
		QuotedMessageID: req.QuotedMessageID,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to marshal send file request body")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to marshal request body",
		})
	}

	return h.doGreenAPIPost(
		c,
		requestBody,
		req.IDInstance,
		req.APITokenInstance,
		"sendFileByUrl",
	)
}

// doGreenAPIGet выполняет GET-запрос к указанному методу GREEN-API
// и возвращает клиенту ответ внешнего сервиса как есть.
func (h *Handler) doGreenAPIGet(c fiber.Ctx, methodName string) error {
	nFunc, done := logs.LogFunction(methodName)
	defer done()

	idInstance := c.Query("idInstance", "")
	if idInstance == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "idInstance is required",
		})
	}

	apiTokenInstance := c.Query("apiTokenInstance", "")
	if apiTokenInstance == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "apiTokenInstance is required",
		})
	}

	logger := log.With().
		Str("func", nFunc).
		Str("method", c.Method()).
		Str("path", c.Path()).
		Str("green_api_method", methodName).
		Str("green_api_base_url", h.baseURL).
		Str("id_instance", idInstance).
		Logger()

	url := h.buildGreenAPIURL(idInstance, apiTokenInstance, methodName)

	logger.Info().Msg("starting request to GREEN-API")

	// Собираем исходящий HTTP-запрос во внешний сервис.
	req, err := http.NewRequestWithContext(c.Context(), http.MethodGet, url, nil)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to build request to GREEN-API")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to build request",
		})
	}

	// Выполняем запрос к GREEN-API.
	resp, err := h.httpClient.Do(req)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to send request to GREEN-API")

		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "failed to request GREEN-API",
		})
	}
	defer resp.Body.Close()

	// Читаем тело ответа целиком.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to read GREEN-API response body")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to read response body",
		})
	}

	if resp.StatusCode >= http.StatusBadRequest {
		logger.Warn().
			Int("status_code", resp.StatusCode).
			Int("body_size", len(body)).
			Msg("GREEN-API returned non-success status")
	} else {
		logger.Info().
			Int("status_code", resp.StatusCode).
			Int("body_size", len(body)).
			Msg("GREEN-API response received successfully")
	}

	// Возвращаем клиенту тот же статус и тело ответа,
	// которое пришло от GREEN-API.
	return c.Status(resp.StatusCode).
		Type("json").
		Send(body)
}

// doGreenAPIPost выполняет POST-запрос к указанному методу GREEN-API
// и возвращает клиенту статус и тело ответа внешнего сервиса без изменений.
func (h *Handler) doGreenAPIPost(
	c fiber.Ctx,
	requestBody []byte,
	idInstance string,
	apiTokenInstance string,
	methodName string,
) error {
	nFunc, done := logs.LogFunction(methodName)
	defer done()

	logger := log.With().
		Str("func", nFunc).
		Str("method", c.Method()).
		Str("path", c.Path()).
		Str("green_api_method", methodName).
		Str("green_api_base_url", h.baseURL).
		Str("id_instance", idInstance).
		Logger()

	url := h.buildGreenAPIURL(idInstance, apiTokenInstance, methodName)

	logger.Info().
		Int("request_body_size", len(requestBody)).
		Msg("starting POST request to GREEN-API")

	// Собираем исходящий HTTP-запрос во внешний сервис GREEN-API.
	req, err := http.NewRequestWithContext(
		c.Context(),
		http.MethodPost,
		url,
		bytes.NewReader(requestBody),
	)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to build request to GREEN-API")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to build request",
		})
	}

	// Указываем тип отправляемого тела запроса.
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос к GREEN-API.
	resp, err := h.httpClient.Do(req)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to send request to GREEN-API")

		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "failed to request GREEN-API",
		})
	}
	defer resp.Body.Close()

	// Читаем тело ответа от GREEN-API целиком.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to read GREEN-API response body")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to read response body",
		})
	}

	if resp.StatusCode >= http.StatusBadRequest {
		logger.Warn().
			Int("status_code", resp.StatusCode).
			Int("response_body_size", len(body)).
			Msg("GREEN-API returned non-success status")
	} else {
		logger.Info().
			Int("status_code", resp.StatusCode).
			Int("response_body_size", len(body)).
			Msg("GREEN-API response received successfully")
	}

	// Возвращаем клиенту тот же HTTP-статус и то же тело ответа,
	// которое пришло от GREEN-API.
	return c.Status(resp.StatusCode).
		Type("json").
		Send(body)
}

// buildValidationErrorResponse формирует и возвращает HTTP-ответ с ошибкой валидации.
// Если ошибка содержит список validator.ValidationErrors, функция собирает
// детализированный ответ по каждому невалидному полю и пишет warning в лог.
// Во всех остальных случаях возвращается общий ответ о некорректном теле запроса
// и ошибка пишется в лог.
func buildValidationErrorResponse(c fiber.Ctx, logger zerolog.Logger, err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		fieldErrors := make([]fiber.Map, 0, len(validationErrors))

		for _, e := range validationErrors {
			fieldErrors = append(fieldErrors, fiber.Map{
				"field": e.Field(),
				"tag":   e.Tag(),
				"param": e.Param(),
				"error": validation.ErrorMessage(e),
			})
		}

		logger.Warn().
			Int("fields_count", len(fieldErrors)).
			Interface("fields", fieldErrors).
			Msg("request validation failed")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "validation failed",
			"fields": fieldErrors,
		})
	}

	logger.Warn().
		Err(err).
		Msg("failed to parse request body")

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "invalid request body",
	})
}

// buildGreenAPIURL собирает URL для вызова метода GREEN-API.
func (h *Handler) buildGreenAPIURL(idInstance, apiTokenInstance, methodName string) string {
	return fmt.Sprintf(
		"%s/waInstance%s/%s/%s",
		h.baseURL,
		idInstance,
		methodName,
		apiTokenInstance,
	)
}

// normalizeChatID нормализует chatId.
// Если пользователь ввёл обычный номер телефона без суффикса,
// считаем это личным чатом и добавляем @c.us.
// Если уже передан полный chatId, возвращаем его как есть.
func (h *Handler) normalizeChatID(value string) string {
	value = strings.TrimSpace(value)

	if value == "" {
		return value
	}

	if strings.HasSuffix(value, "@c.us") || strings.HasSuffix(value, "@g.us") {
		return value
	}

	// Если пользователь уже ввёл какое-то значение с @,
	// но это не стандартный chatId, оставляем как есть.
	// В таком случае ошибку вернёт уже GREEN-API.
	if strings.Contains(value, "@") {
		return value
	}

	return value + "@c.us"
}
