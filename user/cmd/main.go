package main

import (
	"context"
	"log"
	"net"

	"github.com/htan06/Moments/user/config"
	"github.com/htan06/Moments/user/internal/adapter/inbound/http"
	"github.com/htan06/Moments/user/internal/adapter/outbound/connection"
	"github.com/htan06/Moments/user/internal/app"
	"google.golang.org/grpc"

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

	requireAccessToken := http.NewJWTMiddleware(config.AuthConfig).Require()
	requireActiveUser := http.RequireUserActive()
	requirePendingUser := http.RequireUserPending()

	httpRouter := gin.Default()
	v1 := httpRouter.Group("/api/v1")

	grpcServer := grpc.NewServer()

	userModule := app.NewUserModule(context.Background(), postgresConn, redisConn, objectStorageConn, *config.KafkaWriterConfig, config.AppConfig.GRPCAddress)
	userModule.RegisterHTTPRouter(v1, requireAccessToken, requireActiveUser, requirePendingUser)
	userModule.RegisterGRPCRouter(grpcServer)

	followModule := app.NewFollowModule(postgresConn, *config.KafkaWriterConfig)
	followModule.RegisterHTTPRouter(v1, requireAccessToken, requireActiveUser)

	lis, err := net.Listen("tcp", config.AppConfig.GRPCAddress)
	if err != nil {
		log.Fatal(err)
	}

	go log.Fatal(grpcServer.Serve(lis))
	go log.Fatal(httpRouter.Run(config.AppConfig.AppAddress))
}
