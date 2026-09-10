package application

import (
	"user-service/internal/delivery/handler"
	"user-service/internal/infra"
	"user-service/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FollowModule struct {
	followHandler *handler.FollowHandler
}

func InitFollowModule(
	postgresConn *pgxpool.Pool,
) *FollowModule {
	followRepo := infra.NewPostgresFollowRepository(postgresConn)
	followProducer := infra.NewKafakFollowProducer()

	createFollowUsecase := usecase.NewCreateFollowUsecase(followRepo, followProducer)
	getFollowingUsecase := usecase.NewGetFollowingUsecase(followRepo)
	getFollowersUsecase := usecase.NewGetFollowersUsecase(followRepo)
	removeFollowerUsecase := usecase.NewRemovefollowerUsecase(followRepo, followProducer)
	UnfollowUsecase := usecase.NewUnfollowUsecase(followRepo, followProducer)

	followHandler := handler.NewFollowHandler(createFollowUsecase, getFollowingUsecase, getFollowersUsecase, removeFollowerUsecase, UnfollowUsecase)

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
	me.DELETE("/:id/remove", requireAccessToken, requireActiveUser, fm.followHandler.HandlerRemoveFollower)
	me.DELETE("/:id/unfollow", requireAccessToken, requireActiveUser, fm.followHandler.HandlerUnfollow)
}
