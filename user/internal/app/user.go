package app

import (
	"context"

	"github.com/htan06/Moments/grpc/user"
	grpcin "github.com/htan06/Moments/user/internal/adapter/inbound/grpc"
	"github.com/htan06/Moments/user/internal/adapter/inbound/http"
	"github.com/htan06/Moments/user/internal/adapter/outbound"
	"github.com/htan06/Moments/user/internal/application/feature"
	"google.golang.org/grpc"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type UserModule struct {
	httpHandler *http.UserHandler
	grpcHandler *grpcin.ProfileHandler
}

func NewUserModule(
	ctx context.Context,
	postgresConn *pgxpool.Pool,
	redisConn *redis.Client,
	storageConn *s3.Client,
	kafkaWriterCfg kafka.WriterConfig,
	grpcAddress string,
) *UserModule {
	userRepo := outbound.NewPostgresUserRepository(postgresConn)
	cacheRepo := outbound.NewRedisCacheRepository(redisConn)
	objectStorage := outbound.NewRustFSStorage(storageConn)
	userProducer := outbound.NewKafakUserProducer(kafkaWriterCfg)

	getProfile := feature.NewGetProfile(userRepo)
	updateProfile := feature.NewUpdateProfile(userRepo)
	changeAvatar := feature.NewChangeAvatar(userRepo, userRepo, cacheRepo, objectStorage, userProducer)
	search := feature.NewSearch(userRepo)
	creaProfile := feature.NewCreateProfile(userRepo)
	getProfilesSummary := feature.NewGetProfilesSummary(userRepo)

	httpHandler := http.NewUserHandler(creaProfile, getProfile, updateProfile, changeAvatar, search)
	grpcHandler := grpcin.NewProfileHandler(getProfilesSummary)

	return &UserModule{
		httpHandler: httpHandler,
		grpcHandler: grpcHandler,
	}
}

func (um *UserModule) RegisterHTTPRouter(
	r *gin.RouterGroup,
	requireAccessToken gin.HandlerFunc,
	requireActiveUser gin.HandlerFunc,
	requirePendingUser gin.HandlerFunc,
) {
	user := r.Group("/profiles")

	user.GET("/:username", requireAccessToken, requireActiveUser, um.httpHandler.HandlerGetProfile)
	user.GET("", requireAccessToken, requireActiveUser, um.httpHandler.HandleFindProfiles)
	user.POST("", requireAccessToken, requirePendingUser, um.httpHandler.HandlerCreateProfile)
	user.PATCH("/me", requireAccessToken, requireActiveUser, um.httpHandler.HandlerUpdateProfile)
	user.GET("/upload-avatar-url", requireAccessToken, requireActiveUser, um.httpHandler.HandlerGetUrlUploadAvatar)
	user.PATCH("/me/avatar", requireAccessToken, requireActiveUser, um.httpHandler.HandlerCompletedUpload)
}

func (um *UserModule) RegisterGRPCRouter(s *grpc.Server) {
	user.RegisterUserServiceServer(s, um.grpcHandler)
}