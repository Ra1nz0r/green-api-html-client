package server

import (
	"context"
	"fmt"
	"green-api-html-client/internal/config"
	"green-api-html-client/internal/handlers"
	"green-api-html-client/internal/logs"
	"green-api-html-client/internal/validation"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	swaggo "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/rs/zerolog/log"
)

const serviceName = "green-api-html-client"

// Run запускает HTTP-сервер и корректно завершает его по SIGINT/SIGTERM.
func Run() {
	// Загружаем конфигурацию приложения из TOML-файла или переменных окружения.
	cfg := config.Get()

	// Настраиваем глобальный логгер приложения.
	if err := logs.Setup(cfg.LoggerConfig(serviceName)); err != nil {
		panic(fmt.Errorf("failed setup logger: %w", err))
	}

	// Создаём HTTP-клиент для исходящих запросов в GREEN-API.
	httpClient := &http.Client{
		Timeout: cfg.GreenAPI.RequestTimeout,
	}

	// Создаём обработчики и передаём им все необходимые зависимости.
	h := handlers.New(
		httpClient,
		cfg.GreenAPI.BaseURL,
	)

	// Создаём и настраиваем экземпляр Fiber-приложения.
	app := newFiberApp(cfg)

	// Регистрируем маршруты приложения.
	registerRoutes(app, h)

	host := cfg.HTTPAddress()
	log.Info().
		Str("host", host).
		Msg("starting HTTP server")

	// Канал нужен, чтобы получить результат работы Listen:
	// либо сервер завершился штатно, либо остановился с ошибкой.
	serverErrCh := make(chan error, 1)

	// Запускаем HTTP-сервер в отдельной горутине.
	go func() {
		errListen := app.Listen(host)
		if errListen != nil {
			serverErrCh <- errListen
			return
		}
		serverErrCh <- nil
	}()

	// Подписываемся на сигналы завершения процесса.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// Ожидаем либо внешний сигнал остановки, либо неожиданное завершение сервера.
	select {
	case sig := <-sigCh:
		log.Info().
			Str("signal", sig.String()).
			Msg("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		// Пытаемся корректно завершить сервер, дав активным запросам время завершиться.
		if errShutdown := app.ShutdownWithContext(shutdownCtx); errShutdown != nil {
			log.Error().
				Err(errShutdown).
				Msg("failed to shutdown HTTP server gracefully")
		} else {
			log.Info().Msg("HTTP server shutdown completed")
		}

		// Дожидаемся завершения Listen() после shutdown.
		if errServe := <-serverErrCh; errServe != nil {
			log.Error().
				Err(errServe).
				Msg("server stopped with error after shutdown")
		}

	case errServe := <-serverErrCh:
		// Сюда попадаем, если сервер завершился сам до получения SIGINT/SIGTERM.
		// Обычно это означает ошибку запуска или неожиданную остановку.
		if errServe != nil {
			log.Fatal().
				Err(errServe).
				Msg("HTTP server stopped unexpectedly")
		}

		log.Info().Msg("HTTP server stopped")
	}
}

// newFiberApp создаёт и конфигурирует экземпляр Fiber-приложения.
func newFiberApp(cfg *config.Config) *fiber.App {
	// Создаём новый экземпляр валидатора.
	valid := validation.New()

	return fiber.New(fiber.Config{
		// Учитывать регистр символов в маршрутах.
		// Например, /API и /api будут считаться разными путями.
		CaseSensitive: true,

		// Учитывать завершающий слэш в маршрутах.
		// Например, /api/test и /api/test/ будут разными путями.
		StrictRouting: true,

		// Имя приложения. Может использоваться в служебных заголовках и логике фреймворка.
		AppName: "GREEN-API HTML Client",

		// Максимальное время чтения входящего запроса.
		ReadTimeout: cfg.Server.ReadTimeout,

		// Максимальное время записи ответа клиенту.
		WriteTimeout: cfg.Server.WriteTimeout,

		// Время жизни idle-соединения keep-alive без активных запросов.
		IdleTimeout: cfg.Server.IdleTimeout,

		// Инициализируем валидатор.
		StructValidator: valid,
	})
}

// registerRoutes регистрирует маршруты приложения.
func registerRoutes(app *fiber.App, h *handlers.Handler) {
	app.Get("/", handlers.IndexPage)
	app.Use("/static", static.New("./web/static"))

	app.Get("/swagger/*", swaggo.HandlerDefault)

	app.Get("/api/settings", h.GetSettings)
	app.Get("/api/state", h.GetStateInstance)
	app.Post("/api/message", h.SendMessage)
	app.Post("/api/file", h.SendFileByUrl)
}
