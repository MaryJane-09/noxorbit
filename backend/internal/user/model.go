package user

import "github.com/gofrs/uuid"

type User struct {
	Id       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}

type UserResponse struct {
	ID    *uuid.UUID `json:"id,omitempty"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
