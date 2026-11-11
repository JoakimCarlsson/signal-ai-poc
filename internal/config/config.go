package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// How long the server is given. Shutdown drains in-flight requests,
// ReadHeader bounds how long a client may take to send its headers, Read how
// long it may take to send a whole request, Write how long a request may
// take from its headers to the end of its response, and Idle how long a
// kept-alive connection waits for the next request.
const (
	ShutdownTimeout   = 10 * time.Second
	ReadHeaderTimeout = 10 * time.Second
	ReadTimeout       = 30 * time.Second
	WriteTimeout      = 30 * time.Second
	IdleTimeout       = 2 * time.Minute
)

// The defaults for variables that are unset.
const (
	defaultAddr     = ":9901"
	defaultLogLevel = "info"
	defaultLogFmt   = "json"
	defaultService  = "shop"
)

// Config is everything the process reads from its environment.
type Config struct {
	Addr string // ADDR
	Log  Log
	OTel OTel
}

// Log is how the process writes its own log lines.
type Log struct {
	Level  string // LOG_LEVEL
	Format string // LOG_FORMAT: "json" or "text"
}

// OTel is where telemetry goes. Endpoint empty means no export: the process
// still logs to stdout, it just does not ship anything.
type OTel struct {
	ServiceName string // OTEL_SERVICE_NAME
	Endpoint    string // OTEL_EXPORTER_OTLP_ENDPOINT
	Headers     string // OTEL_EXPORTER_OTLP_HEADERS ("k=v,k=v")
}

// Load reads the configuration from the environment, after loading a .env
// file in the working directory when there is one. It returns an error when
// the .env file cannot be read.
func Load() (Config, error) {
	if err := loadDotenv(); err != nil {
		return Config{}, err
	}
	return Config{
		Addr: env("ADDR", defaultAddr),
		Log: Log{
			Level:  env("LOG_LEVEL", defaultLogLevel),
			Format: env("LOG_FORMAT", defaultLogFmt),
		},
		OTel: OTel{
			ServiceName: env("OTEL_SERVICE_NAME", defaultService),
			Endpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
			Headers:     os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"),
		},
	}, nil
}

// env returns the variable key, or fallback when it is unset or empty.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotenv loads .env into the environment without overriding variables
// already set. A missing file is not an error.
func loadDotenv() error {
	err := godotenv.Load()
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("config: reading .env: %w", err)
}
