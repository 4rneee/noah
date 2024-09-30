package models

import (
	"gorm.io/datatypes"
	"time"
)

type Post struct {
	ID        uint                        `json:"-" gorm:"primaryKey"`
	CreatedAt time.Time                   `json:"created_at"`
	UpdatedAt time.Time                   `json:"updated_at"`
	UserName  string                      `json:"user_name"`
	Title     string                      `json:"title"`
	Content   string                      `json:"content"`
	Images    datatypes.JSONSlice[string] `json:"images"`
}
