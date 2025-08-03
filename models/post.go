package models

import (
	"gorm.io/datatypes"
	"time"
)

type Post struct {
	ID        uint                        `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time                   `json:"created_at"`
	UpdatedAt time.Time                   `json:"updated_at"`
	UserName  string                      `json:"user_name" gorm:"index"`
	User      User                        `json:"-" gorm:"foreignKey:UserName;references:Name;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Title     string                      `json:"title"`
	Content   string                      `json:"content"`
	Images    datatypes.JSONSlice[string] `json:"images"`
	Comments  []Comment                   `json:"comments" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	EmbedVideo string                     `json:"embed_video"`
}
