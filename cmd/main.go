package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/api/middleware"
	"github.com/htan06/echo-messenger-rest-api/internal/config"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth"
	"github.com/htan06/echo-messenger-rest-api/internal/module/friend"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user"
	"github.com/htan06/echo-messenger-rest-api/internal/security"
	"github.com/joho/godotenv"
)

func main() {
	wd, _ := os.Getwd()
	fmt.Println("cwd:", wd)
	
	err := godotenv.Load()
	if err != nil {
		log.Println("WAR: Cannot load .env file")
	}

	privateKeyPath := os.Getenv("PRIVATE_KEY_PATH")
	privateData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatal("Cannot load jwt private key",  err.Error())
	}

	publicKeyPath := os.Getenv("PUBLIC_KEY_PATH")
	publicData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatal("Cannot load jwt public key", err.Error())
	}
	
	jwtConfig := config.GetJWTConfig(publicData, privateData)
	jwtProvider := security.NewJWTProvider(jwtConfig)

	redisConn := config.GetRedisConn()
	postgresConn := config.GetPostgresConn()
	gmailDialer := config.GetGmailDialer()
	mailAddress := config.GetMailAddress()

	jwtMiddleware := middleware.NewJWTMiddleware(jwtProvider)

	router := gin.Default()
	
	router.Use(middleware.CORSMiddleware())
	
	v1 := router.Group("/api/v1")


	authModule := auth.InitAuthModule(postgresConn, redisConn, gmailDialer, jwtProvider, mailAddress)
	authModule.RegisterRouter(v1, jwtMiddleware.RequireRefreshToken())

	userModule := user.InitUserModule(postgresConn)
	userModule.RegisterRouter(v1, jwtMiddleware.RequireAccessToken())

	friendModule := friend.InitFriendModule(postgresConn)
	friendModule.RegisterRouter(v1, jwtMiddleware.RequireAccessToken())
	router.Run("0.0.0.0:8080")
}