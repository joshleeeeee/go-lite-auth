package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email     string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"` // never expose password
	Nickname  string         `gorm:"size:50" json:"nickname"`
	Avatar    string         `gorm:"size:255" json:"avatar"`
	Status    int            `gorm:"default:1" json:"status"` // 1: active, 0: disabled
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// Client represents an OAuth client application
type Client struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	ClientID     string         `gorm:"uniqueIndex;size:100;not null" json:"client_id"`
	ClientSecret string         `gorm:"size:255;not null" json:"-"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	RedirectURI  string         `gorm:"size:500;not null" json:"redirect_uri"`
	Description  string         `gorm:"size:500" json:"description"`
	Status       int            `gorm:"default:1" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Client) TableName() string {
	return "clients"
}
