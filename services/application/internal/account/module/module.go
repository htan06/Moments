// Package module is the account module's facade: it wires the module from shared
// clients, registers its routes and (later) exposes methods to other modules.
package module

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/htan06/Moments/application/internal/account/adapter/httpapi"
	"github.com/htan06/Moments/application/internal/account/adapter/mq"
	"github.com/htan06/Moments/application/internal/account/adapter/password"
	"github.com/htan06/Moments/application/internal/account/adapter/postgres"
	"github.com/htan06/Moments/application/internal/account/adapter/session"
	"github.com/htan06/Moments/application/internal/account/command"
	"github.com/htan06/Moments/application/internal/account/domain"
	"github.com/htan06/Moments/application/internal/account/query"
	platformmq "github.com/htan06/Moments/application/internal/platform/mq"
	"github.com/htan06/Moments/application/internal/platform/token"
)

// mailTimeout bounds one background OTP mail publish.
const mailTimeout = 10 * time.Second

// Policy is the configurable business policy of the module.
type Policy struct {
	OTPTTL            time.Duration
	OTPResendCooldown time.Duration
	OTPMaxAttempts    int
	LoginMaxAttempts  int
	LoginWindow       time.Duration
	BcryptCost        int
	ReservedUsernames []string
	MailExchange      string
	MailRoutingKey    string
}

type Deps struct {
	DB        *pgxpool.Pool
	Redis     *redis.Client
	Publisher *platformmq.Publisher
	Tokens    *token.Manager
	// Profiles is the user module as seen by this module; cmd/api passes
	// adapter/user.Profiles (ACCOUNT-R7).
	Profiles interface {
		command.ProfileNameValidator
		command.ProfileCreator
	}
	Log    *slog.Logger
	Policy Policy
}

type Module struct {
	auth *httpapi.Auth
}

func New(d Deps) (*Module, error) {
	p := d.Policy
	otp := domain.OTPPolicy{TTL: p.OTPTTL, ResendCooldown: p.OTPResendCooldown, MaxAttempts: p.OTPMaxAttempts}
	usernames := domain.NewUsernameRules(p.ReservedUsernames)
	log := d.Log.With(slog.String("module", "account"))

	passwords, err := password.NewBcrypt(p.BcryptCost)
	if err != nil {
		return nil, err
	}

	accounts := postgres.NewAccounts(d.DB)
	sessions := session.NewVerificationSessions(d.Redis)
	attempts := session.NewLoginAttempts(d.Redis, p.LoginMaxAttempts, p.LoginWindow)
	mail := mq.NewMailPublisher(d.Publisher, mq.MailRoute{Exchange: p.MailExchange, RoutingKey: p.MailRoutingKey})
	async := command.GoAsync(mailTimeout)
	random := command.CryptoRandom{}

	return &Module{
		auth: &httpapi.Auth{
			UsernameAvailability: query.NewUsernameAvailabilityHandler(accounts, usernames),
			StartRegister: command.NewStartRegisterHandler(accounts, sessions, passwords, d.Profiles, mail, async,
				random, usernames, otp, log),
			ResendRegisterOTP: command.NewResendRegisterOTPHandler(sessions, mail, async, random, otp, log),
			VerifyRegister:    command.NewVerifyRegisterHandler(accounts, d.Profiles, sessions, d.Tokens, otp),
			Login:             command.NewLoginHandler(accounts, passwords, attempts, d.Tokens),
			RefreshToken:      command.NewRefreshTokenHandler(accounts, d.Tokens),
			Log:               log,
		},
	}, nil
}

// RegisterRoutes mounts the module's HTTP routes.
func (m *Module) RegisterRoutes(r gin.IRouter) {
	m.auth.Register(r)
}
