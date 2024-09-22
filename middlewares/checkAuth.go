package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/4rneee/noah-updater/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func CheckAuth(c *gin.Context) {
	session := sessions.Default(c)
	t := session.Get("token")

	if t == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	token_str, ok := t.(string)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
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
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		c.Redirect(http.StatusFound, "/login")
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
		return
	}

	c.Set("current_user", user)

	c.Next()
}
