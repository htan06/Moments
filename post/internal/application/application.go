package application

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/post/internal/delivery/handler"
	"github.com/htan06/Moments/post/internal/infra"
	"github.com/htan06/Moments/post/internal/usecase"
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

	handler := handler.NewPostHandler(
		usecase.NewCreatePostUC(postRepo, userRepo, objectStorage, cacheRepo, imgProcessor, videoProcessor, postProducer),
		usecase.NewGetPostUC(postRepo),
		usecase.NewGetUserPostsUC(postRepo),
		usecase.NewDeletePostUC(postRepo, postProducer),
		usecase.NewLikePostUC(postRepo, cacheRepo, postProducer),
		usecase.NewUnlikePostUC(postRepo, postProducer),
		usecase.NewGetRepostsUC(postRepo),
		usecase.NewCreateRepostUC(postRepo),
		usecase.NewDeleteRepostUC(postRepo),
	)

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

	me := r.Group("/users/me/posts", authenticated...)

	me.POST("/posts/session", pm.postHandler.HandlerCreatePostSession)
	me.POST("/posts/session/urls", pm.postHandler.HandlerRequestUploadURLs)
	me.POST("/posts", pm.postHandler.HandlerUploadPost)
	me.DELETE("/posts/:postID", pm.postHandler.HandlerDeletePost)
	me.DELETE("/reposts/:repostID", pm.postHandler.HandlerDeleteRepost)

	users := r.Group("/users/:username", authenticated...)
	users.GET("/posts", pm.postHandler.HandlerGetPostsByUsername)
	users.GET("/reposts", pm.postHandler.HandlerGetRepostsByUsername)

	post := r.Group("/posts", authenticated...)
	post.GET("/:postID", pm.postHandler.HandlerGetPost)
	post.POST("/:postID/likes", pm.postHandler.HandlerLikePost)
	post.DELETE("/:postID/likes", pm.postHandler.HandlerUnlikePost)

	repost := r.Group("/reposts", authenticated...)
	repost.POST("", pm.postHandler.HandlerCreateRepost)
}
