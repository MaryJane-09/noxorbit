package validate

import "errors"

func ValidatePassword(password string) error {
	if len(password) == 0 {
		return errors.New("Password cannot be empty")
	}
	if len(password) < 8 {
		return errors.New("Password is too short")
	}
	if len(password) > 72 {
		return errors.New("Password is too long")
	}
	return nil
}