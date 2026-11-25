package command

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/htan06/Moments/application/internal/account/domain"
)

// StartRegister submits everything needed to create the account. The account
// is only created once the otp sent to the email is verified (VerifyRegister).
type StartRegister struct {
	Username string
	Name     string
	Email    string
	Password string
}

type StartRegisterHandler struct {
	accounts  AccountReader
	sessions  SessionCreator
	passwords PasswordHasher
	names     ProfileNameValidator
	mailer    otpMailer
	random    Random
	usernames domain.UsernameRules
	policy    domain.OTPPolicy
	now       func() time.Time
}

func NewStartRegisterHandler(accounts AccountReader, sessions SessionCreator, passwords PasswordHasher,
	names ProfileNameValidator, sender OTPSender, async Async, random Random, usernames domain.UsernameRules, policy domain.OTPPolicy,
	log *slog.Logger) *StartRegisterHandler {
	return &StartRegisterHandler{
		accounts:  accounts,
		sessions:  sessions,
		passwords: passwords,
		names:     names,
		mailer:    otpMailer{sender: sender, async: async, log: log},
		random:    random,
		usernames: usernames,
		policy:    policy,
		now:       time.Now,
	}
}

// Handle validates the input, keeps it in a new session and sends an otp. It
// always returns a session id for valid input. If the email is already
// registered the session is a dummy: no otp is sent and every otp is wrong
// (ACCOUNT-R1). Username availability is not checked here (the client checks it
// while typing) and the username is not reserved: the unique index decides
// when the account is created.
func (h *StartRegisterHandler) Handle(ctx context.Context, cmd StartRegister) (sessionID string, err error) {
	username, err := h.usernames.Parse(cmd.Username)
	if err != nil {
		return "", err
	}
	// The profile name rule belongs to the user module, which owns Profile.
	name, err := h.names.ValidateName(cmd.Name)
	if err != nil {
		return "", err
	}
	email, err := domain.ParseEmail(cmd.Email)
	if err != nil {
		return "", err
	}
	if err := domain.ValidatePassword(cmd.Password); err != nil {
		return "", err
	}

	// Hash before looking at the email: both branches pay for bcrypt, so the
	// response time does not reveal whether the email is registered (ACCOUNT-R1).
	hash, err := h.passwords.Hash(cmd.Password)
	if err != nil {
		return "", fmt.Errorf("start register: %w", err)
	}
	exists, err := h.accounts.EmailExists(ctx, email)
	if err != nil {
		return "", fmt.Errorf("start register: %w", err)
	}
	id, err := h.random.SessionID()
	if err != nil {
		return "", err
	}
	// An otp is generated for dummy sessions too so both branches do the same work.
	otp, err := h.random.OTP()
	if err != nil {
		return "", err
	}
	reg := domain.Registration{Username: username, Name: name, PasswordHash: hash}
	s := domain.NewRegisterSession(id, email, reg, otp, exists, h.now(), h.policy)
	if err := h.sessions.Create(ctx, s); err != nil {
		return "", fmt.Errorf("start register: %w", err)
	}
	if !s.IsDummy {
		h.mailer.send(email, otp, s.ExpiresAt)
	}
	return s.ID, nil
}
