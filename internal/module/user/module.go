package user

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/infra"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

type UserModule struct {
	userHandler *UserHandler
}

func InitUserModule(
	postgresConn *pgxpool.Pool,
	redisConn *redis.Client,
	storageConn *minio.Client,
) *UserModule {
	userRepo := infra.NewPostgresUserRepository(postgresConn)
	cacheRepo := infra.NewRedisCacheRepository(redisConn)
	objectStorage := infra.NewMinIOStorage(storageConn)
	
	getProfileUsecase := usecase.NewGetProfileUsecase(userRepo)
	updateProfileUsecase := usecase.NewUpdateProfileUsecase(userRepo)
	changeAvatarUsecase := usecase.NewChangeAvatarUsecase(userRepo, cacheRepo, objectStorage)
	userHandler := NewUserHandler(getProfileUsecase, updateProfileUsecase, changeAvatarUsecase)

	return &UserModule{
		userHandler: userHandler,
	}
}

func (um *UserModule) RegisterRouter(r *gin.RouterGroup, requireAccessTokenMiddleware gin.HandlerFunc) {
	user := r.Group("/profiles")

	user.GET("/:username", requireAccessTokenMiddleware, um.userHandler.HandlerGetProfile)
	user.PATCH("/me", requireAccessTokenMiddleware, um.userHandler.HandlerUpdateProfile)
	user.GET("/upload-avatar-url", requireAccessTokenMiddleware, um.userHandler.HandlerGetUrlUploadAvatar)
	user.PATCH("/me/avatar", requireAccessTokenMiddleware, um.userHandler.HandlerCompletedUpload)
}
