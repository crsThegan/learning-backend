package routes

import (
	"github.com/gin-gonic/gin"
	"goproj/internal/auth"
	"goproj/internal/handlers"
	"net/http"
)

func Setup(vh *handlers.ValueHandler, uh *handlers.UserHandler) {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	router.POST("/login", uh.Login)
	router.POST("/logout", auth.Require("any"), uh.Logout)

	admin := router.Group("", auth.Require("admin"))

	admin.GET("/values/:id", vh.Get)
	admin.DELETE("/values/:id", vh.Delete)
	admin.PATCH("/values/:id", vh.Set)
	admin.GET("/values", vh.GetAll)
	admin.POST("/values", vh.Append)

	router.Run()
}
