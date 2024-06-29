package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenVerificationRequest struct {
	Token string `json:"token"`
}

type TokenVerificationResponse struct {
	Valid  bool   `json:"valid"`
	UserID int    `json:"user_id,omitempty"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
}

func (t *User) BeforeCreate(scope *gorm.Scope) error {
	t.CreatedAt = time.Now()
	return nil
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&User{})
}
