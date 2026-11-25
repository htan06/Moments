package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/post/internal/adapter/inbound/http/middleware"
	"github.com/htan06/Moments/post/internal/adapter/outbound/connection"
	"github.com/htan06/Moments/post/internal/config"

	// "github.com/htan06/Moments/post/internal/adapter/outbound/config"
	"github.com/htan06/Moments/post/internal/application"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("WAR: Cannot load .env file")
	}

	config.GetStorageAddress()
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	redisConn := connection.GetRedisConn(config.RedisConfig)
	postgresConn := connection.GetPostgresConn(config.PostgresConfig)
	objStorageCli := connection.GetRustFSConn(config.RustFSConfig)

	requireAccessToken := middleware.NewJWTMiddleware(config.AuthConfig).Require()
	requireActiveUser := middleware.RequireUserStatus("ACTIVE")

	router := gin.Default()
	v1 := router.Group("/api/v1")

	applicationService := application.InitApplicationService(postgresConn, redisConn, objStorageCli, config)
	applicationService.RegisterRouter(v1, requireAccessToken, requireActiveUser)

	router.Run(config.AppConfig.Address)
}
