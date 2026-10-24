package messagequeue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/htan06/Moments/post/internal/application/domain/event"
	"github.com/htan06/Moments/post/internal/errs"
	"github.com/htan06/Moments/post/internal/ports/inbound"
	"github.com/htan06/Moments/post/internal/ports/outbound"
)

type InteractionWorker struct {
	interactionConsumer inbound.InteractionConsumer
	postRepo            outbound.PostRepository
	counterRepo         outbound.CounterRepository
	like                map[int64]int64
	comment             map[int64]int64
	repost              map[int64]int64
}

func NewInteractionWorker(
	interactionConsumer inbound.InteractionConsumer,
	postRepo outbound.PostRepository,
	counterRepo outbound.CounterRepository,
) *InteractionWorker {
	return &InteractionWorker{
		interactionConsumer: interactionConsumer,
		postRepo:            postRepo,
		counterRepo:         counterRepo,
		like:                make(map[int64]int64),
		comment:             make(map[int64]int64),
		repost:              make(map[int64]int64),
	}
}

func (i *InteractionWorker) Run(ctx context.Context) {
	log.Println("INFO: Interaction worker is running...")

	ticker := time.NewTicker(15 * time.Second)

	eventChan, err := i.interactionConsumer.ReadMessage(ctx)
	if err != nil {
		log.Println(err.Error())
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("INFO: Iteraction worker shutdown")
			return
		case <-ticker.C:
			if len(i.like) == 0 {
				continue
			}
			if err := i.FlushLikeCount(ctx); err != nil {
				log.Println(err.Error())
			}
		case event := <-eventChan:
			i.processEvent(ctx, event)
		}
	}

}

func (i *InteractionWorker) processEvent(ctx context.Context, iEvent event.InteractionEvent) {
	var changeVal int64

	if iEvent.Action == event.Created {
		changeVal = 1
	} else {
		changeVal = -1
	}

	switch iEvent.TypeInteraction {
	case event.LikeInteraction:
		i.updateLikeCount(ctx, iEvent.PostID, changeVal)
	case event.CommentInteraction:
		i.comment[iEvent.PostID] += changeVal
	case event.RepostInteraction:
		i.repost[iEvent.PostID] += changeVal
	}

}

func (i *InteractionWorker) FlushLikeCount(ctx context.Context) error {
	if err := i.postRepo.UpadateBatchLikeCount(ctx, i.like); err != nil {
		return fmt.Errorf("InteractionWorker.FlushLikeCount: %w", err)
	}
	if err := i.interactionConsumer.Commit(ctx); err != nil {
		return fmt.Errorf("InteractionWorker.FlushLikeCount: %w", err)
	}
	i.like = make(map[int64]int64)
	return nil
}

func (i *InteractionWorker) updateLikeCount(ctx context.Context, postID int64, changeVal int64) error {
	i.like[postID] += changeVal

	switch changeVal {
	case 1:
		if err := i.counterRepo.IncPostLikes(ctx, postID); err != nil {
			if appErr, ok := errors.AsType[*errs.Error](err); !ok || appErr.Code != errs.PostLikeCacheNotFound {
				return fmt.Errorf("InteractionWorker.updateLikeCount: %w", err)
			}
			likeCountDB, err := i.postRepo.GetLikeCount(ctx, postID)

			if err != nil {
				return fmt.Errorf("InteractionWorker.updateLikeCount: %w", err)
			}

			likeCount := likeCountDB + i.like[postID]

			if err := i.counterRepo.SetPostLikesIfNotExists(ctx, postID, int(likeCount)); err != nil {
				return fmt.Errorf("InteractionWorker.updateLikeCount: %w", err)
			}
		}
	case -1:
		if err := i.counterRepo.DecPostLikes(ctx, postID); err != nil {
			if appErr, ok := errors.AsType[*errs.Error](err); !ok || appErr.Code != errs.PostLikeCacheNotFound {
				return fmt.Errorf("InteractionWorker.updateLikeCount: %w", err)
			}
			likeCountDB, err := i.postRepo.GetLikeCount(ctx, postID)

			if err != nil {
				return fmt.Errorf("InteractionWorker.updateLikeCount: %w", err)
			}

			likeCount := likeCountDB + i.like[postID]

			if err := i.counterRepo.SetPostLikesIfNotExists(ctx, postID, int(likeCount)); err != nil {
				return fmt.Errorf("InteractionWorker.updateLikeCount: %w", err)
			}
		}
	}
	return nil
}
