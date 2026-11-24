// Package module is the user module's facade: it wires the module from shared
// clients, registers its routes and (later) exposes methods to other modules.
package module

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	platformmq "github.com/htan06/Moments/application/internal/platform/mq"
	"github.com/htan06/Moments/application/internal/platform/token"
	"github.com/htan06/Moments/application/internal/user/adapter/httpapi"
	"github.com/htan06/Moments/application/internal/user/adapter/mq"
	"github.com/htan06/Moments/application/internal/user/adapter/password"
	"github.com/htan06/Moments/application/internal/user/adapter/postgres"
	"github.com/htan06/Moments/application/internal/user/adapter/session"
	"github.com/htan06/Moments/application/internal/user/command"
	"github.com/htan06/Moments/application/internal/user/domain"
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
	Log       *slog.Logger
	Policy    Policy
}

type Module struct {
	auth *httpapi.Auth
}

func New(d Deps) (*Module, error) {
	p := d.Policy
	otp := domain.OTPPolicy{TTL: p.OTPTTL, ResendCooldown: p.OTPResendCooldown, MaxAttempts: p.OTPMaxAttempts}
	usernames := domain.NewUsernameRules(p.ReservedUsernames)
	log := d.Log.With(slog.String("module", "user"))

	passwords, err := password.NewBcrypt(p.BcryptCost)
	if err != nil {
		return nil, err
	}
	dummyHash, err := newDummyHash(passwords)
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
			StartRegister:     command.NewStartRegisterHandler(accounts, sessions, mail, async, random, otp, log),
			ResendRegisterOTP: command.NewResendRegisterOTPHandler(sessions, mail, async, random, otp, log),
			VerifyRegisterOTP: command.NewVerifyRegisterOTPHandler(sessions, otp),
			CompleteRegister:  command.NewCompleteRegisterHandler(accounts, sessions, passwords, d.Tokens, random, usernames),
			Login:             command.NewLoginHandler(accounts, passwords, attempts, d.Tokens, dummyHash),
			RefreshToken:      command.NewRefreshTokenHandler(accounts, d.Tokens),
			Log:               log,
		},
	}, nil
}

// RegisterRoutes mounts the module's HTTP routes.
func (m *Module) RegisterRoutes(r gin.IRouter) {
	m.auth.Register(r)
}

// newDummyHash hashes a random password with the real cost, so a login for an
// unknown user costs the same as a real comparison (USER-R1).
func newDummyHash(p *password.Bcrypt) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("user: dummy hash: %w", err)
	}
	return p.Hash(hex.EncodeToString(b))
}
