package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/post/config"
	"github.com/htan06/Moments/post/internal/application"
	"github.com/htan06/Moments/post/internal/delivery/middleware"
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

	router := gin.Default()
	v1 := router.Group("/api/v1")

	postModule := application.InitPostModule(postgresConn, redisConn, objectStorageConn)
	postModule.RegisterRouter(v1, requireAccessToken, requireActiveUser)

	router.Run("0.0.0.0:8083")
}
