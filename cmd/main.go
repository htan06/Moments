package main

import (
	"context"
	"log"
	"os"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/internal/api/middleware"
	"github.com/htan06/Moments/internal/config"
	"github.com/htan06/Moments/internal/module/auth"
	"github.com/htan06/Moments/internal/module/follow"
	"github.com/htan06/Moments/internal/module/post"
	"github.com/htan06/Moments/internal/module/user"
	"github.com/htan06/Moments/internal/security"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("WAR: Cannot load .env file")
	}

	if err := vips.Startup(nil); err != nil {
		log.Fatal("libvips not avaiable")
	}
	defer vips.Shutdown()

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
	requireActiveUserMW := middleware.RequireUserStatus("ACTIVE")
	requirePendingUserMW := middleware.RequireUserStatus("PENDING")
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	v1 := router.Group("/api/v1")

	authModule := auth.InitAuthModule(postgresConn, redisConn, gmailDialer, mailAddress, jwtConfig)
	authModule.RegisterRouter(v1, jwtMiddleware.RequireAccessToken(), jwtMiddleware.RequireRefreshToken(), requireActiveUserMW, requirePendingUserMW)

	userModule := user.InitUserModule(context.Background(), postgresConn, redisConn, objectStorageConn)
	userModule.RegisterRouter(v1, jwtMiddleware.RequireAccessToken(), requireActiveUserMW, requirePendingUserMW)

	followModule := follow.InitFollowModule(postgresConn)
	followModule.RegisterRouter(v1, jwtMiddleware.RequireAccessToken(), requireActiveUserMW)

	postModule := post.InitPostModule(postgresConn, redisConn, objectStorageConn)
	postModule.RegisterRouter(v1, jwtMiddleware.RequireAccessToken(), requireActiveUserMW)

	router.Run("0.0.0.0:8080")
}
