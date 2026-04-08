package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/joeshaw/envdecode"

	"green-api-html-client/internal/logs"
)

const (
	defaultConfigDir  = "."
	defaultConfigFile = "green-api-client_config.cfg"
)

var (
	ConfigPath = defaultConfigFile
)

var (
	once     sync.Once
	instance *Config
)

type Config struct {
	Server   ServerConfig   `toml:"server"`
	Log      LogConfig      `toml:"log"`
	GreenAPI GreenAPIConfig `toml:"green_api"`
}

type ServerConfig struct {
	Host            string        `toml:"host" env:"GREEN_HTML_SERVER_HOST"`                         // Адрес, на котором запускается HTTP-сервер
	Port            int           `toml:"port" env:"GREEN_HTML_SERVER_PORT"`                         // Порт HTTP-сервера
	ReadTimeout     time.Duration `toml:"read_timeout" env:"GREEN_HTML_SERVER_READ_TIMEOUT"`         // Максимальное время чтения входящего запроса
	WriteTimeout    time.Duration `toml:"write_timeout" env:"GREEN_HTML_SERVER_WRITE_TIMEOUT"`       // Максимальное время записи ответа клиенту
	IdleTimeout     time.Duration `toml:"idle_timeout" env:"GREEN_HTML_SERVER_IDLE_TIMEOUT"`         // Время жизни keep-alive соединения без активных запросов
	ShutdownTimeout time.Duration `toml:"shutdown_timeout" env:"GREEN_HTML_SERVER_SHUTDOWN_TIMEOUT"` // Максимальное время graceful shutdown
}

type LogConfig struct {
	Level  string `toml:"level" env:"GREEN_HTML_LOG_LEVEL"`     // Уровень логирования: trace, debug, info, warn, error
	Dir    string `toml:"dir" env:"GREEN_HTML_LOG_DIR"`         // Каталог для файлов логов
	Pretty bool   `toml:"pretty" env:"GREEN_HTML_LOG_PRETTY"`   // Человекочитаемый вывод в консоль
	ToFile bool   `toml:"to_file" env:"GREEN_HTML_LOG_TO_FILE"` // Писать ли логи в файл
	Caller bool   `toml:"caller" env:"GREEN_HTML_LOG_CALLER"`   // Добавлять file:line места вызова
}

type GreenAPIConfig struct {
	BaseURL string `toml:"base_url" env:"GREEN_HTML_GREEN_API_BASE_URL"` // Базовый URL GREEN-API, например https://1105.api.green-api.com
	RequestTimeout time.Duration `toml:"request_timeout" env:"GREEN_HTML_GREEN_API_REQUEST_TIMEOUT"` // Таймаут исходящих HTTP-запросов в GREEN-API
}

// Get возвращает единственный экземпляр конфигурации.
func Get() *Config {
	once.Do(func() {
		instance = new(Config)

		if err := load(instance); err != nil {
			log.Fatalf("config.Get(): failed to load configuration: %v", err)
		}

		setDefaults(instance)
	})

	return instance
}

// LoggerConfig возвращает конфигурацию логгера в формате,
// который ожидает пакет logs.
func (c *Config) LoggerConfig(serviceName string) logs.LoggerConfig {
	return logs.LoggerConfig{
		Service:             serviceName,
		IncludeServiceField: true,
		FileName:            serviceName + ".log",
		Level:               c.Log.Level,
		Pretty:              c.Log.Pretty,
		ToFile:              c.Log.ToFile,
		Dir:                 c.Log.Dir,
		Caller:              c.Log.Caller,
	}
}

// HTTPAddress возвращает адрес сервера в формате host:port.
func (c *Config) HTTPAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// load загружает конфигурацию из TOML-файла или, если файл не найден, из env.
func load(cfg *Config) error {
	f := ConfigPath

	if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
		if !strings.Contains(f, string(os.PathSeparator)) {
			f = defaultConfigDir + string(os.PathSeparator) + f
		}

		if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
			if err := envdecode.Decode(cfg); err != nil {
				return fmt.Errorf("load config from env: %w", err)
			}
			return nil
		}
	}

	if _, err := toml.DecodeFile(f, cfg); err != nil {
		return fmt.Errorf("load config from file: %w", err)
	}

	return nil
}

// setDefaults задаёт значения по умолчанию для необязательных полей.
func setDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 5 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 10 * time.Second
	}
	if cfg.Server.IdleTimeout == 0 {
		cfg.Server.IdleTimeout = 120 * time.Second
	}
	if cfg.Server.ShutdownTimeout == 0 {
		cfg.Server.ShutdownTimeout = 10 * time.Second
	}

	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Dir == "" {
		cfg.Log.Dir = "log"
	}

	if cfg.GreenAPI.RequestTimeout == 0 {
		cfg.GreenAPI.RequestTimeout = 15 * time.Second
	}
}
