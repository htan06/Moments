package infra

// import (
// 	"context"
// 	"fmt"

// 	"github.com/redis/go-redis/v9"
// )

// type RedisCounterRepository struct {
// 	conn redis.Client
// }

// func NewRedisCounterRepository(conn redis.Client) *RedisCounterRepository {
// 	return &RedisCounterRepository{
// 		conn: conn,
// 	}
// }

// func (rc *RedisCounterRepository) GetFollower(ctx context.Context, username string) (*int, error) {
// 	key := fmt.Sprintf("follower-count:%s", username)
// 	cmd := rc.conn.Get(ctx, key)
// 	count, err := cmd.Int()
// 	if err != nil {
// 		return
// 	}
// }

// func (rc *RedisCounterRepository) GetFollowing(ctx context.Context, username string) (*int, error)
