package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/4rneee/noah-updater/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func sanetizeUserName(name string) (string, bool) {
	name = strings.TrimSpace(name)
	// TODO: add more checks, maybe only allow letters, numbers, '_' and '-'
	if name == "" {
		return name, false
	}
	return name, true
}

func validPassword(password string) bool {
	return len(password) > 0 && len(password) <= 72
}

// <=============== GET /register ===============>
func RegisterHTML(c *gin.Context) {
	c.HTML(http.StatusOK, "register.tmpl", gin.H{})
}

// <=============== POST /register ===============>
type RegisterUserInput struct {
	UserName       string `form:"username" binding:"required"`
	Password       string `form:"password" binding:"required"`
	GlobalPassword string `form:"global_password" binding:"required"`
}

func Register(c *gin.Context) {
	var input RegisterUserInput

	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "register.tmpl", gin.H{
			"error": "Invalid request",
		})
		c.Error(err)
		return
	}

    if input.GlobalPassword != os.Getenv("GLOBAL_PASSWORD") {
        c.HTML(http.StatusBadRequest, "register.tmpl", gin.H{
            "error": "Invalid global password",
        })
    }

	var count int64
	models.DB.
		Table("users").
		Where("name = ?", input.UserName).
		Count(&count)
	if count > 0 {
		c.HTML(http.StatusBadRequest, "register.tmpl", gin.H{
			"error": "The username has already been taken.",
		})
		return
	}

	sanetized_name, valid_name := sanetizeUserName(input.UserName)
	if !valid_name {
		c.HTML(http.StatusBadRequest, "register.tmpl", gin.H{
			"error": "Invalid username",
		})
		return
	}

	if !validPassword(input.Password) {
		c.HTML(http.StatusBadRequest, "register.tmpl", gin.H{
			"error": "Invalid password",
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "register.tmpl", gin.H{
			"error": "Internal Server Error",
		})
		c.Error(err)
		return
	}

	insert := models.User{
		Name:        sanetized_name,
		DisplayName: sanetized_name, // default name, can be updated later
		Password:    hash,
		CreatedAt:   time.Time{}, // will be set by the DB
		UpdatedAt:   time.Time{}, // will be set by the DB
		Posts:       []models.Post{},
	}

	err = models.DB.
		Create(&insert).
		Error

	if err != nil {
		c.HTML(http.StatusInternalServerError, "register.tmpl", gin.H{
			"error": "Internal Server Error",
		})
		c.Error(err)
		return
	}

	c.Redirect(http.StatusFound, "/login")
}

// <=============== GET /login ===============>
type LoginUserInput struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var input LoginUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Error(err)
		return
	}

	var user models.User
	err := models.DB.
		Table("users").
		Where("name = ?", input.UserName).
		First(&user).
		Error

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		c.Error(err)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(input.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		c.Error(err)
		return
	}

	// TODO: generate and store token
	token := fmt.Sprintf("%v_token", user.Name)
	c.JSON(http.StatusOK, gin.H{"token": token})
	c.Redirect(http.StatusFound, "/posts")
}

// TODO: update user fields such as display name
