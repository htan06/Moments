// Package module is the account module's facade: it wires the module from shared
// clients, registers its routes and exposes methods to other modules. Inputs
// and outputs are built-in types or types declared here, never domain entities.
package module

import (
	"context"
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
	"github.com/htan06/Moments/application/internal/platform/httpx"
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
	// adapter/user.Profiles (ACCOUNT-R5, ACCOUNT-R7).
	Profiles interface {
		command.ProfileNameValidator
		command.ProfileCreator
		command.ProfileUsernameUpdater
	}
	Log    *slog.Logger
	Policy Policy
}

type Module struct {
	auth    *httpapi.Auth
	account *httpapi.Account
	status  *query.AccountStatusHandler
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
			ResendRegisterOTP: command.NewResendOTPHandler(domain.VerificationRegister, sessions, mail, async,
				random, otp, log),
			VerifyRegister: command.NewVerifyRegisterHandler(accounts, d.Profiles, sessions, d.Tokens, otp),
			ForgotPassword: command.NewForgotPasswordHandler(accounts, sessions, mail, async, random, otp, log),
			ResendForgotOTP: command.NewResendOTPHandler(domain.VerificationForgotPassword, sessions, mail, async,
				random, otp, log),
			ResetPassword: command.NewResetPasswordHandler(accounts, sessions, passwords, otp),
			Login:         command.NewLoginHandler(accounts, passwords, attempts, d.Tokens),
			RefreshToken:  command.NewRefreshTokenHandler(accounts, d.Tokens),
			Log:           log,
		},
		account: &httpapi.Account{
			ChangePassword: command.NewChangePasswordHandler(accounts, passwords),
			ChangeUsername: command.NewChangeUsernameHandler(accounts, d.Profiles, d.Tokens, usernames),
			RequireAuth:    httpx.Auth(d.Tokens),
			Log:            log,
		},
		status: query.NewAccountStatusHandler(accounts),
	}, nil
}

// RegisterRoutes mounts the module's HTTP routes.
func (m *Module) RegisterRoutes(r gin.IRouter) {
	m.auth.Register(r)
	m.account.Register(r)
}

// Status is what other modules may know about an account.
type Status struct {
	// Exists is false when no account has the id; the other fields are then empty.
	Exists bool
	// Active: only ACTIVE accounts may use features that need sign-in (ACCOUNT-R6).
	Active   bool
	Username string
}

// GetStatus tells whether the account exists, whether it is ACTIVE, and its
// username. A missing account is not an error: it returns Exists = false.
func (m *Module) GetStatus(ctx context.Context, userID int64) (Status, error) {
	r, err := m.status.Handle(ctx, query.AccountStatus{UserID: userID})
	if err != nil {
		return Status{}, err
	}
	return Status{Exists: r.Exists, Active: r.Active, Username: r.Username}, nil
}
