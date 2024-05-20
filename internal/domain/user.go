package domain

import "time"

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}

type Token struct {
	ID         uint      `gorm:"primaryKey" json:"-"`
	UserID     uint      `gorm:"not null" json:"-"`
	Token      string    `gorm:"not null" json:"token"`
	Expiration time.Time `gorm:"not null" json:"-"`
}
