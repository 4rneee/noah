package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/4rneee/noah-updater/models"
	"github.com/gin-gonic/gin"
)

// <=============== GET /posts ===============>
type GetPostsInput struct {
	Token    string `json:"token" binding:"required"`
	Page     uint   `json:"page"`
	PageSize uint   `json:"page_size"`
}

func GetPosts(c *gin.Context) {
	var input GetPostsInput

	// TODO: maybe get page and page size from url params

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

	c.HTML(http.StatusOK, "posts.tmpl", gin.H{
		"posts": posts,
	})

	return
}

// <=============== GET /create ===============>
func CreateHTML(c *gin.Context) {
	c.HTML(http.StatusOK, "create.tmpl", gin.H{})
}

// <=============== POST /create ===============>
type PostInput struct {
	Title   string `form:"title" binding:"required"`
	Content string `form:"content" binding:"required"`
}

func CreatePost(c *gin.Context) {
	var input PostInput

	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "create.tmpl", gin.H{
			"error": "Invalid request",
		})
		c.Error(err)
		return
	}

	// the current_user variable should exist and be of type models.User
	cur_user, exists := c.Get("current_user")
	if !exists {
		c.HTML(http.StatusInternalServerError, "create.tmpl", gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	user, ok := cur_user.(models.User)
	if !ok {
		c.HTML(http.StatusInternalServerError, "create.tmpl", gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.HTML(http.StatusBadRequest, "create.tmpl", gin.H{
			"error": "Invalid request",
		})
		c.Error(err)
		return
	}

	files := form.File["images"]

	images := make([]string, len(files))
	for idx, file := range files {
		// generate file name from current time
		// (don't use original filename because we don't want to overwrite existing files)
		name := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))

		err = c.SaveUploadedFile(file, filepath.Join("uploads", name))
		if err != nil {
			c.HTML(http.StatusInternalServerError, "create.tmpl", gin.H{
				"error": "Internal Server Error",
			})
			c.Error(err)
			return
		}
        images[idx] = name
	}

	post := models.Post{
		UserName: user.Name,
		Title:    input.Title,
		Content:  input.Content,
		Images:   images,
	}

	err = models.DB.
		Create(&post).
		Error

	if err != nil {
		c.HTML(http.StatusInternalServerError, "create.tmpl", gin.H{
			"error": "Internal Server Error",
		})
		c.Error(err)
		return
	}

	c.Redirect(http.StatusFound, "/posts")
}
