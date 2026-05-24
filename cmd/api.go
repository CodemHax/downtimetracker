package cmd

import (
	"downtimetracker/internals/database/mongo"
	"downtimetracker/internals/database/redis"
	"downtimetracker/internals/models"
	"downtimetracker/internals/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)


func AddWeb(context *gin.Context) {
	var req models.WebIn

	err := context.ShouldBindJSON(&req)
	if err != nil {
		context.JSON(400, gin.H{"error": err.Error()})
		return
	}

	email := context.GetString("email")
	if email == "" {
		context.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	websiteURL := req.Website
	if websiteURL != "" && !strings.HasPrefix(websiteURL, "http://") && !strings.HasPrefix(websiteURL, "https://") {
		websiteURL = "https://" + websiteURL
	}

	if mongo.WebsiteExists(email, websiteURL) {
		context.JSON(400, gin.H{"error": "Website already added"})
		return
	}

	err = mongo.AddWebsite(email, websiteURL)

	if err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}

	context.JSON(200, gin.H{"message": "Website Added successfully"})
}


func Verify(context *gin.Context) {
	email := context.Query("email")
	token := context.Query("token")

	ctx := context.Request.Context()
	storedToken, err := redis.RDB.Get(ctx, email).Result()
	if err != nil || storedToken != token {
		htmlErr := `<!DOCTYPE html>
<html>
<head>
    <title>Link Expired</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600&display=swap" rel="stylesheet">
    <style>
        body { background: #000; color: #fff; font-family: 'Inter', sans-serif; display: flex; align-items: center; justify-content: center; height: 100vh; text-align: center; margin: 0; }
        .box { max-width: 400px; padding: 40px; }
        h2 { font-weight: 500; font-size: 2rem; margin-bottom: 10px; color: #ff3333; }
        p { color: #888; font-size: 0.9rem; }
        a { color: #fff; text-decoration: none; border-bottom: 1px solid #fff; }
    </style>
</head>
<body>
    <div class="box">
        <h2>Link Expired</h2>
        <p>This verification link is invalid or has expired.</p>
        <p>Please <a href="/register.html">register</a> again.</p>
    </div>
</body>
</html>`
		context.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(htmlErr))
		return
	}

	err = mongo.VerifyEmail(email)
	if err != nil {
		context.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(`<h1>Server Error</h1>`))
		return
	}

	_, err = redis.RDB.Del(ctx, email).Result()
	if err != nil {
		context.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(`<h1>Server Error</h1>`))
		return
	}

	htmlSuccess := `<!DOCTYPE html>
<html>
<head>
    <title>Email Verified</title>
    <meta http-equiv="refresh" content="3;url=/login.html" />
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600&display=swap" rel="stylesheet">
    <style>
        body { background: #000; color: #fff; font-family: 'Inter', sans-serif; display: flex; align-items: center; justify-content: center; height: 100vh; text-align: center; margin: 0; }
        .box { max-width: 400px; padding: 40px; }
        h2 { font-weight: 500; font-size: 2rem; margin-bottom: 10px; }
        p { color: #888; font-size: 0.9rem; }
        a { color: #fff; text-decoration: none; border-bottom: 1px solid #fff; }
    </style>
</head>
<body>
    <div class="box">
        <h2>Verified</h2>
        <p>Your email has been successfully verified.</p>
        <p>Redirecting to <a href="/login.html">login</a>...</p>
    </div>
</body>
</html>`

	context.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlSuccess))
}


func GetWebsites(context *gin.Context) {
	email := context.GetString("email")

	websites, err := mongo.GetWebsites(email)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx := context.Request.Context()
	var sites []map[string]string

	for _, url := range websites {
		status, err := utils.GetLastStatus(ctx, email, url)
		if err != nil || status == "" {
			status = "PENDING"
		}
		sites = append(sites, map[string]string{
			"url":    url,
			"status": status,
		})
	}

	context.JSON(http.StatusOK, gin.H{"websites": sites})
}


func DeleteWebsite(context *gin.Context) {
	email := context.GetString("email")
	website := context.Query("website")
	
	if email == "" || website == "" {
		context.JSON(400, gin.H{"error": "website query parameter is required"})
		return
	}
	
	err := mongo.RemoveWebsite(email, website)
	if err != nil {
		context.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	ctx := context.Request.Context()
	statusKey := "status:" + email + ":" + website
	redis.RDB.Del(ctx, statusKey)

	context.JSON(200, gin.H{"message": "Website deleted successfully"})
}


func ForceCheck(context *gin.Context) {
	go utils.CheckAllWebsites()
	context.JSON(200, gin.H{"message": "Website check initiated in the background"})
}
