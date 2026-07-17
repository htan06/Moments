package post

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/module/post/infra"
	"github.com/htan06/echo-messenger-rest-api/internal/module/post/usecase"
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
	userRepo := infra.NewPostgresUserRepository(postgresConn)
	cacheRepo := infra.NewRedisCacheRepository(redisConn)
	objectStorage := infra.NewMinIOStorage(storageConn)

	createPostUC := usecase.NewCreatePostUC(userRepo, objectStorage, cacheRepo)

	handler := NewPostHandler(*createPostUC)

	return &PostModule{
		postHandler: handler,
	}
}

func (pm *PostModule) RegisterRouter(r *gin.RouterGroup, requireAccessTokenMiddleware gin.HandlerFunc) {
	me := r.Group("/users/me/posts")

	me.POST("", requireAccessTokenMiddleware, pm.postHandler.handlerCreatePost)
}
