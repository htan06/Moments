package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/config"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/infra"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/usecase"
	"github.com/htan06/echo-messenger-rest-api/internal/security"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"gopkg.in/gomail.v2"
)

type AuthModule struct {
	authHandler *AuthHandler
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

	registerUsecase := usecase.NewRegisterUsecase(otpProvider, cacheRepository, emailOTPSender)
	verifyRegisterOTP := usecase.NewVerifyRegisterOTPUsecase(userRepo, cacheRepository, jwtProvider)
	loginPasswordUsecase := usecase.NewLoginPasswordUsecase(userRepo, jwtProvider)
	changePasswordUsecase := usecase.NewChangePasswordUsecase(userRepo)
	refreshTokenUsecase := usecase.NewRefreshTokenUsecase(userRepo, jwtProvider)

	authHandler := NewAuthHandler(
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

func (am *AuthModule) RegisterRouter(r *gin.RouterGroup, requireAccessToken gin.HandlerFunc, requireRefreshToken gin.HandlerFunc) {
	auth := r.Group("/auth")

	auth.POST("/register", am.authHandler.handleRegisterUsecase)
	auth.POST("/register/verify-otp", am.authHandler.handleVerifyRegisterOTPUsecase)
	auth.POST("/login/password", am.authHandler.handleLoginPasswordUsecase)
	auth.PATCH("/change-password", requireAccessToken, am.authHandler.handleChangePasswordUsecase)
	auth.POST("/refresh-token", requireRefreshToken, am.authHandler.handleRefreshTokenUsecase)
}
