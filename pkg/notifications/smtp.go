package notifications

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
)

// Compile-time check to ensure SMTPProvider implements EmailProvider interface
var _ EmailProvider = (*SMTPProvider)(nil)

// SMTPProvider implements email sending via SMTP
type SMTPProvider struct {
	host       string
	port       int
	username   string
	password   string
	fromEmail  string
	fromName   string
	encryption string // tls, starttls, none
}

// NewSMTPProvider creates a new SMTP email provider
func NewSMTPProvider(host string, port int, username, password, fromEmail, fromName, encryption string) *SMTPProvider {
	return &SMTPProvider{
		host:       host,
		port:       port,
		username:   username,
		password:   password,
		fromEmail:  fromEmail,
		fromName:   fromName,
		encryption: encryption,
	}
}

// SendEmail sends an email via SMTP
func (p *SMTPProvider) SendEmail(ctx context.Context, to, subject, htmlBody string) error {
	// Build email headers and body
	from := p.fromEmail
	if p.fromName != "" {
		from = fmt.Sprintf("%s <%s>", p.fromName, p.fromEmail)
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", p.host, p.port)

	// Setup authentication
	var auth smtp.Auth
	if p.username != "" && p.password != "" {
		auth = smtp.PlainAuth("", p.username, p.password, p.host)
	}

	// Send email based on encryption type
	switch p.encryption {
	case "tls":
		return p.sendWithTLS(addr, auth, to, []byte(message))
	case "starttls":
		return p.sendWithSTARTTLS(addr, auth, to, []byte(message))
	case "none", "":
		return p.sendPlain(addr, auth, to, []byte(message))
	default:
		return fmt.Errorf("unsupported encryption type: %s", p.encryption)
	}
}

// sendWithTLS sends email using direct TLS connection
func (p *SMTPProvider) sendWithTLS(addr string, auth smtp.Auth, to string, msg []byte) error {
	// Create TLS config
	tlsConfig := &tls.Config{
		ServerName: p.host,
		MinVersion: tls.VersionTLS12,
	}

	// Connect with TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect with TLS: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, p.host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Auth
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Set sender
	if err = client.Mail(p.fromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send data
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// sendWithSTARTTLS sends email using STARTTLS
func (p *SMTPProvider) sendWithSTARTTLS(addr string, auth smtp.Auth, to string, msg []byte) error {
	// Connect
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// STARTTLS
	tlsConfig := &tls.Config{
		ServerName: p.host,
		MinVersion: tls.VersionTLS12,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	// Auth
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Set sender
	if err = client.Mail(p.fromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send data
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// sendPlain sends email without encryption (not recommended for production)
func (p *SMTPProvider) sendPlain(addr string, auth smtp.Auth, to string, msg []byte) error {
	return smtp.SendMail(addr, auth, p.fromEmail, []string{to}, msg)
}

// GetName returns the provider name
func (p *SMTPProvider) GetName() string {
	return "smtp"
}
