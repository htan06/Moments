package application

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/auth/config"
	"github.com/htan06/Moments/auth/internal/delivery/handler"
	"github.com/htan06/Moments/auth/internal/infra"
	"github.com/htan06/Moments/auth/internal/security"
	"github.com/htan06/Moments/auth/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"gopkg.in/gomail.v2"
)

type AuthModule struct {
	authHandler *handler.AuthHandler
}

func InitAuthModule(
	postgresConn *pgxpool.Pool,
	redisConn *redis.Client,
	dialer *gomail.Dialer,
	mailAddress *config.MailAddress,
	jwtConfig *config.JWTConfig,
) *AuthModule {

	userRepo := infra.NewPostgresUserRepository(postgresConn)
	cacheRepository := infra.NewRedisCacheRepository(redisConn)
	emailOTPSender := infra.NewGmailOTPSender(dialer, mailAddress)
	otpProvider := security.NewOTPProvider()
	jwtProvider := security.NewJWTProvider(jwtConfig)

	activeUserUC := usecase.NewActiveUserUC(userRepo, jwtProvider)
	registerUsecase := usecase.NewRegisterUsecase(otpProvider, cacheRepository, emailOTPSender)
	verifyRegisterOTP := usecase.NewVerifyRegisterOTPUsecase(userRepo, cacheRepository, jwtProvider)
	loginPasswordUsecase := usecase.NewLoginPasswordUsecase(userRepo, jwtProvider)
	changePasswordUsecase := usecase.NewChangePasswordUsecase(userRepo)
	refreshTokenUsecase := usecase.NewRefreshTokenUsecase(userRepo, jwtProvider)

	authHandler := handler.NewAuthHandler(
		activeUserUC,
		registerUsecase,
		verifyRegisterOTP,
		loginPasswordUsecase,
		changePasswordUsecase,
		refreshTokenUsecase,
		jwtConfig)

	return &AuthModule{
		authHandler: authHandler,
	}
}

func (am *AuthModule) RegisterRouter(
	r *gin.RouterGroup,
	requireAccessToken gin.HandlerFunc,
	requireRefreshToken gin.HandlerFunc,
	requireUserActive gin.HandlerFunc,
	requireUserPending gin.HandlerFunc,
) {
	auth := r.Group("/auth")

	// auth.GET("/me", requireAccessToken, requireUserActive, am.authHandler.handleGetCurrentUser)
	auth.POST("/register", am.authHandler.HandleRegisterUsecase)
	auth.PATCH("/active", requireAccessToken, requireUserPending, am.authHandler.HandleActiveUserUC)
	auth.POST("/register/verify-otp", am.authHandler.HandleVerifyRegisterOTPUsecase)
	auth.POST("/login/password", am.authHandler.HandleLoginPasswordUsecase)
	auth.PATCH("/change-password", requireAccessToken, requireUserActive, am.authHandler.HandleChangePasswordUsecase)
	auth.POST("/refresh-token", requireRefreshToken, requireUserActive, am.authHandler.HandleRefreshTokenUsecase)
}
