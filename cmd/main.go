package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/api/middleware"
	"github.com/htan06/echo-messenger-rest-api/internal/config"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth"
	"github.com/htan06/echo-messenger-rest-api/internal/module/follow"
	"github.com/htan06/echo-messenger-rest-api/internal/module/post"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user"
	"github.com/htan06/echo-messenger-rest-api/internal/security"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("WAR: Cannot load .env file")
	}

	privateKeyPath := os.Getenv("PRIVATE_KEY_PATH")
	privateData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatal("Cannot load jwt private key", err.Error())
	}

	publicKeyPath := os.Getenv("PUBLIC_KEY_PATH")
	publicData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatal("Cannot load jwt public key", err.Error())
	}

	config.GetStorageAddress()
	jwtConfig := config.GetJWTConfig(publicData, privateData)
	jwtProvider := security.NewJWTProvider(jwtConfig)

	redisConn := config.GetRedisConn()
	postgresConn := config.GetPostgresConn()
	gmailDialer := config.GetGmailDialer()
	mailAddress := config.GetMailAddress()
	objectStorageConn := config.GetMinIOConn()

	jwtMiddleware := middleware.NewJWTMiddleware(jwtProvider)

	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	v1 := router.Group("/api/v1")

	authModule := auth.InitAuthModule(postgresConn, redisConn, gmailDialer, mailAddress, jwtConfig)
	authModule.RegisterRouter(v1, jwtMiddleware.RequireAccessToken(), jwtMiddleware.RequireRefreshToken())

	userModule := user.InitUserModule(postgresConn, redisConn, objectStorageConn)
	userModule.RegisterRouter(v1, jwtMiddleware.RequireAccessToken())

	followModule := follow.InitFollowModule(postgresConn)
	followModule.RegisterRouter(v1, jwtMiddleware.RequireAccessToken())

	postModule := post.InitPostModule(postgresConn, redisConn, objectStorageConn)
	postModule.RegisterRouter(v1, jwtMiddleware.RequireAccessToken())

	router.Run("0.0.0.0:8080")
}
