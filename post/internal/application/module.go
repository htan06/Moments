package application

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/post/internal/delivery/handler"
	"github.com/htan06/Moments/post/internal/infra"
	"github.com/htan06/Moments/post/internal/usecase"
	"github.com/htan06/Moments/post/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

type PostModule struct {
	postHandler *handler.PostHandler
}

func InitPostModule(
	postgresConn *pgxpool.Pool,
	redisConn *redis.Client,
	storageConn *minio.Client,
) *PostModule {
	postRepo := infra.NewPostgresPostRepository(postgresConn)
	userRepo := infra.NewPostgresUserRepository(postgresConn)
	cacheRepo := infra.NewRedisCacheRepository(redisConn)
	objectStorage := infra.NewMinIOStorage(storageConn)
	imgProcessor := infra.NewGoVipsProcessImg()
	videoProcessor := infra.NewFfmpegProcessVideo()
	postProducer := infra.NewKafkaPostProducer()
	interactionConsumer := infra.NewKafkaInteractionConsumer()

	handler := handler.NewPostHandler(
		usecase.NewCreatePostUC(postRepo, userRepo, objectStorage, cacheRepo, imgProcessor, videoProcessor, postProducer),
		usecase.NewGetPostUC(postRepo),
		usecase.NewGetPostsByCursorUC(postRepo, cacheRepo),
		usecase.NewDeletePostUC(postRepo, postProducer),
		usecase.NewLikePostUC(postRepo, cacheRepo, postProducer),
		usecase.NewUnlikePostUC(postRepo, postProducer),
		usecase.NewGetRepostsUC(postRepo),
		usecase.NewCreateRepostUC(postRepo),
		usecase.NewDeleteRepostUC(postRepo),
	)

	interacionWorker := worker.NewInteractionWorker(interactionConsumer, postRepo, cacheRepo)
	go interacionWorker.Run(context.Background())

	return &PostModule{
		postHandler: handler,
	}
}

func (pm *PostModule) RegisterRouter(
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
}
