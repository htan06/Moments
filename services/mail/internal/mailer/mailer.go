// Package mailer sends email through an SMTP server, logging in with a
// username and password (for Gmail: the account and an app password).
//
// Port 465 uses implicit TLS; any other port must offer STARTTLS. Credentials
// are never sent over an unencrypted connection.
package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"
)

const implicitTLSPort = "465"

type Config struct {
	Host     string
	Port     string
	Username string
	Password string // app password: never log
	// From is the sender, e.g. "Moments <moments.app@gmail.com>". Gmail only
	// accepts the account's own address or a verified alias.
	From    string
	Timeout time.Duration // bounds one whole SMTP conversation
}

// Message is one email to send.
type Message struct {
	To      string // bare address
	Subject string
	Text    string
	HTML    string
}

type Mailer struct {
	cfg  Config
	from *mail.Address
	tls  *tls.Config
}

func New(cfg Config) (*Mailer, error) {
	from, err := mail.ParseAddress(cfg.From)
	if err != nil {
		return nil, errors.New("mailer: invalid from address")
	}
	if cfg.Timeout <= 0 {
		return nil, errors.New("mailer: timeout must be > 0")
	}
	return &Mailer{
		cfg:  cfg,
		from: from,
		tls:  &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12},
	}, nil
}

// Verify connects and logs in, so a wrong host or app password fails at
// startup instead of on the first task.
func (m *Mailer) Verify(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, m.cfg.Timeout)
	defer cancel()
	c, done, err := m.open(ctx)
	if err != nil {
		return err
	}
	defer done()
	_ = c.Quit()
	return nil
}

// Send delivers msg. A nil error means the server accepted the message.
func (m *Mailer) Send(ctx context.Context, msg Message) error {
	if strings.ContainsAny(msg.To, "\r\n") {
		return errors.New("mailer: recipient contains a line break")
	}
	body, err := build(m.from, msg, time.Now())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, m.cfg.Timeout)
	defer cancel()
	c, done, err := m.open(ctx)
	if err != nil {
		return err
	}
	defer done()

	if err := c.Mail(m.from.Address); err != nil {
		return wrap("mail", err)
	}
	if err := c.Rcpt(msg.To); err != nil {
		return wrap("rcpt", err)
	}
	w, err := c.Data()
	if err != nil {
		return wrap("data", err)
	}
	if _, err := w.Write(body); err != nil {
		return wrap("data", err)
	}
	// Close sends the final "." and reads the server's verdict on the message.
	if err := w.Close(); err != nil {
		return wrap("data", err)
	}
	// The message is accepted at this point; a failed QUIT does not undo it.
	_ = c.Quit()
	return nil
}

// open dials, switches to TLS and logs in. done releases the connection.
func (m *Mailer) open(ctx context.Context) (*smtp.Client, func(), error) {
	conn, err := m.dial(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("mailer: dial: %w", err)
	}
	// net/smtp has no context support: the deadline bounds every read and
	// write, and cancelling ctx (shutdown) closes the connection.
	if dl, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(dl)
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })

	c, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		stop()
		_ = conn.Close()
		return nil, nil, wrap("greeting", err)
	}
	done := func() {
		stop()
		_ = c.Close()
	}
	if m.cfg.Port != implicitTLSPort {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			done()
			return nil, nil, errors.New("mailer: server does not offer STARTTLS")
		}
		if err := c.StartTLS(m.tls); err != nil {
			done()
			return nil, nil, wrap("starttls", err)
		}
	}
	// PlainAuth also refuses to send the password without TLS.
	if err := c.Auth(smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)); err != nil {
		done()
		return nil, nil, wrap("auth", err)
	}
	return c, done, nil
}

func (m *Mailer) dial(ctx context.Context) (net.Conn, error) {
	addr := net.JoinHostPort(m.cfg.Host, m.cfg.Port)
	if m.cfg.Port == implicitTLSPort {
		d := tls.Dialer{Config: m.tls}
		return d.DialContext(ctx, "tcp", addr)
	}
	var d net.Dialer
	return d.DialContext(ctx, "tcp", addr)
}

// SMTPError is an error reply from the server. It keeps only the stage and
// the reply code, not the server's text: that text can echo the recipient
// address, which must not reach the logs.
type SMTPError struct {
	Stage string // greeting, starttls, auth, mail, rcpt, data
	Code  int
}

func (e *SMTPError) Error() string {
	return fmt.Sprintf("mailer: smtp %s: reply %d", e.Stage, e.Code)
}

func wrap(stage string, err error) error {
	var tp *textproto.Error
	if errors.As(err, &tp) {
		return &SMTPError{Stage: stage, Code: tp.Code}
	}
	return fmt.Errorf("mailer: smtp %s: %w", stage, err)
}

// IsRejected reports whether the server permanently refused this message:
// a 5xx reply to the recipient or the content (unknown mailbox, message
// refused). Retrying will not help.
//
// A 5xx at greeting, auth or MAIL FROM (wrong app password, sender not
// allowed) is a configuration problem that hits every message, so it is not
// counted: such tasks are retried until they expire, and the logs show why.
func IsRejected(err error) bool {
	var e *SMTPError
	return errors.As(err, &e) && e.Code >= 500 && (e.Stage == "rcpt" || e.Stage == "data")
}
