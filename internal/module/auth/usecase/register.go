package usecase

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/domain"
	"github.com/htan06/echo-messenger-rest-api/internal/security"
	"golang.org/x/crypto/bcrypt"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

type RegisterCmd struct {
	Email    string
	Password string
	Name     string
	Username string
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

func (ru *RegisterUsecase) Excute(ctx context.Context, cmd RegisterCmd) error {
	if !usernameRegex.MatchString(cmd.Username) {
		return errs.NewError(errs.Invalid, nil, errs.UsernameInvalid)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), 10)
	if err != nil {
		return err
	}

	otp := ru.otpProvider.RandOTP()

	userPending := domain.UserPending{
		OTP:          otp,
		Email:        cmd.Email,
		PasswordHash: string(passwordHash),
		Name:         cmd.Name,
		Username:     cmd.Username,
	}

	key := fmt.Sprintf("auth-register:%s", cmd.Email)

	if err := ru.cacheRepo.SetUserPendingIfNotExists(ctx, key, userPending, time.Minute*5); err != nil {
		return err
	}

	go ru.mailSender.SendOTP(ctx, cmd.Email, otp)
	return nil
}
