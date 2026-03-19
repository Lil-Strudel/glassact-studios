package app

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/mail"
	"net/smtp"
	"strconv"
	"time"
)

type Mailer struct {
	host     string
	port     int
	username string
	password string
}

func NewMailer(host string, port int, username, password string) *Mailer {
	return &Mailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

func (m *Mailer) Send(to, subject, htmlBody, textBody string) error {
	from := mail.Address{Name: "GlassAct Studios", Address: "no-reply@glassactstudios.com"}
	toAddr := mail.Address{Address: to}

	message := buildMessage(from, toAddr, subject, textBody, htmlBody)

	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	return smtp.SendMail(m.host+":"+strconv.Itoa(m.port), auth, from.Address, []string{toAddr.Address}, message)
}

func randString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func buildMessage(from, to mail.Address, subject, textBody, htmlBody string) []byte {
	msgID := fmt.Sprintf("<%s@glassactstudios.com>", randString(12))
	date := time.Now().Format(time.RFC1123Z)
	boundary := "alt-" + randString(12)

	headers := ""
	headers += fmt.Sprintf("From: %s\r\n", from.String())
	headers += fmt.Sprintf("To: %s\r\n", to.String())
	headers += fmt.Sprintf("Subject: %s\r\n", subject)
	headers += "MIME-Version: 1.0\r\n"
	headers += fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	headers += fmt.Sprintf("Date: %s\r\n", date)
	headers += fmt.Sprintf("Message-ID: %s\r\n", msgID)

	body := ""
	body += fmt.Sprintf("--%s\r\n", boundary)
	body += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	body += "Content-Transfer-Encoding: 7bit\r\n"
	body += "\r\n"
	body += textBody + "\r\n"

	if htmlBody != "" {
		body += fmt.Sprintf("--%s\r\n", boundary)
		body += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
		body += "Content-Transfer-Encoding: 7bit\r\n"
		body += "\r\n"
		body += htmlBody + "\r\n"
	}

	body += fmt.Sprintf("--%s--\r\n", boundary)

	return []byte(headers + "\r\n" + body)
}
