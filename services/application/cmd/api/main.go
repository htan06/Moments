// Command api runs the Moments HTTP API.
// Migrations are applied separately (moments-migrate); the app never migrates.
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/htan06/Moments/application/config"
	"github.com/htan06/Moments/application/internal/platform"
	"github.com/htan06/Moments/application/internal/platform/httpx"
	usermodule "github.com/htan06/Moments/application/internal/user/module"
)

const (
	startupTimeout  = 15 * time.Second
	shutdownTimeout = 15 * time.Second
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("api stopped", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Local dev: load .env from the working directory if present. Variables
	// already set in the environment win; production sets them directly.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	startCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	// One set of shared clients (pools) for the whole app.
	plat, err := platform.New(startCtx, cfg, log)
	if err != nil {
		return err
	}
	defer plat.Close()

	engine := httpx.NewEngine(log)
	engine.GET("/healthz", func(c *gin.Context) {
		if err := plat.Ping(c.Request.Context()); err != nil {
			httpx.Error(c, http.StatusServiceUnavailable, "UNAVAILABLE", err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	})

	user, err := usermodule.New(usermodule.Deps{
		DB:        plat.DB,
		Redis:     plat.Redis,
		Publisher: plat.Publisher,
		Tokens:    plat.Tokens,
		Log:       log,
		Policy: usermodule.Policy{
			OTPTTL:            cfg.User.OTPTTL,
			OTPResendCooldown: cfg.User.OTPResendCooldown,
			OTPMaxAttempts:    cfg.User.OTPMaxAttempts,
			LoginMaxAttempts:  cfg.User.LoginMaxAttempts,
			LoginWindow:       cfg.User.LoginWindow,
			BcryptCost:        cfg.User.BcryptCost,
			ReservedUsernames: cfg.User.ReservedUsernames,
			MailExchange:      cfg.User.MailExchange,
			MailRoutingKey:    cfg.User.MailRoutingKey,
		},
	})
	if err != nil {
		return err
	}
	user.RegisterRoutes(engine)

	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("api listening", slog.String("addr", cfg.HTTP.Addr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
	}

	log.Info("api shutting down")
	shutCtx, cancelShut := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShut()
	return srv.Shutdown(shutCtx)
}
