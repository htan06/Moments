package friend

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/module/friend/infra"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FriendModule struct {
	friendHandler *FriendHandler
}

func InitFriendModule(
	postgresConn *pgxpool.Pool,
) *FriendModule {
	friendRepo := infra.NewPostgresUserRepository(postgresConn)

	friendService := NewFriendService(friendRepo)

	friendHandler := NewFriendHandler(friendService)

	return &FriendModule{
		friendHandler: friendHandler,
	}
}

func (fm *FriendModule) RegisterRouter(r *gin.RouterGroup, middleware gin.HandlerFunc) {
	profile := r.Group("/profile")
	friend := r.Group("/friend")

	profile.GET("/:username", fm.friendHandler.HandlerFindUserByUsername)
	friend.POST("/friend-request", middleware, fm.friendHandler.HandlerFriendRequest)
}