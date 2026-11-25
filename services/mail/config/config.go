// Package config loads the mail service configuration from environment
// variables. Names are listed in docs/env-name.txt. Values are never logged.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	RabbitMQ RabbitMQ
	SMTP     SMTP
	Worker   Worker
}

type RabbitMQ struct {
	Host     string
	Port     string
	User     string
	Password string
	Queue    string // bound to MAIL_EXCHANGE by the dev
}

// URL builds the AMQP connection URL. Contains credentials: never log it.
func (r RabbitMQ) URL() string {
	u := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(r.User, r.Password),
		Host:   net.JoinHostPort(r.Host, r.Port),
		Path:   "/",
	}
	return u.String()
}

type SMTP struct {
	Host     string
	Port     string
	Username string
	Password string // app password
	From     string
}

// Worker holds processing policy. These are defaults, not read from the
// environment yet, so no variable names are invented for them.
type Worker struct {
	Count          int           // tasks handled in parallel
	RetryDelay     time.Duration // wait before requeueing a failed send
	ReconnectDelay time.Duration // wait before reconnecting to RabbitMQ
	SMTPTimeout    time.Duration // one whole SMTP conversation
}

func defaultWorker() Worker {
	return Worker{
		Count:          2,
		RetryDelay:     5 * time.Second,
		ReconnectDelay: 5 * time.Second,
		SMTPTimeout:    30 * time.Second,
	}
}

// Load reads configuration from the environment. It reports every missing or
// malformed variable at once.
func Load() (Config, error) {
	var l loader
	cfg := Config{
		RabbitMQ: RabbitMQ{
			Host:     l.str("RABBITMQ_HOST"),
			Port:     l.port("RABBITMQ_PORT"),
			User:     l.str("RABBITMQ_USER"),
			Password: l.str("RABBITMQ_PASS"),
			Queue:    l.str("MAIL_QUEUE"),
		},
		SMTP: SMTP{
			Host:     l.str("SMTP_HOST"),
			Port:     l.port("SMTP_PORT"),
			Username: l.str("SMTP_USERNAME"),
			Password: l.str("SMTP_PASSWORD"),
			From:     l.str("SMTP_FROM"),
		},
		Worker: defaultWorker(),
	}
	if err := errors.Join(l.errs...); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

type loader struct {
	errs []error
}

func (l *loader) str(name string) string {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		l.errs = append(l.errs, fmt.Errorf("%s is required", name))
	}
	return v
}

func (l *loader) port(name string) string {
	v := l.str(name)
	if v == "" {
		return ""
	}
	if n, err := strconv.Atoi(v); err != nil || n <= 0 || n > 65535 {
		l.errs = append(l.errs, fmt.Errorf("%s must be a port number", name))
	}
	return v
}
