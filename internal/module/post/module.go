package post

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/internal/module/post/infra"
	"github.com/htan06/Moments/internal/module/post/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

type PostModule struct {
	postHandler *PostHandler
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
	imgProcessor := infra.NewProcessImg()

	createPostUC := usecase.NewCreatePostUC(postRepo, userRepo, objectStorage, cacheRepo, imgProcessor)
	getPostUC := usecase.NewGetPostUC(postRepo)
	getUserPostsUC := usecase.NewGetUserPostsUC(postRepo)
	deletePostUC := usecase.NewDeletePostUC(postRepo)

	handler := NewPostHandler(*createPostUC, *getPostUC, *getUserPostsUC, *deletePostUC)

	return &PostModule{
		postHandler: handler,
	}
}

func (pm *PostModule) RegisterRouter(
	r *gin.RouterGroup,
	requireAccessToken gin.HandlerFunc,
	requireActiveUser gin.HandlerFunc,
) {
	me := r.Group("/users/me/posts")

	me.POST("/session", requireAccessToken, requireActiveUser, pm.postHandler.handlerCreatePostSession)
	me.POST("", requireAccessToken, requireActiveUser, pm.postHandler.handlerUploadPost)
	me.DELETE("/:id", requireAccessToken, requireActiveUser, pm.postHandler.handlerDeletePost)

	post := r.Group("/posts")
	post.GET("/:postID", requireAccessToken, requireActiveUser, pm.postHandler.handlerGetPost)

	user := r.Group("/users/:username/posts")
	user.GET("", requireAccessToken, requireActiveUser, pm.postHandler.handlerGetPostsByUsername)
}
