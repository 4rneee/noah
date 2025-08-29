package controllers

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/4rneee/noah/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var invalidArgument = errors.New("invalid argument")

// <=============== GET /posts ===============>
func GetPosts(c *gin.Context) {
	const PAGE_SIZE = 5

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.Redirect(http.StatusFound, "/posts")
		return
	}

	var count int64 = 0
	err = models.DB.
		Table("posts").
		Count(&count).
		Error
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		c.Error(err)
		return
	}

	// count / PAGE_SIZE rounded up
	last_page := (int(count) + (PAGE_SIZE - 1)) / PAGE_SIZE

	// if there are no posts, correct last_page
	if last_page == 0 {
		last_page = 1
	}

	if page > last_page {
		c.Redirect(http.StatusFound, fmt.Sprintf("/posts?page=%v", last_page))
		return
	}

	var posts []models.Post
	err = models.DB.
		Order("created_at desc").
		Offset(PAGE_SIZE * (page - 1)).
		Limit(PAGE_SIZE).
		Preload("Comments").
		Find(&posts).
		Error
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		c.Error(err)
		return
	}


	user, ok := get_current_user(c)
	if !ok {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	redaction_date := get_redaction_date(&user)

	for i := range posts {
		if redaction_date.Compare(posts[i].CreatedAt) < 0 {
			posts[i].Redact()
		}
	}

	next_page := strconv.Itoa(page + 1)
	if page+1 > last_page {
		next_page = ""
	}

	prev_page := strconv.Itoa(page - 1)
	if page-1 < 1 {
		prev_page = ""
	}

	c.HTML(http.StatusOK, "posts.tmpl", gin.H{
		"posts":     posts,
		"prev_page": prev_page,
		"next_page": next_page,
	})
}

// <=============== GET /create ===============>
func CreateHTML(c *gin.Context) {
	c.HTML(http.StatusOK, "create.tmpl", gin.H{})
}

// <=============== POST /create ===============>
type PostInput struct {
	Title   string `form:"title" binding:"required"`
	Content string `form:"content" binding:"required"`
	VideoLink string `form:"video_link"`
}

func CreatePost(c *gin.Context) {
	var input PostInput

	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "create.tmpl", gin.H{
			"error": "Invalid request",
			"title": input.Title,
			"content": input.Content,
			"video_link": input.VideoLink,
		})
		c.Error(err)
		return
	}

	user, ok := get_current_user(c)
	if !ok {
		c.HTML(http.StatusInternalServerError, "create.tmpl", gin.H{
			"error": "Internal Server Error",
			"title": input.Title,
			"content": input.Content,
			"video_link": input.VideoLink,
		})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.HTML(http.StatusBadRequest, "create.tmpl", gin.H{
			"error": "Invalid request",
			"title": input.Title,
			"content": input.Content,
			"video_link": input.VideoLink,
		})
		c.Error(err)
		return
	}

	files := form.File["images"]
	images, err := store_files(c, files)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "create.tmpl", gin.H{
			"error": "Internal Server Error",
			"title": input.Title,
			"content": input.Content,
			"video_link": input.VideoLink,
		})
		c.Error(err)
		return
	}


	embed_link := strings.TrimSpace(input.VideoLink)
	if embed_link != "" {
		err, embed_link = getYouTubeEmbedLink(input.VideoLink)
		if err != nil {
			c.HTML(http.StatusBadRequest, "create.tmpl", gin.H{
				"error": err.Error(),
				"title": input.Title,
				"content": input.Content,
				"video_link": input.VideoLink,
			})
			c.Error(err)
			return
		}
	}

	post := models.Post{
		UserName: user.Name,
		Title:    input.Title,
		Content:  input.Content,
		Images:   images,
		EmbedVideo: embed_link,
	}

	err = models.DB.
		Create(&post).
		Error

	if err != nil {
		c.HTML(http.StatusInternalServerError, "create.tmpl", gin.H{
			"error": "Internal Server Error",
			"title": input.Title,
			"content": input.Content,
			"video_link": input.VideoLink,
		})
		c.Error(err)
		return
	}

	c.Redirect(http.StatusFound, "/posts")
}

// <=============== GET /uploads/:filename ===============>
func Uploads(c *gin.Context) {
	file_name := filepath.Clean(c.Param("filename"))
	path := filepath.Join("uploads", file_name)
	c.File(path)
}

// <=============== GET /post/:id ===============>
func GetPost(c *gin.Context) {
	var post models.Post
	err := get_post_with_comments(&post, c.Param("id"))

	if err == invalidArgument {
		c.String(http.StatusBadRequest, "Invalid id")
		return
	} else if err == gorm.ErrRecordNotFound {
		c.String(http.StatusNotFound, "Post not found")
		return
	} else if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		c.Error(err)
		return
	}

	user, ok := get_current_user(c)
	if !ok {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		c.Error(err)
		return
	}
	redaction_date := get_redaction_date(&user)
	if redaction_date.Compare(post.CreatedAt) < 0 {
		c.Redirect(http.StatusFound, "/create")
		return
	}


	c.HTML(http.StatusOK, "post.tmpl", gin.H{
		"post": post,
	})
}

// <=============== POST /post/id ===============>
func PostComment(c *gin.Context) {
	var post models.Post
	err := get_post_with_comments(&post, c.Param("id"))

	if err == invalidArgument {
		c.String(http.StatusBadRequest, "Invalid id")
		return
	} else if err == gorm.ErrRecordNotFound {
		c.String(http.StatusNotFound, "Post not found")
		return
	} else if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		c.Error(err)
		return
	}

	user, ok := get_current_user(c)
	if !ok {
		c.HTML(http.StatusInternalServerError, "post.tmpl", gin.H{
			"post":  post,
			"error": "Internal Server Error",
		})
		return
	}

	content := c.PostForm("content")

	form, err := c.MultipartForm()
	if err != nil {
		c.HTML(http.StatusBadRequest, "post.tmpl", gin.H{
			"post":  post,
			"error": "Invalid request",
		})
		c.Error(err)
		return
	}

	files := form.File["images"]
	if len(content) == 0 && len(files) == 0 {
		c.HTML(http.StatusBadRequest, "post.tmpl", gin.H{
			"post":  post,
			"error": "A comment requires text or at least one image",
		})
		return
	}

	images, err := store_files(c, files)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "create.tmpl", gin.H{
			"error": "Internal Server Error",
		})
		c.Error(err)
		return
	}

	comment := models.Comment{
		PostID:   post.ID,
		UserName: user.Name,
		Content:  content,
		Images:   images,
	}

	err = models.DB.
		Model(&post).
		Association("Comments").
		Append(&comment)

	if err != nil {
		c.HTML(http.StatusInternalServerError, "post.tmpl", gin.H{
			"post":  post,
			"error": "Internal Server Error",
		})
		c.Error(err)
		return
	}

	c.HTML(http.StatusOK, "post.tmpl", gin.H{
		"post": post,
	})
}

func get_current_user(c *gin.Context) (models.User, bool) {
	var user models.User
	// the current_user variable should exist and be of type models.User
	cur_user, exists := c.Get("current_user")
	if !exists {
		return user, false
	}

	user, ok := cur_user.(models.User)
	if !ok {
		return user, false
	}

	return user, true
}

func store_files(c *gin.Context, files []*multipart.FileHeader) ([]string, error) {
	var err error = nil
	file_names := make([]string, len(files))

	for idx, file := range files {
		// generate file name from current time
		// (don't use original filename because we don't want to overwrite existing files)
		name := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))

		err = c.SaveUploadedFile(file, filepath.Join("uploads", name))
		if err != nil {
			break
		}
		file_names[idx] = name
	}

	return file_names, err
}

func get_post_with_comments(post *models.Post, str_id string) error {
	id, err := strconv.Atoi(str_id)
	if err != nil {
		return err
	}

	if id < 0 {
		return invalidArgument
	}

	return models.DB.
		Table("posts").
		Where("id = ?", uint(id)).
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at asc")
		}).
		First(post).
		Error
}


func getYouTubeEmbedLink(rawURL string) (error, string) {
	// regex to match youtube links in the following forms:
	// https://www.youtube.com/watch?v=<ID>
	// https://youtu.be/<ID>?feature=shared
	// https://www.youtube.com/embed/<ID>
	// the https:// and www. are optional and additional query parameters are ignored
    re, err := regexp.Compile(`(?:https?://)?(?:(?:www\.)?youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})(?:&.*)?|(?:(?:youtu\.be|(?:www\.)?youtube\.com/embed)/([a-zA-Z0-9_-]{11})(?:\?.*)?)(?:&.*)?)`)
	if err != nil {
		return fmt.Errorf("Internal server error"), ""
	}
    match := re.FindStringSubmatch(rawURL)
    if len(match) != 3 {
        return fmt.Errorf("URL is not a valid YouTube link"), ""
    }

	video_id := match[1]
	if video_id == "" {
		video_id = match[2]
	}
	if video_id == "" {
        return fmt.Errorf("URL is not a valid YouTube link"), ""
	}

    return nil, fmt.Sprintf("https://www.youtube.com/embed/%s", video_id)
}

// Returns the date after which posts should get redacted i.e. the latest post + VISIBILITY_WINDOW
// If no post exists it returns time.Time{} i.e. January 1, year 1, 00:00
func get_redaction_date(user *models.User) time.Time {
	window, err := strconv.Atoi(os.Getenv("VISIBILITY_WINDOW"))
	if err != nil || window < 0 {
		// negative or non-numeric window values make all posts always visible
		return time.Now()
	}

	var latest_post models.Post
	err = models.DB.
		Table("posts").
		Where("user_name = ?", user.Name).
		Order("created_at desc").
		First(&latest_post).
		Error
	if err != nil {
		return time.Time{}
	}

	return latest_post.CreatedAt.Add(time.Duration(window) * time.Hour * 24)
}
