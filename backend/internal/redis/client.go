package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/redis/go-redis/v9"
)

// Client wraps Redis client functionality
type Client struct {
	client  *redis.Client
	enabled bool
	ctx     context.Context
}

// NewClient creates a new Redis client
func NewClient(cfg *config.Config) *Client {
	if !cfg.Redis.Enabled {
		log.Println("Redis is disabled, WebSocket messages will not sync across regions")
		return &Client{
			enabled: false,
			ctx:     context.Background(),
		}
	}

	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	options := &redis.Options{
		Addr:         addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	}

	// Enable TLS if configured (for Upstash or secure connections)
	if cfg.Redis.TLS {
		options.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		log.Println("Redis TLS enabled")
	}

	rdb := redis.NewClient(options)

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("WARNING: Failed to connect to Redis at %s: %v. WebSocket sync disabled.", addr, err)
		return &Client{
			enabled: false,
			ctx:     ctx,
		}
	}

	log.Printf("Redis connected successfully at %s", addr)
	return &Client{
		client:  rdb,
		enabled: true,
		ctx:     ctx,
	}
}

// IsEnabled returns whether Redis is enabled
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// Publish publishes a message to a channel
func (c *Client) Publish(channel string, message []byte) error {
	if !c.enabled {
		return nil // Silently skip if Redis is disabled
	}

	return c.client.Publish(c.ctx, channel, message).Err()
}

// Subscribe subscribes to a channel and returns a pubsub
func (c *Client) Subscribe(channel string) *redis.PubSub {
	if !c.enabled {
		return nil
	}

	return c.client.Subscribe(c.ctx, channel)
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if !c.enabled || c.client == nil {
		return nil
	}

	return c.client.Close()
}

// Ping checks if Redis is still connected
func (c *Client) Ping() error {
	if !c.enabled {
		return nil
	}

	return c.client.Ping(c.ctx).Err()
}
