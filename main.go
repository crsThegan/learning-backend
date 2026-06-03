package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"goproj/auth"
	"goproj/config"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type Value struct {
	Data int `json:"data" binding:"required"`
}

var values []Value

type HeaderError struct {
	Msg string `json:"error"`
}

func (e HeaderError) Error() string {
	return e.Msg
}

func checkHeaders(c *gin.Context) {
	ct := c.GetHeader("Content-Type")
	if ct == "" {
		err := HeaderError{"wrong request 'Content-Type' header"}
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	if !strings.HasPrefix(ct, "application/json") {
		err := HeaderError{"wrong 'Content-Type' header value"}
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	c.Next()
}

func login(c *gin.Context) {
	var user auth.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "incorrect auth arguments",
		})
		return
	}

	token, err := auth.GetToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not generate a token: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s is authorized for role %s", user.Login, user.Role),
		"jwt":     token,
	})
}

func getValue(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id not a number",
		})
		return
	}
	if id < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id must be non-negative",
		})
		return
	}
	if int(id) >= len(values) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "value not found",
		})
		return
	}

	c.JSON(http.StatusOK, values[id])
}

type ValuesResponse struct {
	ID    int   `json:"id"`
	Value Value `json:"value"`
}

func getValues(c *gin.Context) {
	q := c.Query("value_id")
	if q != "" {
		c.Redirect(http.StatusMovedPermanently, "/values/"+q)
		return
	}

	response := make([]ValuesResponse, 0)
	for i, v := range values {
		response = append(response, ValuesResponse{
			ID:    i,
			Value: v,
		})
	}
	c.JSON(http.StatusOK, response)
}

func deleteValue(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil || id < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id not a non-negative number",
		})
		return
	}
	if int(id) >= len(values) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "value not found",
		})
		return
	}
	values = slices.Delete(values, int(id), int(id+1))
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func appendValue(c *gin.Context) {
	var value Value
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong data format",
		})
		return
	}

	values = append(values, value)

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    value.Data,
	})
}

func updateValue(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil || id < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id not a non-negative number",
		})
		return
	}

	if int(id) >= len(values) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "value not found",
		})
		return
	}

	var value Value
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong data format",
		})
		return
	}
	values[id] = value
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func logout(c *gin.Context) {
	ah := c.GetHeader("Authorization")
	if !strings.HasPrefix(ah, "Bearer ") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no bearer token",
		})
		return
	}

	tokenStr := strings.TrimPrefix(ah, "Bearer ")

	if err := auth.InvalidateToken(tokenStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user successfully logged out",
	})
}

func fillValues(data ...int) {
	for _, v := range data {
		values = append(values, Value{v})
	}
}

func main() {
	if err := config.Load(); err != nil {
		log.Fatalln("Config error:", err)
	}

	values = make([]Value, 0)
	fillValues(43, 56, 23, 41, 78, 53)

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	router.POST("/login", login)
	router.POST("/logout", checkHeaders, auth.Require("any"), logout)

	admin := router.Group("", auth.Require("admin"))

	admin.GET("/values/:id", getValue)
	admin.DELETE("/values/:id", deleteValue)
	admin.PATCH("/values/:id", checkHeaders, updateValue)
	admin.GET("/values", getValues)
	admin.POST("/values", checkHeaders, appendValue)

	router.Run()
}
