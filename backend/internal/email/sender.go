package email

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type Config struct {
	Mode     string
	From     string
	Host     string
	Port     string
	Username string
	Password string
}

type Sender interface {
	SendOTP(to, code string) error
}

func NewSender(cfg Config) (Sender, error) {
	switch strings.ToLower(cfg.Mode) {
	case "log":
		return logSender{}, nil
	case "smtp":
		if cfg.From == "" || cfg.Host == "" || cfg.Port == "" {
			return nil, errors.New("EMAIL_FROM, SMTP_HOST and SMTP_PORT are required in smtp mode")
		}

		return &smtpSender{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unsupported EMAIL_MODE %q", cfg.Mode)
	}
}

type logSender struct{}

func (logSender) SendOTP(to, code string) error {
	log.Printf("development OTP for %s: %s", to, code)

	return nil
}

type smtpSender struct {
	cfg Config
}

func (sender *smtpSender) SendOTP(to, code string) error {
	address := sender.cfg.Host + ":" + sender.cfg.Port

	var auth smtp.Auth

	if sender.cfg.Username != "" {
		auth = smtp.PlainAuth("", sender.cfg.Username, sender.cfg.Password, sender.cfg.Host)
	}

	body := "From: " + sender.cfg.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: Your sign-in code\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		"Your sign-in code is: " + code + "\r\n" +
		"This code expires soon. If you did not request it, ignore this email.\r\n"

	return smtp.SendMail(address, auth, sender.cfg.From, []string{to}, []byte(body))
}
