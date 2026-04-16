package cmd

import (
	"downtimetracker/internals/database/mongo"
	"downtimetracker/internals/database/redis"
	"downtimetracker/internals/mail"
	"downtimetracker/internals/models"
	"downtimetracker/internals/utils"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Register godoc
// @Summary      Register User
// @Description  Create a new user account with email and password, and send verification email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.AuthIn true "Credentials"
// @Success      200 {object} models.AuthOut
// @Failure      400 {object} map[string]string
// @Router       /auth/register [post]
func Register(context *gin.Context) {
	var req models.AuthIn
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	err = mongo.RegisterUser(req.Email, string(hash))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Email Verification Logic
	token := utils.TokenGen()
	ctx := context.Request.Context()
	errSet := redis.RDB.Set(ctx, req.Email, token, 350*time.Second).Err()
	if errSet != nil {
		context.JSON(500, gin.H{"error": "error saving OTP"})
		return
	}

	uri := os.Getenv("LINK")
	if uri == "" {
		uri = "http://localhost:8080"
	}
	verifyLink := uri + "/verify?email=" + req.Email + "&token=" + url.QueryEscape(token)
	body := `<h2>Downtime Tracker - Verify Your Account</h2>
<p>Please verify your email address by clicking the link below:</p>
<p><a href="` + verifyLink + `">Click here to verify</a></p>
<p><strong>This link is valid for 5 minutes.</strong></p>`

	if err := mail.SendMail(req.Email, body); err != nil {
		log.Printf("[ERROR] Failed to send verification email: %v", err)
	}

	context.JSON(200, models.AuthOut{Message: "Registration successful. Please check your email to verify your account."})
}

// Login godoc
// @Summary      Login User
// @Description  Login with email and password to receive JWT cookies
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.AuthIn true "Credentials"
// @Success      200 {object} models.AuthOut
// @Failure      400 {object} map[string]string
// @Router       /auth/login [post]
func Login(context *gin.Context) {
	var req models.AuthIn
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, verified, err := mongo.GetUserAuth(req.Email)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !verified {
		context.JSON(http.StatusForbidden, gin.H{"error": "Email not verified. Please check your inbox."})
		return
	}

	access, refresh, err := utils.GenerateTokens(req.Email)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	ctx := context.Request.Context()
	err = redis.RDB.Set(ctx, "session:"+req.Email, refresh, 7*24*time.Hour).Err()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Session deployment failed"})
		return
	}

	context.SetCookie("access_token", access, 15*60, "/", "", false, true)
	context.SetCookie("refresh_token", refresh, 7*24*60*60, "/", "", false, true)

	context.JSON(200, models.AuthOut{Message: "Successfully logged in"})
}

// Logout godoc
// @Summary      Logout User
// @Description  Deletes session from Redis and clears cookies
// @Tags         auth
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /auth/logout [post]
func Logout(context *gin.Context) {
	email := context.GetString("email")
	if email != "" {
		ctx := context.Request.Context()
		redis.RDB.Del(ctx, "session:"+email)
	}

	context.SetCookie("access_token", "", -1, "/", "", false, true)
	context.SetCookie("refresh_token", "", -1, "/", "", false, true)

	context.JSON(200, gin.H{"message": "Logged out successfully"})
}

// GetMe godoc
// @Summary      Get current user
// @Description  Returns the email of the currently authenticated user based on the secure cookie
// @Tags         auth
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /auth/me [get]
func GetMe(context *gin.Context) {
	email := context.GetString("email")
	if email == "" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized session"})
		return
	}
	context.JSON(200, gin.H{"email": email})
}
