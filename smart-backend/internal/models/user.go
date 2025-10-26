package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID              uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Phone           string         `gorm:"uniqueIndex:idx_phone_deleted_at;not null" json:"phone"`
	Password        string         `gorm:"not null" json:"-"` // Never expose password in JSON
	TokenVersion    int            `gorm:"default:0;not null" json:"-"` // Token version for invalidation
	CurrentDeviceID string         `gorm:"type:varchar(255);default:''" json:"-"` // Track current device for device-based token invalidation
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"uniqueIndex:idx_phone_deleted_at;index" json:"-"` // Soft delete support with composite unique index
}

// BeforeCreate is a GORM hook that hashes the password and generates UUID before saving to database
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Generate UUID if not set
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	// Hash the password with bcrypt (cost 10)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies if the provided password matches the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
