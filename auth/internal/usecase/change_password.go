package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/auth/internal/domain"
	"github.com/htan06/Moments/auth/internal/errs"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordCmd struct {
	Email           string
	CurrentPassword string
	NewPassword     string
}

type ChangePasswordUsecase struct {
	userRepo domain.UserRepository
}

func NewChangePasswordUsecase(
	userRepo domain.UserRepository,
) *ChangePasswordUsecase {
	return &ChangePasswordUsecase{
		userRepo: userRepo,
	}
}

func (cu *ChangePasswordUsecase) Execute(ctx context.Context, cmd ChangePasswordCmd) error {
	user, err := cu.userRepo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return errs.NewError(errs.Invalid, nil, domain.UsernameOrPasswordInvalid)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.CurrentPassword)); err != nil {
		return errs.NewError(errs.Invalid, nil, domain.UsernameOrPasswordInvalid)
	}

	if user.Status != domain.UserStatusActive {
		return errs.NewError(errs.AuthenticationFailure, nil, domain.UserNonActiveErr)
	}

	if len(cmd.NewPassword) < 6 {
		return errs.NewError(errs.Invalid, nil, domain.PasswordTooShort)
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(cmd.NewPassword), 10)
	if err != nil {
		return fmt.Errorf("ChangePasswordUsecase.Excute %w", err)
	}

	user.PasswordHash = string(newPasswordHash)
	if err := cu.userRepo.UpdatePassword(ctx, user); err != nil {
		return fmt.Errorf("ChangePasswordUsecase.Excute %w", err)
	}
	return nil
}
