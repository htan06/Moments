package mailer

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"strings"
	"time"
)

// build renders msg as a multipart/alternative MIME message (text first,
// HTML last: clients show the last part they can display).
func build(from *mail.Address, msg Message, now time.Time) ([]byte, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	header := []string{
		"From: " + from.String(),
		"To: " + msg.To,
		"Subject: " + mime.QEncoding.Encode("utf-8", msg.Subject),
		"Date: " + now.Format(time.RFC1123Z),
		"Message-ID: " + messageID(from.Address),
		"MIME-Version: 1.0",
		"Content-Type: " + mime.FormatMediaType("multipart/alternative", map[string]string{"boundary": mw.Boundary()}),
	}
	buf.WriteString(strings.Join(header, "\r\n"))
	buf.WriteString("\r\n\r\n")

	if err := part(mw, "text/plain; charset=utf-8", msg.Text); err != nil {
		return nil, err
	}
	if err := part(mw, "text/html; charset=utf-8", msg.HTML); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("mailer: close multipart: %w", err)
	}
	return buf.Bytes(), nil
}

// part writes one quoted-printable part: it keeps UTF-8 text safe for 7-bit
// servers and wraps long HTML lines under the SMTP 998-byte limit.
func part(mw *multipart.Writer, contentType, content string) error {
	w, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {contentType},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	if err != nil {
		return fmt.Errorf("mailer: create part: %w", err)
	}
	qp := quotedprintable.NewWriter(w)
	if _, err := qp.Write([]byte(content)); err != nil {
		return fmt.Errorf("mailer: write part: %w", err)
	}
	if err := qp.Close(); err != nil {
		return fmt.Errorf("mailer: write part: %w", err)
	}
	return nil
}

func messageID(fromAddr string) string {
	domain := "localhost"
	if i := strings.LastIndexByte(fromAddr, '@'); i >= 0 {
		domain = fromAddr[i+1:]
	}
	b := make([]byte, 16)
	_, _ = rand.Read(b) // crypto/rand.Read never fails on supported platforms
	return "<" + hex.EncodeToString(b) + "@" + domain + ">"
}
