package user

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/infra"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserModule struct {
	userHandler *UserHandler
}

func InitUserModule(
	postgresConn *pgxpool.Pool,
) *UserModule {
	userRepo := infra.NewPostgresUserRepository(postgresConn)

	userService := NewUserService(userRepo)

	userHandler := NewUserHandler(userService)

	return &UserModule{
		userHandler: userHandler,
	}
}

func (um *UserModule) RegisterRouter(r *gin.RouterGroup, requireAccessTokenMiddleware gin.HandlerFunc) {
	user := r.Group("/users")

	user.GET("/me/profile", requireAccessTokenMiddleware, um.userHandler.HandleGetCurrentUserProfile)
	user.GET("/:username/profile", requireAccessTokenMiddleware, um.userHandler.HandlerFindUserByUsername)
	user.PATCH("/me/profile", requireAccessTokenMiddleware, um.userHandler.HandleUpdateProfile)
	user.PATCH("/me/setting/read-status", requireAccessTokenMiddleware, um.userHandler.HandleChangeReadStatus)
	user.PATCH("/me/username", requireAccessTokenMiddleware, um.userHandler.HandleUpdateUsername)
}
