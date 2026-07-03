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

func (fm *FriendModule) RegisterRouter(r *gin.RouterGroup, requireAccessToken gin.HandlerFunc) {
	friend := r.Group("/friends")

	friend.GET("", requireAccessToken, fm.friendHandler.HandlerGetListFriends)

	friendRequests := r.Group("/friend-requests")
	friendRequests.POST("", requireAccessToken, fm.friendHandler.HandlerCreateFriendRequest)
	friendRequests.GET("/sent", requireAccessToken, fm.friendHandler.HandlerGetSentFriendRequests)
	friendRequests.GET("/received", requireAccessToken, fm.friendHandler.HandlerGetReceivedFriendRequests)
	friendRequests.POST("/:id/accept", requireAccessToken, fm.friendHandler.HandlerAcceptFriendRequest)
	friendRequests.POST("/:id/reject", requireAccessToken, fm.friendHandler.HandlerRejectFriendRequest)
	friendRequests.POST("/:id/cancel", requireAccessToken, fm.friendHandler.HandlerCancelFriendRequest)
}
