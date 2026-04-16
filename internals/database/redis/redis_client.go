package redis

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	RDB *redis.Client
)

func ConnectRedis() {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	addr := host + ":" + port
	db := 0
	if dbStr != "" {
		if val, err := strconv.Atoi(dbStr); err == nil {
			db = val
		}
	}
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	}
}
