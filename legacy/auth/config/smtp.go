package config

import (
	"log"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func GetGmailDialer() *gomail.Dialer {
	host := os.Getenv("SMTP_HOST")
	username := os.Getenv("SMTP_USERNAME")
	appPassword := os.Getenv("SMTP_APP_PASSWORD")

	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Fatal("SMTP config: " + err.Error())
	}

	return gomail.NewDialer(
		host,
		port,
		username,
		appPassword,
	)
}

type MailAddress struct {
	Name  string
	Email string
}

func GetMailAddress() *MailAddress {
	return &MailAddress{
		Name:  os.Getenv("EMAIL_NAME"),
		Email: os.Getenv("EMAIL_ADDRESS"),
	}
}
