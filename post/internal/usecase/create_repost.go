package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/post/internal/domain"
)

type CreateRepostCmd struct {
	UserID int64
	PostID int64
}

type CreateRepostUC struct {
	postRepo domain.PostRepository
}

func NewCreateRepostUC(postRepo domain.PostRepository) *CreateRepostUC {
	return &CreateRepostUC{
		postRepo: postRepo,
	}
}

func (uc *CreateRepostUC) Execute(ctx context.Context, cmd CreateRepostCmd) (int64, error) {
	id, err := uc.postRepo.CreateRepost(ctx, cmd.UserID, cmd.PostID)
	if err != nil {
		return 0, fmt.Errorf("CreateRepostUC.Excecute: %w", err)
	}

	return id, nil
}
