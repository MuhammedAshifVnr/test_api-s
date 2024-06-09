package models

import "github.com/jinzhu/gorm"

type User struct {
    gorm.Model
    Name     string `gorm:"not null"`
    Email    string `gorm:"not null;unique"`
    Password string `gorm:"not null"`
}
