package main

import (
	"github.com/4rneee/noah-updater/controllers"
	"github.com/4rneee/noah-updater/models"

	"github.com/gin-gonic/gin"
)

func main() {
	models.ConnectDatabase()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/register", controllers.RegisterHTML)
	r.POST("/register", controllers.Register)
	r.GET("/login", controllers.Login)
	r.GET("/posts", controllers.GetPosts)
	r.POST("/post", controllers.CreatePost)

	r.Run(":8080")
}
