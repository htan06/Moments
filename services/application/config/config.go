// Package config loads application configuration from environment variables.
// Variable names are listed in docs/env-name.txt. Values are never logged.
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
	HTTP     HTTP
	Postgres Postgres
	Redis    Redis
	RabbitMQ RabbitMQ
	ObjStore ObjStore
	Token    Token
	Account  Account
}

// Account holds the account module's policy values. Defaults are the business
// defaults from docs/domain.md.
type Account struct {
	OTPTTL            time.Duration // ACCOUNT-R2: verification session lifetime
	OTPResendCooldown time.Duration // ACCOUNT-R2: gap between two OTP sends
	OTPMaxAttempts    int           // ACCOUNT-R2: wrong OTPs that delete the session
	LoginMaxAttempts  int           // ACCOUNT-R3: failed logins before blocking
	LoginWindow       time.Duration // ACCOUNT-R3: counting / blocking window
	BcryptCost        int
	ReservedUsernames []string // ACCOUNT-R5

	MailExchange   string // RabbitMQ exchange of the mail service
	MailRoutingKey string // routing key of OTP mail tasks
}

func defaultAccount() Account {
	return Account{
		OTPTTL:            10 * time.Minute,
		OTPResendCooldown: 60 * time.Second,
		OTPMaxAttempts:    5,
		LoginMaxAttempts:  5,
		LoginWindow:       15 * time.Minute,
		BcryptCost:        12,
		ReservedUsernames: []string{"admin", "moments", "support", "api"},
	}
}

type HTTP struct {
	Addr string
}

type Postgres struct {
	URI      string
	MaxConns int32
	MinConns int32
}

type Redis struct {
	Addr     string
	Password string
}

type RabbitMQ struct {
	Host     string
	Port     string
	User     string
	Password string
	// PublishChannels is how many AMQP channels the publisher may open on its
	// single connection, i.e. how many publishes run in parallel.
	PublishChannels int
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

type ObjStore struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketRaw       string
	PublicBaseURL   string // CDN in front of the media bucket
}

type Token struct {
	PrivateKeyPath string
	PublicKeyPath  string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
}

// Load reads configuration from the environment. It reports every missing or
// malformed variable at once.
func Load() (Config, error) {
	var l loader
	cfg := Config{
		HTTP: HTTP{
			Addr: l.str("APPLICATION_ADDRESS"),
		},
		Postgres: Postgres{
			URI:      l.str("DATABASE_POSTGRES_URI"),
			MaxConns: l.int32("DATABASE_POSTGRES_MAX_CONNS"),
			MinConns: l.int32("DATABASE_POSTGRES_MIN_CONNS"),
		},
		Redis: Redis{
			Addr:     l.str("REDIS_ADDRESS"),
			Password: l.optional("REDIS_PASSWORD"),
		},
		RabbitMQ: RabbitMQ{
			Host:     l.str("RABBITMQ_HOST"),
			Port:     l.str("RABBITMQ_PORT"),
			User:     l.str("RABBITMQ_USER"),
			Password: l.str("RABBITMQ_PASS"),

			PublishChannels: int(l.int32("RABBITMQ_PUBLISH_CHANNELS")),
		},
		ObjStore: ObjStore{
			Endpoint:        l.str("RUSTFS_ENDPOINT_URL"),
			AccessKeyID:     l.str("RUSTFS_ACCESS_KEY_ID"),
			SecretAccessKey: l.str("RUSTFS_SECRET_ACCESS_KEY"),
			BucketRaw:       l.str("RUSTFS_BUCKET_RAW"),
			PublicBaseURL:   l.str("CDN_BASE_URL"),
		},
		Token: Token{
			PrivateKeyPath: l.str("PRIVATE_KEY_PATH"),
			PublicKeyPath:  l.str("PUBLIC_KEY_PATH"),
			AccessTTL:      l.duration("ACCESS_TTL"),
			RefreshTTL:     l.duration("REFRESH_TTL"),
		},
	}
	cfg.Account = defaultAccount()
	cfg.Account.MailExchange = l.str("MAIL_EXCHANGE")
	cfg.Account.MailRoutingKey = l.str("MAIL_OTP_ROUTING_KEY")
	if cfg.Postgres.MinConns > cfg.Postgres.MaxConns {
		l.errs = append(l.errs, errors.New("DATABASE_POSTGRES_MIN_CONNS must be <= DATABASE_POSTGRES_MAX_CONNS"))
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

func (l *loader) optional(name string) string {
	return os.Getenv(name)
}

func (l *loader) int32(name string) int32 {
	raw := l.str(name)
	if raw == "" {
		return 0
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || n <= 0 {
		l.errs = append(l.errs, fmt.Errorf("%s must be a positive integer", name))
		return 0
	}
	return int32(n)
}

// duration parses a Go duration string such as "15m" or "720h" (no "d" unit).
func (l *loader) duration(name string) time.Duration {
	raw := l.str(name)
	if raw == "" {
		return 0
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		l.errs = append(l.errs, fmt.Errorf("%s must be a positive duration like 15m or 720h", name))
		return 0
	}
	return d
}
