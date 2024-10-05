package controllers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/4rneee/noah-updater/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
		return
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
func LoginHTML(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tmpl", gin.H{})
}

// <=============== POST /login ===============>
type LoginUserInput struct {
	UserName string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var input LoginUserInput

	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "login.tmpl", gin.H{
			"error": "Invalid request",
		})
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
		c.HTML(http.StatusBadRequest, "login.tmpl", gin.H{
			"error": "Invalid username or password",
		})
		return
	} else if err != nil {
		c.HTML(http.StatusInternalServerError, "login.tmpl", gin.H{
			"error": "Internal Server Error",
		})
		c.Error(err)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(input.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		c.HTML(http.StatusBadRequest, "login.tmpl", gin.H{
			"error": "Invalid username or password",
		})
		return
	} else if err != nil {
		c.HTML(http.StatusInternalServerError, "login.tmpl", gin.H{
			"error": "Internal Server Error",
		})
		c.Error(err)
		return
	}

	session := sessions.Default(c)

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": input.UserName,
		"exp":      time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days
	}).SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.tmpl", gin.H{
			"error": "Internal Server Error",
		})
		c.Error(err)
		return
	}

	session.Set("token", token)
	session.Save()
	c.Redirect(http.StatusFound, "/posts")
}

// <=============== GET /logout ===============>
func Logout(c *gin.Context) {
	session := sessions.Default(c)
    session.Delete("token")
    session.Save()
    c.Redirect(http.StatusTemporaryRedirect, "/login")
}

// TODO: update user fields such as display name
