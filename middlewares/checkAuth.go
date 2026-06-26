package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/4rneee/noah/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CheckAuth(c *gin.Context) {
	session := sessions.Default(c)
	t := session.Get("token")

	if t == nil {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}

	token_str, ok := t.(string)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}

	token, err := jwt.Parse(token_str, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.Redirect(http.StatusFound, "/login")
		c.Error(err)
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}

	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}

	var user models.User
	err = models.DB.
		Table("users").
		Where("name = ?", claims["username"]).
		First(&user).
		Error
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}

	c.Set("current_user", user)

	c.Next()
}

func BasicAuth(c *gin.Context) {
	username, password, ok := c.Request.BasicAuth()

	if !ok {
		c.Header("WWW-Authenticate",  "Basic realm=\"Authorization Required\"")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var user models.User
	err := models.DB.
		Table("users").
		Where("name = ?", username).
		First(&user).
		Error

	if err == gorm.ErrRecordNotFound {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Set("current_user", user)

	c.Next()
}
