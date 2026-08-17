package usecase

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/htan06/Moments/internal/config"
	"github.com/htan06/Moments/internal/errs"
	"github.com/htan06/Moments/internal/module/auth/domain"
	"github.com/htan06/Moments/internal/security"
	"golang.org/x/crypto/bcrypt"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

type RegisterCmd struct {
	Email    string
	Password string
	// Name     string
	// Username string
}

type RegisterUsecase struct {
	userRepo    domain.UserRepository
	otpProvider *security.OTPProvider
	cacheRepo   domain.CacheRepository
	mailSender  domain.EmailOTPSender
}

func NewRegisterUsecase(
	otpProvider *security.OTPProvider,
	cacheRepo domain.CacheRepository,
	mailSender domain.EmailOTPSender,
) *RegisterUsecase {
	return &RegisterUsecase{
		otpProvider: otpProvider,
		cacheRepo:   cacheRepo,
		mailSender:  mailSender,
	}
}

func (ru *RegisterUsecase) Execute(ctx context.Context, cmd RegisterCmd) error {
	if len(cmd.Password) < 6 {
		return errs.NewError(errs.Invalid, nil, domain.PasswordTooShort)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), 10)
	if err != nil {
		return fmt.Errorf("RegisterUsecase.Excute %w", err)
	}

	userPending := domain.UserPending{
		Email:        cmd.Email,
		PasswordHash: string(passwordHash),
		OTP:          ru.otpProvider.RandOTP(),
	}

	key := fmt.Sprintf("%s:%s", config.UserRegisterPrefix, cmd.Email)
	if err := ru.cacheRepo.SetUserPendingIfNotExists(ctx, key, userPending, time.Minute*5); err != nil {
		return fmt.Errorf("RegisterUsecase.Excute %w", err)
	}

	go ru.mailSender.SendOTP(ctx, cmd.Email, userPending.OTP)
	return nil
}
