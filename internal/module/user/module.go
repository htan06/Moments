package user

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/infra"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserModule struct {
	userHandler *UserHandler
}

func InitUserModule(
	postgresConn *pgxpool.Pool,
) *UserModule {
	userRepo := infra.NewPostgresUserRepository(postgresConn)

	// userService := NewUserService(userRepo)
	getProfileUsecase := usecase.NewGetProfileUsecase(userRepo)
	updateProfileUsecase := usecase.NewUpdateProfileUsecase(userRepo)
	userHandler := NewUserHandler(getProfileUsecase, updateProfileUsecase)

	return &UserModule{
		userHandler: userHandler,
	}
}

func (um *UserModule) RegisterRouter(r *gin.RouterGroup, requireAccessTokenMiddleware gin.HandlerFunc) {
	user := r.Group("/profiles")

	user.GET("/:username", requireAccessTokenMiddleware, um.userHandler.HandlerGetProfile)
	user.PATCH("/me", requireAccessTokenMiddleware, um.userHandler.HandlerUpdateProfile)
	// user.PATCH("/me/setting/read-status", requireAccessTokenMiddleware, um.userHandler.HandleChangeReadStatus)
	// user.PATCH("/me/username", requireAccessTokenMiddleware, um.userHandler.HandleUpdateUsername)
}
