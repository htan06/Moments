package follow

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/module/follow/infra"
	"github.com/htan06/echo-messenger-rest-api/internal/module/follow/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

type followModule struct {
	followHandler *FollowHandler
}

func InitFollowModule(
	postgresConn *pgxpool.Pool,
) *followModule {
	followRepo := infra.NewPostgresFollowRepository(postgresConn)

	createFollowUsecase := usecase.NewCreateFollowUsecase(followRepo)
	getFollowingUsecase := usecase.NewGetFollowingUsecase(followRepo)
	getFollowersUsecase := usecase.NewGetFollowersUsecase(followRepo)
	removeFollowerUsecase := usecase.NewRemovefollowerUsecase(followRepo)
	UnfollowUsecase := usecase.NewUnfollowUsecase(followRepo)

	followHandler := NewFollowHandler(createFollowUsecase, getFollowingUsecase, getFollowersUsecase, removeFollowerUsecase, UnfollowUsecase)

	return &followModule{
		followHandler: followHandler,
	}
}

func (fm *followModule) RegisterRouter(r *gin.RouterGroup, requireAccessTokenMiddleware gin.HandlerFunc) {
	users := r.Group("/users/:username/")
	users.GET("/following", requireAccessTokenMiddleware, fm.followHandler.handlerGetFollowing)
	users.GET("/followers", requireAccessTokenMiddleware, fm.followHandler.handlerGetFollowers)

	me := r.Group("/users/me/follows")
	me.POST("", requireAccessTokenMiddleware, fm.followHandler.handlerCreateFollow)
	me.DELETE("/:id/remove", requireAccessTokenMiddleware, fm.followHandler.handlerRemoveFollower)
	me.DELETE("/:id/unfollow", requireAccessTokenMiddleware, fm.followHandler.handlerUnfollow)
}
