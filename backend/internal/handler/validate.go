package handler

import (
	"net/mail"
	"strings"
)

func validateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil && strings.Contains(email, "@")
}
