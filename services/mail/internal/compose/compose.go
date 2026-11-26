// Package compose renders the emails this service sends. Each email has a
// plain-text and an HTML version; clients that cannot show HTML use the text.
package compose

import (
	"embed"
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"
	"time"
)

//go:embed templates
var files embed.FS

// Templates are part of the binary: a broken one fails at startup, not on
// the first task.
var (
	otpHTML = htmltemplate.Must(htmltemplate.ParseFS(files, "templates/otp.html"))
	otpText = texttemplate.Must(texttemplate.ParseFS(files, "templates/otp.txt"))
)

const otpSubject = "Your Moments verification code"

// Email is a rendered email, ready to send.
type Email struct {
	Subject string
	Text    string
	HTML    string
}

// OTP renders the OTP email. validFor is how long the code stays valid from
// now; it is shown rounded up to whole minutes.
func OTP(code string, validFor time.Duration) (Email, error) {
	data := struct {
		Subject   string
		Code      string
		ExpiresIn string
	}{otpSubject, code, minutes(validFor)}

	var html, text strings.Builder
	if err := otpHTML.Execute(&html, data); err != nil {
		return Email{}, fmt.Errorf("compose: otp html: %w", err)
	}
	if err := otpText.Execute(&text, data); err != nil {
		return Email{}, fmt.Errorf("compose: otp text: %w", err)
	}
	return Email{Subject: otpSubject, Text: text.String(), HTML: html.String()}, nil
}

// minutes rounds up: a fresh 10-minute OTP that waited 2s in the queue still
// shows "10 minutes" (the policy the user knows), and 30s left shows
// "1 minute". It can overstate the remaining time by under a minute.
func minutes(d time.Duration) string {
	m := int((d + time.Minute - 1) / time.Minute)
	if m <= 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", m)
}
