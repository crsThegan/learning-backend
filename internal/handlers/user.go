package handlers

import (
	"context"
	"fmt"
	"goproj/internal/auth"
	"goproj/internal/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserHandler struct {
	DB *pgxpool.Pool
}

type userResponse struct {
	ID int `json:"id" binding:"required"`
	models.User
}

type userRequest struct {
	models.User
}

func (h *UserHandler) Login(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "incorrect auth arguments",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var response userResponse
	err := h.DB.QueryRow(ctx,
		`SELECT id, name, role
		 FROM users
		 WHERE name = $1
		 AND role = $2`,
		req.Login, req.Role).Scan(&response.ID, &response.Login, &response.Role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("User %s with role %s not found", req.Login, req.Role),
		})
		return
	}

	token, err := auth.GetToken(req.User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("could not generate a token for user %s: %s",
				req.Login, err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s is authorized for role %s", req.Login, req.Role),
		"jwt":     token,
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
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
