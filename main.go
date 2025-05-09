package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/4rneee/noah/controllers"
	"github.com/4rneee/noah/middlewares"
	"github.com/4rneee/noah/models"
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
	r.StaticFile("navbar.css", "./static/navbar.css")
	r.LoadHTMLGlob("templates/*")

	store := cookie.NewStore([]byte(os.Getenv("SECRET")))
	store.Options(sessions.Options{
		MaxAge:   60 * 60 * 24 * 30, // 30 days
		SameSite: http.SameSiteStrictMode,
	})
	r.Use(sessions.Sessions("login", store))

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/posts")
	})
	r.GET("/register", controllers.RegisterHTML)
	r.POST("/register", controllers.Register)
	r.GET("/login", controllers.LoginHTML)
	r.POST("/login", controllers.Login)
	r.GET("/logout", controllers.Logout)
	r.GET("/posts", middlewares.CheckAuth, controllers.GetPosts)
	r.GET("/create", middlewares.CheckAuth, controllers.CreateHTML)
	r.POST("/create", middlewares.CheckAuth, controllers.CreatePost)
	r.GET("/uploads/:filename", middlewares.CheckAuth, controllers.Uploads)
	r.GET("/post/:id", middlewares.CheckAuth, controllers.GetPost)
	r.POST("/post/:id", middlewares.CheckAuth, controllers.PostComment)

	r.Run() // automatically uses the 'PORT' env variable
}
