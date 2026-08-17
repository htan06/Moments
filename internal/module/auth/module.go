package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/internal/config"
	"github.com/htan06/Moments/internal/module/auth/infra"
	"github.com/htan06/Moments/internal/module/auth/usecase"
	"github.com/htan06/Moments/internal/security"
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

	getCurrentUserUC := usecase.NewGetCurrentUserUC(userRepo)
	activeUserUC := usecase.NewActiveUserUC(userRepo, jwtProvider)
	registerUsecase := usecase.NewRegisterUsecase(otpProvider, cacheRepository, emailOTPSender)
	verifyRegisterOTP := usecase.NewVerifyRegisterOTPUsecase(userRepo, cacheRepository, jwtProvider)
	loginPasswordUsecase := usecase.NewLoginPasswordUsecase(userRepo, jwtProvider)
	changePasswordUsecase := usecase.NewChangePasswordUsecase(userRepo)
	refreshTokenUsecase := usecase.NewRefreshTokenUsecase(userRepo, jwtProvider)

	authHandler := NewAuthHandler(
		getCurrentUserUC,
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

	auth.GET("/me", requireAccessToken, requireUserActive, am.authHandler.handleGetCurrentUser)
	auth.POST("/register", am.authHandler.handleRegisterUsecase)
	auth.PATCH("/active", requireAccessToken, requireUserPending, am.authHandler.handleActiveUserUC)
	auth.POST("/register/verify-otp", am.authHandler.handleVerifyRegisterOTPUsecase)
	auth.POST("/login/password", am.authHandler.handleLoginPasswordUsecase)
	auth.PATCH("/change-password", requireAccessToken, requireUserActive, am.authHandler.handleChangePasswordUsecase)
	auth.POST("/refresh-token", requireRefreshToken, requireUserActive, am.authHandler.handleRefreshTokenUsecase)
}
