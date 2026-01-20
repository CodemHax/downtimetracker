package main

import (
	"downtimetrscker/cmd"
	"downtimetrscker/internals/database/mongo"
	"downtimetrscker/internals/database/redis"
	"downtimetrscker/internals/utlis"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting downtime tracker service")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env: %v", err)
	}
	log.Println("Loaded environment variables")
	mongo.Init()
	log.Println("Initialized MongoDB")
	redis.ConnectRedis()
	log.Println("Connected to Redis")
	r := gin.Default()

	r.GET("/ping", func(context *gin.Context) {
		context.JSON(200, gin.H{"message": "pong"})
	})
	r.GET("/verify", cmd.Verify)
	r.POST("/add", cmd.AddWeb)

	log.Println("Running initial website check")
	utlis.CheckAllWebsites()
	log.Println("Initial check completed")
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			log.Println("Starting scheduled website check")
			utlis.CheckAllWebsites()
			log.Println("Scheduled check completed")
			<-ticker.C
		}
	}()
	log.Println("Starting server on :8080")
	err = r.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
