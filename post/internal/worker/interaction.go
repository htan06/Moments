package worker

import (
	"context"
	"log"

	"github.com/htan06/Moments/post/internal/domain"
)

type InteractionWorker struct {
	interactionConsumer domain.InteractionConsumer
	postRepo            domain.PostRepository

	like    map[int64]int64
	comment map[int64]int64
	repost  map[int64]int64
}

func NewInteractionWorker(
	interactionConsumer domain.InteractionConsumer,
	postRepo domain.PostRepository,
) *InteractionWorker {
	return &InteractionWorker{
		interactionConsumer: interactionConsumer,
		postRepo:            postRepo,
		like:                make(map[int64]int64, 0),
		comment:             make(map[int64]int64, 0),
		repost:              make(map[int64]int64, 0),
	}
}

func (i *InteractionWorker) Run(ctx context.Context) {
	for {
		event, err := i.interactionConsumer.ReadMessage(ctx)
		if err != nil {
			log.Println(err.Error())
		}
		i.processEvent(event)
	}
}

func (i *InteractionWorker) processEvent(event domain.InteractionEvent) {
	switch event.TypeInteraction {
	case domain.LikeInteraction:
		if event.Action == domain.Created {
			i.like[event.PostID] += 1
		} else {
			i.like[event.PostID] -= 1
		}
	case domain.CommentInteraction:
		if event.Action == domain.Created {
			i.comment[event.PostID] += 1
		} else {
			i.comment[event.PostID] -= 1
		}
	case domain.RepostInteraction:
		if event.Action == domain.Created {
			i.repost[event.PostID] += 1
		} else {
			i.repost[event.PostID] -= 1
		}
	}
}
