package application

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	handler "github.com/htan06/Moments/post/internal/adapter/inbound/http"

	kafkaadpter "github.com/htan06/Moments/post/internal/adapter/outbound/kafka"
	postgresadapter "github.com/htan06/Moments/post/internal/adapter/outbound/postgres"
	redisadapter "github.com/htan06/Moments/post/internal/adapter/outbound/redis"
	rustfsadapter "github.com/htan06/Moments/post/internal/adapter/outbound/rustfs"
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
	postRepo := postgresadapter.NewPostgresPostRepository(postgresConn)
	userRepo := postgresadapter.NewPostgresUserRepository(postgresConn)
	sessionRepo := redisadapter.NewRedisSessionRepository(redisConn)
	objectStorage := rustfsadapter.NewMinIOStorage(objStorageCli)
	postProducer := kafkaadpter.NewKafkaPostProducer(cfg.KafkaWriterConfig)
	// interactionConsumer := outbound.NewKafkaInteractionConsumer(cfg.KafkaReaderConfig)

	handler := handler.NewPostHandler(
		feature.NewCreatePostUC(postRepo, userRepo, objectStorage, sessionRepo, postProducer),
		// feature.NewGetPostUC(postRepo),
		feature.NewGetPostsByCursorUC(postRepo),
		feature.NewGetFollowingFeedUC(postRepo),
		feature.NewDeletePostUC(postRepo, postProducer),
		feature.NewLikePostUC(postRepo, postProducer),
		feature.NewUnlikePostUC(postRepo, postProducer),
		// feature.NewGetRepostsUC(postRepo),
		// feature.NewCreateRepostUC(postRepo),
		// feature.NewDeleteRepostUC(postRepo),
	)

	// interacionWorker := messagequeue.NewInteractionWorker(interactionConsumer, postRepo, counterRepo)
	// go interacionWorker.Run(context.Background())

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

	// post.GET("/:postID", pm.postHandler.HandleGetPost)
	post.GET("", pm.postHandler.HandleGetPosts)

	post.POST("/session", pm.postHandler.HandleCreatePostSession)
	post.POST("/session/urls", pm.postHandler.HandleRequestUploadURLs)
	post.POST("", pm.postHandler.HandleUploadPost)
	post.DELETE("/:postID", pm.postHandler.HandleDeletePost)

	post.POST("/:postID/likes", pm.postHandler.HandleLikePost)
	post.DELETE("/:postID/likes", pm.postHandler.HandleUnlikePost)

	// repost := r.Group("/reposts", authenticated...)
	// repost.DELETE("/:repostID", pm.postHandler.HandleDeleteRepost)
	// repost.POST("", pm.postHandler.HandleCreateRepost)
	// repost.GET("", pm.postHandler.HandleGetRepostsByUsername)

	post.GET("/feed/following", pm.postHandler.HandleGetFollowingFeed)
}
