package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	Balance      int
	Username     string
	PasswordHash string
	UUID         uuid.UUID
	CreatedAt    time.Time
}
