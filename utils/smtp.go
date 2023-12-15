package utils

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
)

type SmtpPoster struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewSmtpPoster(host string, port int, username string, password string, from string) *SmtpPoster {
	return &SmtpPoster{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func (s *SmtpPoster) SendMail(to string, subject string, body string) {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.From, s.Password, s.Host)
	err := smtpRequestWithTLS(addr, auth, s.From, []string{to},
		[]byte(formatMail(map[string]string{
			"From":         fmt.Sprintf("%s <%s>", s.Username, s.From),
			"To":           to,
			"Subject":      subject,
			"Content-Type": "text/html; charset=utf-8",
		}, body)))
	if err != nil {
		return
	}
}

func dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, err
	}

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func formatMail(headers map[string]string, body string) (result string) {
	for k, v := range headers {
		result += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	return fmt.Sprintf("%s\r\n%s", result, body)
}

func smtpRequestWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) (err error) {
	client, err := dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err = client.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = writer.Write(msg); err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}
