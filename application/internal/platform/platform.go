// Package platform builds the shared infrastructure clients once and hands
// them to every module. It holds no business logic.
package platform

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/htan06/Moments/application/config"
	"github.com/htan06/Moments/application/internal/platform/mq"
	"github.com/htan06/Moments/application/internal/platform/objstore"
	"github.com/htan06/Moments/application/internal/platform/pg"
	"github.com/htan06/Moments/application/internal/platform/redisx"
	"github.com/htan06/Moments/application/internal/platform/token"
)

// Platform is the set of shared clients. Every field is safe for concurrent use
// and shared by all modules; modules take only what they need in module.New.
type Platform struct {
	Log       *slog.Logger
	DB        *pgxpool.Pool
	Redis     *redis.Client
	Publisher *mq.Publisher
	Store     *objstore.Store
	Tokens    *token.Manager
}

// New connects to every dependency. On error, anything already opened is closed.
func New(ctx context.Context, cfg config.Config, log *slog.Logger) (_ *Platform, err error) {
	p := &Platform{Log: log}
	defer func() {
		if err != nil {
			p.Close()
		}
	}()

	if p.DB, err = pg.Connect(ctx, pg.Config{
		URI:      cfg.Postgres.URI,
		MaxConns: cfg.Postgres.MaxConns,
		MinConns: cfg.Postgres.MinConns,
	}); err != nil {
		return nil, err
	}

	if p.Redis, err = redisx.Connect(ctx, redisx.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	}); err != nil {
		return nil, err
	}

	if p.Publisher, err = mq.NewPublisher(cfg.RabbitMQ.URL(), cfg.RabbitMQ.PublishChannels); err != nil {
		return nil, err
	}

	if p.Store, err = objstore.New(objstore.Config{
		Endpoint:        cfg.ObjStore.Endpoint,
		AccessKeyID:     cfg.ObjStore.AccessKeyID,
		SecretAccessKey: cfg.ObjStore.SecretAccessKey,
		BucketRaw:       cfg.ObjStore.BucketRaw,
		PublicBaseURL:   cfg.ObjStore.PublicBaseURL,
	}); err != nil {
		return nil, err
	}

	keys, err := token.LoadKeys(cfg.Token.PrivateKeyPath, cfg.Token.PublicKeyPath)
	if err != nil {
		return nil, err
	}
	p.Tokens = token.NewManager(keys, cfg.Token.AccessTTL, cfg.Token.RefreshTTL)

	return p, nil
}

// Ping checks the stateful dependencies (used by the health check).
func (p *Platform) Ping(ctx context.Context) error {
	if err := p.DB.Ping(ctx); err != nil {
		return errors.New("postgres unavailable")
	}
	if err := p.Redis.Ping(ctx).Err(); err != nil {
		return errors.New("redis unavailable")
	}
	return nil
}

// Close releases every client that was opened. Safe on a partially built Platform.
func (p *Platform) Close() {
	if p.Publisher != nil {
		_ = p.Publisher.Close()
	}
	if p.Redis != nil {
		_ = p.Redis.Close()
	}
	if p.DB != nil {
		p.DB.Close()
	}
}
