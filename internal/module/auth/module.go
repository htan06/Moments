package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/config"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/infra"
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
	jwtProvider *security.JWTProvier,
	mailAddress *config.MailAddress,
) *AuthModule {

	userRepo := infra.NewPostgresUserRepository(postgresConn)
	cacheRepository := infra.NewRedisCacheRepository(redisConn)
	emailOTPSender := infra.NewGmailOTPSender(dialer, mailAddress)

	otpProvider := security.NewOTPProvider()

	authService := NewAuthenticationService(userRepo, cacheRepository, emailOTPSender, jwtProvider, otpProvider)

	authHandler := NewAuthenticationHandler(authService)

	return &AuthModule{
		authHandler: authHandler,
	}
}

func (am *AuthModule) RegisterRouter(r *gin.RouterGroup, middlewares...gin.HandlerFunc) {
	auth := r.Group("/auth")

	auth.POST("/require-otp", am.authHandler.handleSendOTP)
	auth.POST("/verify-otp", am.authHandler.handleVerifyOTP)
	auth.POST("/register", am.authHandler.handleRegisterUser)
}
