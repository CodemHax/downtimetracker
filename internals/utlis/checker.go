package utlis

import (
	"context"
	"crypto/tls"
	"downtimetrscker/internals/database/mongo"
	"downtimetrscker/internals/database/redis"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"downtimetrscker/internals/mail"
)

const (
	StatusUp   = "UP"
	StatusDown = "DOWN"
)

func getStatusKey(email, url string) string {
	return fmt.Sprintf("status:%s:%s", email, url)
}

func getLastStatus(ctx context.Context, email, url string) string {
	key := getStatusKey(email, url)
	status, err := redis.RDB.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	return status
}

func setStatus(ctx context.Context, email, url, status string) {
	key := getStatusKey(email, url)
	redis.RDB.Set(ctx, key, status, 24*time.Hour)
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
	},
}

func CheckServiceDowntime(url string, recipientEmail string) error {
	ctx := context.Background()

	timeoutStr := os.Getenv("CHECKER_TIMEOUT")
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			httpClient.Timeout = time.Duration(t) * time.Second
		}
	}

	if len(url) < 8 || url[:8] != "https://" {
		return fmt.Errorf("only HTTPS URLs are supported")
	}

	lastStatus := getLastStatus(ctx, recipientEmail, url)

	resp, err := httpClient.Get(url)
	var currentStatus string
	var errMsg string

	if err != nil {
		currentStatus = StatusDown
		errMsg = err.Error()
		if os.IsTimeout(err) {
			errMsg = "timeout: " + errMsg
		}
	} else {
		defer func() {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()

		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			currentStatus = StatusDown
			errMsg = fmt.Sprintf("received HTTP %d", resp.StatusCode)
		} else {
			currentStatus = StatusUp
		}
	}

	if currentStatus != lastStatus {
		setStatus(ctx, recipientEmail, url, currentStatus)

		if currentStatus == StatusDown {
			log.Printf("🔴 Service %s is DOWN for %s (was %s). Error: %s", url, recipientEmail, lastStatus, errMsg)
			_ = mail.SendMail(recipientEmail, fmt.Sprintf(
				"<h2>🔴 Service Down Alert</h2>"+
					"<p>Your service <strong>%s</strong> is <strong>DOWN</strong>.</p>"+
					"<p>Error: %s</p>"+
					"<p>Time: %s</p>",
				url, errMsg, time.Now().Format(time.RFC1123)))
		} else if currentStatus == StatusUp && lastStatus == StatusDown {
			log.Printf("🟢 Service %s is RECOVERED for %s", url, recipientEmail)
			_ = mail.SendMail(recipientEmail, fmt.Sprintf(
				"<h2>🟢 Service Recovered</h2>"+
					"<p>Your service <strong>%s</strong> is <strong>BACK UP</strong>.</p>"+
					"<p>Time: %s</p>",
				url, time.Now().Format(time.RFC1123)))
		}
	} else {
		if currentStatus == StatusDown {
			log.Printf("Service %s still DOWN for %s (no repeat email)", url, recipientEmail)
		} else {
			log.Printf("Service %s is UP for %s", url, recipientEmail)
		}
	}

	if currentStatus == StatusDown {
		return fmt.Errorf("service is down: %s", errMsg)
	}
	return nil
}

type job struct {
	site  string
	email string
}

func CheckAllWebsites() {
	log.Println("Starting check for all websites")
	users, err := mongo.GetAllUsersWithWebsites()
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		return
	}
	concurrencyStr := os.Getenv("CHECKER_CONCURRENCY")
	concurrency := 5
	if concurrencyStr != "" {
		if c, err := strconv.Atoi(concurrencyStr); err == nil && c > 0 {
			concurrency = c
		}
	}
	jobs := make(chan job, 100)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				log.Printf("Checking website %s for %s", j.site, j.email)
				_ = CheckServiceDowntime(j.site, j.email)
			}
		}()
	}
	for _, user := range users {
		for _, website := range user.Websites {
			jobs <- job{site: website, email: user.Email}
		}
	}
	close(jobs)
	wg.Wait()
	log.Println("Completed check for all websites")
}
