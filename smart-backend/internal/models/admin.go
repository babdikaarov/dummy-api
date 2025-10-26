package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	RoleSuper   = "super"
	RoleRegular = "regular"
)

type Admin struct {
	ID           uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex:idx_username_deleted_at;not null" json:"username"`
	Password     string         `gorm:"not null" json:"-"` // Never expose password in JSON
	Role         string         `gorm:"not null" json:"role"` // "super" or "regular"
	TokenVersion int            `gorm:"default:0" json:"-"` // For token invalidation on new login
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"uniqueIndex:idx_username_deleted_at;index" json:"-"` // Soft delete support with composite unique index
}

// BeforeCreate is a GORM hook that hashes the password before saving to database
func (a *Admin) BeforeCreate(tx *gorm.DB) error {
	// Generate UUID if not set (for non-initial admins)
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	// Hash the password with bcrypt (cost 10)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies if the provided password matches the stored hash
func (a *Admin) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}
