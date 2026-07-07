package usecase

import (
	"context"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/domain"
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

func (cu *ChangePasswordUsecase) Excute(ctx context.Context, cmd ChangePasswordCmd) error {
	user, err := cu.userRepo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return errs.NewError(errs.Invalid, nil, errs.InvalidUsernameOrPassword)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.CurrentPassword)); err != nil {
		return errs.NewError(errs.Invalid, nil, errs.InvalidUsernameOrPassword)
	}

	if user.Status != domain.UserActive {
		return errs.NewError(errs.AuthenticationFailure, nil, errs.UserNonActive)
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(cmd.NewPassword), 10)
	if err != nil {
		return err
	}

	user.PasswordHash = string(newPasswordHash)
	if err := cu.userRepo.UpdatePassword(ctx, user); err != nil {
		return err
	}
	return nil
}
