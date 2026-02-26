package cmd

import (
	"downtimetrscker/internals/database/mongo"
	"downtimetrscker/internals/database/redis"
	"downtimetrscker/internals/mail"
	"downtimetrscker/internals/models"
	"downtimetrscker/internals/utlis"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// AddWeb godoc
// @Summary      Add a website
// @Description  Adds a website for monitoring and sends verification email
// @Tags         website
// @Accept       json
// @Produce      json
// @Param        request body models.WebIn true "Website info"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Router       /add [post]
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
		body := `<h2>Downtime Tracker - Verify Your Email</h2>
<p>Please verify your email address by clicking the link below:</p>
<p><a href="` + verifyLink + `">Click here to verify</a></p>
<p>Or copy this link: ` + verifyLink + `</p>
<p><strong>This link is valid for 5 minutes.</strong></p>`

		if err := mail.SendMail(req.Email, body); err != nil {
			log.Printf("[ERROR] Failed to send verification email to %s: %v", req.Email, err)
			context.JSON(500, gin.H{"error": "Failed to send verification email: " + err.Error()})
			return
		}
		log.Printf("[INFO] Verification email sent to %s", req.Email)
		context.JSON(200, gin.H{"message": "Please verify your email. Check your inbox for the verification link."})
	}

}

// Verify godoc
// @Summary      Verify email
// @Description  Verifies a user's email using a token
// @Tags         website
// @Produce      json
// @Param        email query string true "User Email"
// @Param        token query string true "Verification Token"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Router       /verify [get]
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

// GetWebsites godoc
// @Summary      Get all websites for a user
// @Description  Returns all websites associated with the given email
// @Tags         website
// @Produce      json
// @Param        email query string true "User Email"
// @Success      200 {object} map[string][]string
// @Failure      400 {object} map[string]string
// @Router       /websites [get]
func GetWebsites(context *gin.Context) {
	email := context.Query("email")
	if email == "" {
		context.JSON(400, gin.H{"error": "Email is required"})
		return
	}
	websites, err := mongo.GetWebsites(email)
	if err != nil {
		context.JSON(400, gin.H{"error": err.Error()})
		return
	}
	context.JSON(200, gin.H{"websites": websites})
}

// DeleteWebsite godoc
// @Summary      Delete a website for a user
// @Description  Removes a website from the user's list
// @Tags         website
// @Accept       json
// @Produce      json
// @Param        request body models.WebIn true "Delete Website"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Router       /deleteweb [delete]
func DeleteWebsite(context *gin.Context) {
	var req struct {
		Email   string `json:"email"`
		Website string `json:"website"`
	}
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(400, gin.H{"error": err.Error()})
		return
	}
	err := mongo.RemoveWebsite(req.Email, req.Website)
	if err != nil {
		context.JSON(400, gin.H{"error": err.Error()})
		return
	}
	context.JSON(200, gin.H{"message": "Website deleted successfully"})
}

// ForceCheck godoc
// @Summary      Force check all websites
// @Description  Manually initiates a background check for all verified websites
// @Tags         website
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /force-check [post]
func ForceCheck(context *gin.Context) {
	go utlis.CheckAllWebsites()
	context.JSON(200, gin.H{"message": "Website check initiated in the background"})
}
