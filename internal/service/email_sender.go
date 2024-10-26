package service

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailService(host, port, username, password, from string) *EmailService {
	return &EmailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (e *EmailService) SendVerificationEmail(to, url string) error {
	subject := "Verify Your Email"
	bodyText := fmt.Sprintf("Please verify your email using this url: \n%s\n", url)

	tmpl, err := template.New("verify_email.html").ParseFiles("internal/templates/verify_email.html")
	if err != nil {
		return fmt.Errorf("template parse error: %v", err)
	}

	var bodyHTML bytes.Buffer
	data := struct {
		URL string
	}{
		URL: url,
	}

	err = tmpl.Execute(&bodyHTML, data)
	if err != nil {
		return err
	}

	return e.sendEmail(to, subject, bodyText, bodyHTML.String())
}

func (e *EmailService) SendResetPasswordEmail(to, url string) error {
	subject := "Reset Your Password"
	bodyText := fmt.Sprintf("Click the following link to reset your password: \n%s\n", url)

	tmpl, err := template.New("reset_password.html").ParseFiles("internal/templates/reset_password.html")
	if err != nil {
		return fmt.Errorf("template parse error: %v", err)
	}

	var bodyHTML bytes.Buffer
	data := struct {
		URL string
	}{
		URL: url,
	}

	err = tmpl.Execute(&bodyHTML, data)
	if err != nil {
		return err
	}

	return e.sendEmail(to, subject, bodyText, bodyHTML.String())
}

func (e *EmailService) sendEmail(to, subject, bodyText, bodyHTML string) error {
	auth := smtp.PlainAuth("", e.username, e.password, e.host)

	displayFrom := fmt.Sprintf("Kuchak <%s>", e.from)

	// Create email message with proper headers and MIME parts
	boundary := "unique-boundary-1234"
	var buffer bytes.Buffer

	// Headers must come first
	buffer.WriteString(fmt.Sprintf("From: %s\r\n", displayFrom))
	buffer.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buffer.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buffer.WriteString("MIME-Version: 1.0\r\n")
	buffer.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q\r\n", boundary))
	buffer.WriteString("\r\n") // Empty line to separate headers from body

	// Plain text part
	buffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buffer.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buffer.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	buffer.WriteString("\r\n")
	buffer.WriteString(bodyText)
	buffer.WriteString("\r\n")

	// HTML part
	buffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buffer.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	buffer.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buffer.WriteString("\r\n")
	buffer.WriteString(bodyHTML)
	buffer.WriteString("\r\n")

	// Close boundary
	buffer.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	addr := e.host + ":" + e.port

	return smtp.SendMail(addr, auth, e.from, []string{to}, buffer.Bytes())
}
