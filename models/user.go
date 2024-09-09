package models

import "time"

type User struct {
	Name        string    `json:"user_name" gorm:"constraint:OnUpdate:CASCADE,primaryKey"`
	DisplayName string    `json:"display_name"`
	Password    []byte    `json:"-" gorm:"type:BINARY(60)"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
    Posts       []Post    `json:"posts" gorm:"foreignKey:UserName;references:Name"`
}
