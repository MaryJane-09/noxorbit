package validate

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/MaryJane-09/noxorbit/backend/internal/user"
)

// ValidateEmail rejects address lists and display names; callers normalize
// case and whitespace before use.
func ValidateEmail(email string) error {
	if email == "" || len(email) > 254 || strings.TrimSpace(email) != email {
		return errors.New("Invalid email input")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || addr.Name != "" {
		return errors.New("Invalid email input")
	}
	return nil
}

func ValidateRegister(info user.User) error {
	name := strings.TrimSpace(info.Name)
	if n := utf8.RuneCountInString(name); n < 1 || n > 100 {
		return errors.New("Name must contain 1 to 100 characters")
	}
	for _, char := range name {
		if !(unicode.IsLetter(char) || unicode.IsMark(char) || unicode.IsSpace(char) ||
			char == '-' || char == '\'' || char == '’') {
			return errors.New("Name can contain letters, spaces, hyphens and apostrophes only")
		}
	}
	if err := ValidateEmail(info.Email); err != nil {
		return err
	}
	return ValidatePassword(info.Password)
}
