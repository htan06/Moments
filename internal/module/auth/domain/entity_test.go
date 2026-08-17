package domain

import (
	"errors"
	"testing"

	"github.com/htan06/Moments/internal/errs"
)

func TestNewUserSuccess(t *testing.T) {
	phoneNumbers := []string{
		"",
		"0123456789",
		"+12123456789",
	}

	tests := []struct {
		email        string
		phoneNumber  *string
		passwordHash string
	}{
		{
			email:        "email1@example.com",
			phoneNumber:  nil,
			passwordHash: "pwdhash",
		},
		{
			email:        "email2@example.com",
			phoneNumber:  &phoneNumbers[0],
			passwordHash: "pwdhash",
		},
		{
			email:        "email2@example.com",
			phoneNumber:  &phoneNumbers[1],
			passwordHash: "pwdhash",
		},
		{
			email:        "email2@example.com",
			phoneNumber:  &phoneNumbers[2],
			passwordHash: "pwdhash",
		},
	}

	for i, d := range tests {
		user, err := NewUser(
			d.email,
			d.phoneNumber,
			d.passwordHash,
		)
		if user == nil || err != nil {
			t.Errorf("tests[%d]: %s", i, err.Error())
		}
	}
}

func TestNewUserFailureByEmailInvalid(t *testing.T) {
	phoneNumber := "0123456789"
	user, err := NewUser(
		"emailemailemail",
		&phoneNumber,
		"aaa",
	)

	if user != nil {
		t.Fatalf("expected user nil")
	}

	if err == nil {
		t.Fatalf("expected err")
	}

	domainErr, ok := errors.AsType[*errs.Error](err)
	if !ok {
		t.Fatalf("expected domain error")
	}
	if domainErr.Code != EmailInvalid {
		t.Fatalf("expected EmailInvalid error")
	}
}

func TestNewUserFailureByPhoneNumberInvalid(t *testing.T) {
	phoneNumbers := []string{"a123456789", ".981234a6789"}
	tests := []struct {
		email        string
		username     string
		name         string
		phoneNumber  *string
		passwordHash string
	}{
		{
			email:        "email1@example.com",
			username:     "username1",
			name:         "aa",
			phoneNumber:  &phoneNumbers[0],
			passwordHash: "pwdhash",
		},
		{
			email:        "email2@example.com",
			username:     "username2",
			name:         "aaaaaaaaaaaa",
			phoneNumber:  &phoneNumbers[1],
			passwordHash: "pwdhash",
		},
	}

	for i, d := range tests {
		user, err := NewUser(
			d.email,
			d.phoneNumber,
			d.passwordHash,
		)

		if user != nil {
			t.Fatalf("test[%d]: expected user nil", i)
		}

		if err == nil {
			t.Fatalf("test[%d]: expected err", i)
		}

		domainErr, ok := errors.AsType[*errs.Error](err)
		if !ok {
			t.Fatalf("test[%d]: expected domain error", i)
		}
		if domainErr.Code != PhoneNumberInvalid {
			t.Fatalf("test[%d]: expected PhoneNumberInvalid error", i)
		}
	}

}

func TestNewUserFailureByPasswordHashEmpty(t *testing.T) {
	phoneNumber := "0123456789"
	user, err := NewUser(
		"email@email.email",
		&phoneNumber,
		"",
	)

	if user != nil {
		t.Fatalf("expected user nil")
	}

	if err == nil {
		t.Fatalf("expected err")
	}

	domainErr, ok := errors.AsType[*errs.Error](err)
	if !ok {
		t.Fatalf("expected domain error")
	}
	if domainErr.Code != PasswordHashEmpty {
		t.Fatalf("expected PasswordHashEmpty error")
	}
}
