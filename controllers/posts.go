package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/4rneee/noah-updater/models"
	"github.com/gin-gonic/gin"
)

func usernameFromToken(token string) (string, bool) {
	idx := strings.LastIndex(token, "_")
	if idx == -1 {
		return "", false
	}
	if token[idx+1:] != "token" {
		return "", false
	}
	return token[:idx], true
}

// <=============== GET /posts ===============>
type GetPostsInput struct {
	Token    string `json:"token" binding:"required"`
	Page     uint   `json:"page"`
	PageSize uint   `json:"page_size"`
}

func GetPosts(c *gin.Context) {
	var input GetPostsInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Error(err)
		return
	}

	// TODO: validate token
	_, valid_token := usernameFromToken(input.Token)

	if !valid_token {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	if input.Page == 0 {
		input.Page = 1
	}
	if input.PageSize == 0 {
		input.PageSize = 25
	}

	var posts []models.Post
	err := models.DB.
		Order("created_at desc").
		Offset(int(input.PageSize * (input.Page - 1))).
		Limit(int(input.PageSize)).
		Find(&posts).
		Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// <=============== POST /post ===============>
type PostInput struct {
	Token   string `json:"token" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func CreatePost(c *gin.Context) {
	var input PostInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Error(err)
		return
	}

	user_name, valid_token := usernameFromToken(input.Token)
	if !valid_token {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	post := models.Post{
		ID:        0,           // will be set by the DB
		CreatedAt: time.Time{}, // will be set by the DB
		UpdatedAt: time.Time{}, // will be set by the DB
		UserName:  user_name,
		Content:   input.Content,
	}

	err := models.DB.
		Create(&post).
		Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"post": post})
}
