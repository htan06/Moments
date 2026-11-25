package feature

import (
	"context"
	"fmt"
	"log"

	"github.com/htan06/Moments/post/internal/application/ports"
	"github.com/htan06/Moments/post/internal/domain/event"
)

type DeletePostUC struct {
	postDeleter  ports.PostDeleter
	postProducer ports.PostEventPublisher
}

func NewDeletePostUC(
	postDeleter ports.PostDeleter,
	postProducer ports.PostEventPublisher,
) *DeletePostUC {
	return &DeletePostUC{
		postDeleter:  postDeleter,
		postProducer: postProducer,
	}
}

type DeletePostCmd struct {
	UserID int64
	PostID int64
}

func (dp *DeletePostUC) Execute(ctx context.Context, cmd DeletePostCmd) error {
	if err := dp.postDeleter.DeletePostByUserIDAndPostID(ctx, cmd.UserID, cmd.PostID); err != nil {
		return fmt.Errorf("DeletePostUC.Execute: %w", err)
	}

	go func() {
		err := dp.postProducer.SendPostDeleted(context.Background(), event.NewPostDeleted(cmd.PostID, cmd.UserID))
		if err != nil {
			log.Println(err.Error())
		}
	}()

	return nil
}
