package mail

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SendMail(to string, body string) error {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if host == "" || user == "" || pass == "" || from == "" {
		return fmt.Errorf("missing SMTP configuration environment variables")
	}

	port, err := strconv.Atoi(portStr)

	if err != nil {
		return err
	}
	mail := gomail.NewMessage()

	mail.SetHeader("From", from)
	mail.SetHeader("To", to)
	mail.SetHeader("Subject", "Downtime Tracker")
	mail.SetBody("text/html", body)

	d := gomail.NewDialer(host, port, user, pass)
	return d.DialAndSend(mail)
}
