package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/joshleeeeee/go-lite-auth/internal/config"
	"github.com/redis/go-redis/v9"
)

// TicketData stores ticket-related user and service information
type TicketData struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Service  string `json:"service"`
}

var RDB *redis.Client

// InitRedis initializes the Redis connection
func InitRedis(cfg *config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

// Redis key prefixes for different purposes
const (
	PrefixSession    = "session:"
	PrefixBlacklist  = "blacklist:"
	PrefixTicket     = "ticket:"
	PrefixLoginFail  = "login_fail:"
	PrefixRefreshToken = "refresh_token:"
)

// Session operations

func SetSession(ctx context.Context, sessionID string, userID uint, expire time.Duration) error {
	return RDB.Set(ctx, PrefixSession+sessionID, userID, expire).Err()
}

func GetSession(ctx context.Context, sessionID string) (string, error) {
	return RDB.Get(ctx, PrefixSession+sessionID).Result()
}

func DeleteSession(ctx context.Context, sessionID string) error {
	return RDB.Del(ctx, PrefixSession+sessionID).Err()
}

// Token blacklist operations

func AddToBlacklist(ctx context.Context, tokenID string, expire time.Duration) error {
	return RDB.Set(ctx, PrefixBlacklist+tokenID, "1", expire).Err()
}

func IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	result, err := RDB.Exists(ctx, PrefixBlacklist+tokenID).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// SSO Ticket operations (one-time use with service validation)

// SetTicketWithService stores a ticket with associated user and service data
func SetTicketWithService(ctx context.Context, ticketID string, data *TicketData, expire time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal ticket data: %w", err)
	}
	return RDB.Set(ctx, PrefixTicket+ticketID, jsonData, expire).Err()
}

// GetAndDeleteTicketData atomically retrieves and deletes ticket data (one-time use)
func GetAndDeleteTicketData(ctx context.Context, ticketID string) (*TicketData, error) {
	// Use GetDel for atomic get and delete
	result, err := RDB.GetDel(ctx, PrefixTicket+ticketID).Result()
	if err != nil {
		return nil, err
	}

	var data TicketData
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ticket data: %w", err)
	}

	return &data, nil
}

// Login rate limiting

func IncrLoginFail(ctx context.Context, key string, expire time.Duration) (int64, error) {
	count, err := RDB.Incr(ctx, PrefixLoginFail+key).Result()
	if err != nil {
		return 0, err
	}
	// Set expiry only on first failure
	if count == 1 {
		RDB.Expire(ctx, PrefixLoginFail+key, expire)
	}
	return count, nil
}

func GetLoginFailCount(ctx context.Context, key string) (int64, error) {
	count, err := RDB.Get(ctx, PrefixLoginFail+key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

func ClearLoginFail(ctx context.Context, key string) error {
	return RDB.Del(ctx, PrefixLoginFail+key).Err()
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	if RDB != nil {
		return RDB.Close()
	}
	return nil
}
