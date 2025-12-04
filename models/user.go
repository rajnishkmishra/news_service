package models

import "time"

var ActivityType = map[string]string{}

type User struct {
	ID          uint64 `json:"id" gorm:"primary_key;not null;auto_increment"`
	Name        string `json:"name" gorm:"not null"`
	PhoneNumber string `json:"phone_number" gorm:"not null"`
	EmailID     string `json:"email_id"`
}

type UserActivity struct {
	ID           uint64    `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	UserID       uint64    `json:"user_id" gorm:"not null"`
	User         *User     `json:"user" gorm:"foreignKey:UserID;references:ID"`
	NewsID       uint64    `json:"news_id" gorm:"not null"`
	News         *News     `json:"news" gorm:"foreignKey:NewsID;references:ID"`
	ActivityType string    `json:"activity_type" gorm:"type:varchar(255);not null"`
	ActivityTime time.Time `json:"activity_time" gorm:"autoCreateTime"`
}
