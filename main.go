package main

import (
	"context"
	"downtimetracker/cmd"
	_ "downtimetracker/docs"
	"downtimetracker/internals/database/mongo"
	"downtimetracker/internals/database/redis"
	"downtimetracker/internals/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	r.Use(cmd.CORSMiddleware())
	r.Use(static.Serve("/", static.LocalFile("./frontend/html", false)))
	r.Use(static.Serve("/css", static.LocalFile("./frontend/css", false)))
	r.Use(static.Serve("/js", static.LocalFile("./frontend/js", false)))

	r.GET("/config.js", func(c *gin.Context) {
		link := os.Getenv("LINK")
		if link == "" {
			link = "http://localhost:8080"
		}
		c.Header("Content-Type", "application/javascript")
		c.String(200, "window.API_URL = '%s';", link)
	})

	r.GET("/ping", func(context *gin.Context) {
		context.JSON(200, gin.H{"message": "pong"})
	})
	r.GET("/verify", cmd.Verify)
	r.POST("/auth/register", cmd.Register)
	r.POST("/auth/login", cmd.Login)
	r.POST("/auth/logout", cmd.Logout)

	protected := r.Group("/")
	protected.Use(cmd.AuthMiddleware())
	protected.POST("/add", cmd.AddWeb)
	protected.GET("/websites", cmd.GetWebsites)
	protected.DELETE("/deleteweb", cmd.DeleteWebsite)
	protected.POST("/force-check", cmd.ForceCheck)
	protected.GET("/auth/me", cmd.GetMe)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("Running initial website check")
	utils.CheckAllWebsites()
	log.Println("Initial check completed")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping scheduled checks")
				return
			case <-ticker.C:
				log.Println("Starting scheduled website check")
				utils.CheckAllWebsites()
				log.Println("Scheduled check completed")
			}
		}
	}()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
