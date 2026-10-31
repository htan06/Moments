package application

import (
	"user-service/internal/adapter/inbound/handler"
	"user-service/internal/adapter/outbound"
	"user-service/internal/application/feature"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type FollowModule struct {
	followHandler *handler.FollowHandler
}

func InitFollowModule(
	postgresConn *pgxpool.Pool,
	kafkaWriterConfig kafka.WriterConfig,
) *FollowModule {
	followRepo := outbound.NewPostgresFollowRepository(postgresConn)
	followProducer := outbound.NewKafakFollowProducer(kafkaWriterConfig)

	createFollow := feature.NewCreateFollow(followRepo, followProducer)
	getFollowing := feature.NewGetFollowing(followRepo)
	getFollowers := feature.NewGetFollowers(followRepo)
	removeFollower := feature.NewRemovefollower(followRepo, followProducer)
	Unfollow := feature.NewUnfollow(followRepo, followProducer)

	followHandler := handler.NewFollowHandler(createFollow, getFollowing, getFollowers, removeFollower, Unfollow)

	return &FollowModule{
		followHandler: followHandler,
	}
}

func (fm *FollowModule) RegisterRouter(
	r *gin.RouterGroup,
	requireAccessToken gin.HandlerFunc,
	requireActiveUser gin.HandlerFunc,
) {
	users := r.Group("/users/:username/")
	users.GET("/following", requireAccessToken, requireActiveUser, fm.followHandler.HandlerGetFollowing)
	users.GET("/followers", requireAccessToken, requireActiveUser, fm.followHandler.HandlerGetFollowers)

	me := r.Group("/users/me/follows")
	me.POST("", requireAccessToken, requireActiveUser, fm.followHandler.HandlerCreateFollow)
	
	followers := r.Group("/users/me/followers")
	followers.DELETE("/:followerID", requireAccessToken, requireActiveUser, fm.followHandler.HandlerRemoveFollower)
	followings := r.Group("/users/me/followings")
	followings.DELETE("/:followingID", requireAccessToken, requireActiveUser, fm.followHandler.HandlerUnfollow)
}
