package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/htan06/Moments/post/internal/domain"
)

type InteractionWorker struct {
	interactionConsumer domain.InteractionConsumer
	postRepo            domain.PostRepository
	cacheRepo           domain.CacheRepository
	like                map[int64]int64
	comment             map[int64]int64
	repost              map[int64]int64
}

func NewInteractionWorker(
	interactionConsumer domain.InteractionConsumer,
	postRepo domain.PostRepository,
	cacheRepo domain.CacheRepository,
) *InteractionWorker {
	return &InteractionWorker{
		interactionConsumer: interactionConsumer,
		postRepo:            postRepo,
		cacheRepo:           cacheRepo,
		like:                make(map[int64]int64),
		comment:             make(map[int64]int64),
		repost:              make(map[int64]int64),
	}
}

func (i *InteractionWorker) Run(ctx context.Context) {
	log.Println("INFO: Interaction worker is running...")

	ticker := time.NewTicker(5 * time.Second)

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
			i.processEvent(event)
		}
	}

}

func (i *InteractionWorker) processEvent(event domain.InteractionEvent) {
	var changeVal int64

	if event.Action == domain.Created {
		changeVal = 1
	} else {
		changeVal = -1
	}

	switch event.TypeInteraction {
	case domain.LikeInteraction:
		i.like[event.PostID] += changeVal
	case domain.CommentInteraction:
		i.comment[event.PostID] += changeVal
	case domain.RepostInteraction:
		i.repost[event.PostID] += changeVal
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
