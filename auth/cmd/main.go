package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/auth/config"
	"github.com/htan06/Moments/auth/internal/application"
	"github.com/htan06/Moments/auth/internal/delivery/middleware"
	"github.com/htan06/Moments/auth/internal/security"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("WAR: Cannot load .env file")
	}

	privateKeyPath := os.Getenv("PRIVATE_KEY_PATH")
	privateKeyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatal("Cannot load jwt private key", err.Error())
	}

	publicKeyPath := os.Getenv("PUBLIC_KEY_PATH")
	publicKeyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatal("Cannot load jwt public key", err.Error())
	}

	jwtConfig := config.GetJWTConfig(publicKeyData, privateKeyData)
	jwtProvider := security.NewJWTProvider(jwtConfig)

	redisConn := config.GetRedisConn()
	postgresConn := config.GetPostgresConn()
	gmailDialer := config.GetGmailDialer()
	mailAddress := config.GetMailAddress()

	jwtMiddleware := middleware.NewJWTMiddleware(jwtProvider)
	requireActiveUser := middleware.RequireUserStatus("ACTIVE")
	requirePendingUser := middleware.RequireUserStatus("PENDING")

	router := gin.Default()
	v1 := router.Group("/api/v1")

	authModule := application.InitAuthModule(postgresConn, redisConn, gmailDialer, mailAddress, jwtConfig)
	authModule.RegisterRouter(v1, jwtMiddleware.RequireAccessToken(), jwtMiddleware.RequireRefreshToken(), requireActiveUser, requirePendingUser)

	router.Run("0.0.0.0:8082")
}
