package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/4rneee/noah-updater/controllers"
	"github.com/4rneee/noah-updater/middlewares"
	"github.com/4rneee/noah-updater/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%02d.%02d.%04d", day, month, year)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	models.ConnectDatabase()

	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"formatAsDate": formatAsDate,
	})
	r.LoadHTMLGlob("templates/*")

	store := cookie.NewStore([]byte(os.Getenv("SECRET")))
	store.Options(sessions.Options{MaxAge: 60 * 60 * 24}) // expire in a day
	r.Use(sessions.Sessions("login", store))

	r.GET("/register", controllers.RegisterHTML)
	r.POST("/register", controllers.Register)
	r.GET("/login", controllers.LoginHTML)
	r.POST("/login", controllers.Login)
	r.GET("/posts", middlewares.CheckAuth, controllers.GetPosts)
	r.POST("/post", middlewares.CheckAuth, controllers.CreatePost)

	r.Run() // automatically uses the 'PORT' env variable
}
