package repository

import (
	"context"
	"fmt"
	"time"

	"cinema-booking/config"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
	lockTTL time.Duration
}

func NewRedisRepo(cfg *config.Config) (*RedisRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &RedisRepo{
		client:  client,
		lockTTL: time.Duration(cfg.SeatLockTTL) * time.Second,
	}, nil
}

// seatLockKey returns the Redis key for a seat lock.
// Format: seat_lock:{showtimeID}:{seatCode}
func seatLockKey(showtimeID, seatCode string) string {
	return fmt.Sprintf("seat_lock:%s:%s", showtimeID, seatCode)
}

// bookingOwnerKey maps bookingID → userID for ownership verification
func bookingOwnerKey(bookingID string) string {
	return fmt.Sprintf("booking_owner:%s", bookingID)
}

// TryLockSeat attempts to acquire a distributed lock for a seat.
// Uses SET NX EX (atomic) — only one goroutine/process wins.
// Returns (true, expiresAt) if lock acquired, (false, zero) if already locked.
func (r *RedisRepo) TryLockSeat(ctx context.Context, showtimeID, seatCode, userID string) (bool, time.Time, error) {
	key := seatLockKey(showtimeID, seatCode)
	// SET key userID NX PX <ttl_ms>  — atomic: only sets if key doesn't exist
	ok, err := r.client.SetNX(ctx, key, userID, r.lockTTL).Result()
	if err != nil {
		return false, time.Time{}, fmt.Errorf("redis set nx: %w", err)
	}
	if !ok {
		return false, time.Time{}, nil // already locked
	}
	expiresAt := time.Now().Add(r.lockTTL)
	return true, expiresAt, nil
}

// GetSeatLockOwner returns the userID that holds the lock, or "" if not locked.
func (r *RedisRepo) GetSeatLockOwner(ctx context.Context, showtimeID, seatCode string) (string, error) {
	key := seatLockKey(showtimeID, seatCode)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// ReleaseSeatLock removes the lock. Only the lock owner should call this.
func (r *RedisRepo) ReleaseSeatLock(ctx context.Context, showtimeID, seatCode string) error {
	key := seatLockKey(showtimeID, seatCode)
	return r.client.Del(ctx, key).Err()
}

// GetSeatLockTTL returns remaining TTL for a seat lock.
func (r *RedisRepo) GetSeatLockTTL(ctx context.Context, showtimeID, seatCode string) (time.Duration, error) {
	key := seatLockKey(showtimeID, seatCode)
	return r.client.TTL(ctx, key).Result()
}

// SetBookingOwner stores bookingID → userID mapping.
func (r *RedisRepo) SetBookingOwner(ctx context.Context, bookingID, userID string) error {
	key := bookingOwnerKey(bookingID)
	return r.client.Set(ctx, key, userID, r.lockTTL+30*time.Second).Err()
}

// GetBookingOwner retrieves the userID for a booking.
func (r *RedisRepo) GetBookingOwner(ctx context.Context, bookingID string) (string, error) {
	key := bookingOwnerKey(bookingID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Publish publishes a message to a Redis channel.
func (r *RedisRepo) Publish(ctx context.Context, channel, message string) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to a Redis channel.
func (r *RedisRepo) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return r.client.Subscribe(ctx, channel)
}

func (r *RedisRepo) Client() *redis.Client {
	return r.client
}

func (r *RedisRepo) LockTTL() time.Duration {
	return r.lockTTL
}
