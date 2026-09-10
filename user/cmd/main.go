package main

import (
	"context"
	"log"
	"os"
	"user-service/config"
	"user-service/internal/application"
	"user-service/internal/delivery/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("WAR: Cannot load .env file")
	}

	config.GetStorageAddress()
	redisConn := config.GetRedisConn()
	postgresConn := config.GetPostgresConn()
	objectStorageConn := config.GetMinIOConn()

	publicKeyPath := os.Getenv("PUBLIC_KEY_PATH")
	publicKeyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatal("Cannot load jwt public key", err.Error())
	}

	requireAccessToken := middleware.NewJWTMiddleware(publicKeyData).Require()
	requireActiveUser := middleware.RequireUserStatus("ACTIVE")
	requirePendingUser := middleware.RequireUserStatus("PENDING")

	router := gin.Default()
	v1 := router.Group("/api/v1")

	userModule := application.InitUserModule(context.Background(), postgresConn, redisConn, objectStorageConn)
	userModule.RegisterRouter(v1, requireAccessToken, requireActiveUser, requirePendingUser)

	followModule := application.InitFollowModule(postgresConn)
	followModule.RegisterRouter(v1, requireAccessToken, requireActiveUser)

	router.Run("0.0.0.0:8081")
}
