package cmd

import (
	"downtimetrscker/internals/database/mongo"
	"downtimetrscker/internals/database/redis"
	"downtimetrscker/internals/mail"
	"downtimetrscker/internals/models"
	"downtimetrscker/internals/utlis"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func AddWeb(context *gin.Context) {
	var req models.WebIn

	err := context.ShouldBindJSON(&req)
	if err != nil {
		context.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if mongo.WebsiteExists(req.Email, req.Website) {
		context.JSON(400, gin.H{"error": "Website already added"})
		return
	}

	token := utlis.TokenGen()
	ctx := context.Request.Context()
	errSet := redis.RDB.Set(ctx, req.Email, token, 350*time.Second).Err()
	if errSet != nil {
		context.JSON(500, gin.H{
			"error": "error saving OTP: " + errSet.Error(),
		})
		return
	}
	err = mongo.AddWebsite(req.Email, req.Website)

	if err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if mongo.IsVerified(req.Email) {
		context.JSON(200, gin.H{"message": "Website Added successfully"})
	} else {

		uri := os.Getenv("LINK")

		if uri == "" {
			uri = "http://localhost:8080"
		}
		verifyLink := uri + "/verify?email=" + req.Email + "&token=" + url.QueryEscape(token)
		body := "Please verify your email by clicking the link: " + verifyLink + "Valid for five "
		go func() {
			err := mail.SendMail(req.Email, body)
			if err != nil {

			}
		}()
		context.JSON(200, gin.H{"message": "Please verify your email first"})
	}

}

func Verify(context *gin.Context) {
	email := context.Query("email")
	token := context.Query("token")

	ctx := context.Request.Context()
	storedToken, err := redis.RDB.Get(ctx, email).Result()
	if err != nil || storedToken != token {
		context.JSON(400, gin.H{"error": "Invalid or expired token"})
		return
	}

	err = mongo.VerifyEmail(email)
	if err != nil {
		context.JSON(500, gin.H{"error": "Failed to verify email"})
		return
	}

	_, err = redis.RDB.Del(ctx, email).Result()
	if err != nil {
		context.JSON(500, gin.H{"error": "Failed to Verify email Redis"})
		return
	}
	context.JSON(200, gin.H{"message": "Email verified successfully"})
}
