package application

import (
	"context"
	"user-service/internal/adapter/inbound/handler"
	"user-service/internal/adapter/outbound"
	"user-service/internal/application/feature"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type UserModule struct {
	userHandler *handler.UserHandler
	// updatePostCountUC *feature.UpdatePostCount
}

func InitUserModule(
	ctx context.Context,
	postgresConn *pgxpool.Pool,
	redisConn *redis.Client,
	storageConn *s3.Client,
	kafkaWriterCfg kafka.WriterConfig,
) *UserModule {
	userRepo := outbound.NewPostgresUserRepository(postgresConn)
	cacheRepo := outbound.NewRedisCacheRepository(redisConn)
	objectStorage := outbound.NewRustFSStorage(storageConn)
	// userPostConsumer := inbound.NewKafkaUserPostConsumer()
	// userFollowConsumer := inbound.NewKafkaUserFollowConsumer()
	userProducer := outbound.NewKafakUserProducer(kafkaWriterCfg)

	getProfile := feature.NewGetProfile(userRepo)
	updateProfile := feature.NewUpdateProfile(userRepo)
	changeAvatar := feature.NewChangeAvatar(userRepo, cacheRepo, objectStorage, userProducer)
	search := feature.NewSearch(userRepo)
	creaProfile := feature.NewCreateProfile(userRepo)

	userHandler := handler.NewUserHandler(creaProfile, getProfile, updateProfile, changeAvatar, search)

	// incPostCount := feature.NewUpdatePostCount(userRepo, userPostConsumer)
	// updateFollowCount := feature.NewUpdateFollowCount(userRepo, userFollowConsumer)

	// go incPostCount.Run(ctx)
	// go updateFollowCount.Run(ctx)

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
