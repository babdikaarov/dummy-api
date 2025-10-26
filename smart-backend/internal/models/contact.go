package models

import "time"

// Contact represents the application's contact information
// There should be only one record in this table
type Contact struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SupportNumber  int       `gorm:"not null" json:"support_number"`
	EmailSupport   string    `gorm:"not null" json:"email_support"`
	Address        string    `gorm:"not null" json:"address"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName specifies the table name for the Contact model
func (Contact) TableName() string {
	return "contacts"
}
