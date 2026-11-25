package application

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	handler "github.com/htan06/Moments/post/internal/adapter/inbound/http"
	messagequeue "github.com/htan06/Moments/post/internal/adapter/inbound/mq"
	"github.com/htan06/Moments/post/internal/adapter/outbound"
	"github.com/htan06/Moments/post/internal/application/feature"
	"github.com/htan06/Moments/post/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ApplicationService struct {
	postHandler *handler.PostHandler
}

func InitApplicationService(
	postgresConn *pgxpool.Pool,
	redisConn *redis.Client,
	objStorageCli *s3.Client,
	cfg *config.Config,
) *ApplicationService {
	postRepo := outbound.NewPostgresPostRepository(postgresConn)
	userRepo := outbound.NewPostgresUserRepository(postgresConn)
	sessionRepo := outbound.NewRedisSessionRepository(redisConn)
	counterRepo := outbound.NewRedisCounterRepository(redisConn)
	objectStorage := outbound.NewMinIOStorage(objStorageCli)
	postProducer := outbound.NewKafkaPostProducer(cfg.KafkaWriterConfig)
	interactionConsumer := outbound.NewKafkaInteractionConsumer(cfg.KafkaReaderConfig)

	handler := handler.NewPostHandler(
		feature.NewCreatePostUC(postRepo, userRepo, objectStorage, sessionRepo, postProducer),
		feature.NewGetPostUC(postRepo),
		feature.NewGetPostsByCursorUC(postRepo, counterRepo),
		feature.NewGetFollowingFeedUC(postRepo, counterRepo),
		feature.NewDeletePostUC(postRepo, postProducer),
		feature.NewLikePostUC(postRepo, postProducer),
		feature.NewUnlikePostUC(postRepo, postProducer),
		feature.NewGetRepostsUC(postRepo),
		feature.NewCreateRepostUC(postRepo),
		feature.NewDeleteRepostUC(postRepo),
	)

	interacionWorker := messagequeue.NewInteractionWorker(interactionConsumer, postRepo, counterRepo)
	go interacionWorker.Run(context.Background())

	return &ApplicationService{
		postHandler: handler,
	}
}

func (pm *ApplicationService) RegisterRouter(
	r *gin.RouterGroup,
	requireAccessToken gin.HandlerFunc,
	requireActiveUser gin.HandlerFunc,
) {
	authenticated := []gin.HandlerFunc{requireAccessToken, requireActiveUser}

	post := r.Group("/posts", authenticated...)

	post.GET("/:postID", pm.postHandler.HandleGetPost)
	post.GET("", pm.postHandler.HandleGetPosts)

	post.POST("/session", pm.postHandler.HandleCreatePostSession)
	post.POST("/session/urls", pm.postHandler.HandleRequestUploadURLs)
	post.POST("", pm.postHandler.HandleUploadPost)
	post.DELETE("/:postID", pm.postHandler.HandleDeletePost)

	post.POST("/:postID/likes", pm.postHandler.HandleLikePost)
	post.DELETE("/:postID/likes", pm.postHandler.HandleUnlikePost)

	repost := r.Group("/reposts", authenticated...)
	repost.DELETE("/:repostID", pm.postHandler.HandleDeleteRepost)
	repost.POST("", pm.postHandler.HandleCreateRepost)
	repost.GET("", pm.postHandler.HandleGetRepostsByUsername)

	post.GET("/feed/following", pm.postHandler.HandleGetFollowingFeed)
}
