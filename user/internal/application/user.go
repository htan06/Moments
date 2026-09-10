package application

import (
	"context"
	"user-service/internal/delivery/handler"
	"user-service/internal/infra"
	"user-service/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

type UserModule struct {
	userHandler       *handler.UserHandler
	updatePostCountUC *usecase.UpdatePostCountUC
}

func InitUserModule(
	ctx context.Context,
	postgresConn *pgxpool.Pool,
	redisConn *redis.Client,
	storageConn *minio.Client,
) *UserModule {
	userRepo := infra.NewPostgresUserRepository(postgresConn)
	cacheRepo := infra.NewRedisCacheRepository(redisConn)
	objectStorage := infra.NewMinIOStorage(storageConn)
	processImg := infra.NewProcessImg()
	userPostConsumer := infra.NewKafkaUserPostConsumer()
	userFollowConsumer := infra.NewKafkaUserFollowConsumer()

	getProfileUsecase := usecase.NewGetProfileUsecase(userRepo)
	updateProfileUsecase := usecase.NewUpdateProfileUsecase(userRepo)
	changeAvatarUsecase := usecase.NewChangeAvatarUsecase(userRepo, cacheRepo, objectStorage, processImg)
	searchUC := usecase.NewSearchUC(userRepo)
	creaProfileUC := usecase.NewCreateProfileUC(userRepo)

	userHandler := handler.NewUserHandler(creaProfileUC, getProfileUsecase, updateProfileUsecase, changeAvatarUsecase, searchUC)

	incPostCountUC := usecase.NewUpdatePostCountUC(userRepo, userPostConsumer)
	updateFollowCount := usecase.NewUpdateFollowCountUC(userRepo, userFollowConsumer)

	go incPostCountUC.Run(ctx)
	go updateFollowCount.Run(ctx)

	return &UserModule{
		userHandler: userHandler,
	}
}

func (um *UserModule) RegisterRouter(
	r *gin.RouterGroup,
	requireAccessToken gin.HandlerFunc,
	requireActiveUser gin.HandlerFunc,
	requirePendingUser gin.HandlerFunc,
) {
	user := r.Group("/profiles")

	user.GET("/:username", requireAccessToken, requireActiveUser, um.userHandler.HandlerGetProfile)
	user.GET("", requireAccessToken, requireActiveUser, um.userHandler.HandleFindProfiles)
	user.POST("", requireAccessToken, requirePendingUser, um.userHandler.HandlerCreateProfile)
	user.PATCH("/me", requireAccessToken, requireActiveUser, um.userHandler.HandlerUpdateProfile)
	user.GET("/upload-avatar-url", requireAccessToken, requireActiveUser, um.userHandler.HandlerGetUrlUploadAvatar)
	user.PATCH("/me/avatar", requireAccessToken, requireActiveUser, um.userHandler.HandlerCompletedUpload)
}
