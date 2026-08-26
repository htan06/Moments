package follow

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/internal/module/follow/infra"
	"github.com/htan06/Moments/internal/module/follow/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FollowModule struct {
	followHandler *FollowHandler
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

	followHandler := NewFollowHandler(createFollowUsecase, getFollowingUsecase, getFollowersUsecase, removeFollowerUsecase, UnfollowUsecase)

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
	users.GET("/following", requireAccessToken, requireActiveUser, fm.followHandler.handlerGetFollowing)
	users.GET("/followers", requireAccessToken, requireActiveUser, fm.followHandler.handlerGetFollowers)

	me := r.Group("/users/me/follows")
	me.POST("", requireAccessToken, requireActiveUser, fm.followHandler.handlerCreateFollow)
	me.DELETE("/:id/remove", requireAccessToken, requireActiveUser, fm.followHandler.handlerRemoveFollower)
	me.DELETE("/:id/unfollow", requireAccessToken, requireActiveUser, fm.followHandler.handlerUnfollow)
}
