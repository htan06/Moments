package main

import (
	"context"
	"log"
	"user-service/config"
	"user-service/internal/adapter/inbound/middleware"
	"user-service/internal/adapter/outbound/connection"
	"user-service/internal/application"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("WAR: Cannot load .env file: ", err)
	}
	config.GetStorageAddress()

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("WAR: Cannot load .env file", err)
	}

	redisConn := connection.GetRedisConn(config.RedisConfig)
	postgresConn := connection.GetPostgresConn(config.PostgresConfig)
	objectStorageConn := connection.GetRustFSConn(config.RustFSConfig)

	requireAccessToken := middleware.NewJWTMiddleware(config.AuthConfig).Require()
	requireActiveUser := middleware.RequireUserStatus("ACTIVE")
	requirePendingUser := middleware.RequireUserStatus("PENDING")

	router := gin.Default()
	v1 := router.Group("/api/v1")

	userModule := application.InitUserModule(context.Background(), postgresConn, redisConn, objectStorageConn, *config.KafkaWriterConfig)
	userModule.RegisterRouter(v1, requireAccessToken, requireActiveUser, requirePendingUser)

	followModule := application.InitFollowModule(postgresConn, *config.KafkaWriterConfig)
	followModule.RegisterRouter(v1, requireAccessToken, requireActiveUser)

	router.Run(config.AppConfig.Address)
}
