package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a Discord user in the Scribble system
// This model will be used to store user information, streak data, and statistics
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	DiscordID string         `gorm:"uniqueIndex" json:"discord_id"`
	Username  string         `json:"username"`
	Avatar    string         `json:"avatar"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	// TODO: Add streak, submission count, and ranking fields when needed
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}
