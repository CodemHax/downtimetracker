package utils

import (
	"context"
	"crypto/tls"
	"downtimetracker/internals/database/mongo"
	"downtimetracker/internals/database/redis"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"downtimetracker/internals/mail"

	goredis "github.com/redis/go-redis/v9"
)

const (
	StatusUp   = "UP"
	StatusDown = "DOWN"
)

func getStatusKey(email, url string) string {
	return fmt.Sprintf("status:%s:%s", email, url)
}

func GetLastStatus(ctx context.Context, email, url string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	key := getStatusKey(email, url)
	status, err := redis.RDB.Get(timeoutCtx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return "", nil
		}
		return "", err
	}
	return status, nil
}

func setStatus(ctx context.Context, email, url, status string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	key := getStatusKey(email, url)
	return redis.RDB.Set(timeoutCtx, key, status, 0).Err()
}

var httpClient *http.Client

func init() {
	timeout := 10 * time.Second
	if t, err := strconv.Atoi(os.Getenv("CHECKER_TIMEOUT")); err == nil && t > 0 {
		timeout = time.Duration(t) * time.Millisecond
	}
	httpClient = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		},
	}
}

func CheckServiceDowntime(url string, recipientEmail string) error {
	ctx := context.Background()

	if len(url) < 7 || (url[:7] != "http://" && (len(url) < 8 || url[:8] != "https://")) {
		return fmt.Errorf("invalid URL: must start with http:// or https://")
	}

	lastStatus, errGet := GetLastStatus(ctx, recipientEmail, url)
	if errGet != nil {
		log.Printf("[WARNING] Skipping check for %s: Redis lookup failed: %v", url, errGet)
		return errGet
	}

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
			err := resp.Body.Close()
			if err != nil {
				return
			}
		}()

		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			currentStatus = StatusDown
			errMsg = fmt.Sprintf("received HTTP %d", resp.StatusCode)
		} else {
			currentStatus = StatusUp
		}
	}

	if currentStatus != lastStatus {
		_ = setStatus(ctx, recipientEmail, url, currentStatus)

		escapedURL := html.EscapeString(url)
		if currentStatus == StatusDown {
			escapedErrMsg := html.EscapeString(errMsg)
			log.Printf("🔴 Service %s is DOWN for %s (was %s). Error: %s", url, recipientEmail, lastStatus, errMsg)
			if err := mail.SendMail(recipientEmail, fmt.Sprintf(
				"<h2>🔴 Service Down Alert</h2>"+
					"<p>Your service <strong>%s</strong> is <strong>DOWN</strong>.</p>"+
					"<p>Error: %s</p>"+
					"<p>Time: %s</p>",
				escapedURL, escapedErrMsg, time.Now().Format(time.RFC1123))); err != nil {
				log.Printf("[ERROR] Failed to send DOWN alert to %s: %v", recipientEmail, err)
			}
		} else if currentStatus == StatusUp && lastStatus == StatusDown {
			log.Printf("🟢 Service %s is RECOVERED for %s", url, recipientEmail)
			if err := mail.SendMail(recipientEmail, fmt.Sprintf(
				"<h2>🟢 Service Recovered</h2>"+
					"<p>Your service <strong>%s</strong> is <strong>BACK UP</strong>.</p>"+
					"<p>Time: %s</p>",
				escapedURL, time.Now().Format(time.RFC1123))); err != nil {
				log.Printf("[ERROR] Failed to send RECOVERY alert to %s: %v", recipientEmail, err)
			}
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
