package main

import (
	"log"
	"net/http"
	"os"

	"github.com/4rneee/noah-updater/controllers"
	"github.com/4rneee/noah-updater/middlewares"
	"github.com/4rneee/noah-updater/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	models.ConnectDatabase()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	store := cookie.NewStore([]byte(os.Getenv("SECRET")))
	store.Options(sessions.Options{
        MaxAge: 60 * 60 * 24,
        SameSite: http.SameSiteStrictMode,
    })
	r.Use(sessions.Sessions("login", store))

	r.GET("/register", controllers.RegisterHTML)
	r.POST("/register", controllers.Register)
	r.GET("/login", controllers.LoginHTML)
	r.POST("/login", controllers.Login)
	r.GET("/posts", middlewares.CheckAuth, controllers.GetPosts)
	r.POST("/post", middlewares.CheckAuth, controllers.CreatePost)

	r.Run() // automatically uses the 'PORT' env variable
}
