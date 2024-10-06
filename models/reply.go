package models

import (
	"gorm.io/datatypes"
	"time"
)

type Comment struct {
	ID        uint                        `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time                   `json:"created_at"`
	UpdatedAt time.Time                   `json:"updated_at"`
	PostID    uint                        `json:"post_id"`
	UserName  string                      `json:"user_name"`
	User      User                        `json:"-" gorm:"foreignKey:UserName;references:Name;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Content   string                      `json:"content"`
	Images    datatypes.JSONSlice[string] `json:"images"`
}
