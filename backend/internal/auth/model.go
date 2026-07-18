package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	Name          string
	Email         string
	PasswordHash  string
	AvatarURL     *string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
