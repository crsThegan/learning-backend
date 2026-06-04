package handlers

import (
	"context"
	"fmt"
	"goproj/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ValueHandler struct {
	DB *pgxpool.Pool
}

type valueResponse struct {
	ID int `json:"id" binding:"required"`
	models.Value
}

type valueRequest struct {
	models.Value
}

func (h *ValueHandler) Get(c *gin.Context) {
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var response valueResponse
	err = h.DB.QueryRow(ctx,
		`SELECT * from vals WHERE id = $1`,
		id).Scan(&response.ID, &response.Data)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Value with id %d not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ValueHandler) GetAll(c *gin.Context) {
	q := c.Query("value_id")
	if q != "" {
		c.Redirect(http.StatusMovedPermanently, "/values/"+q)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	rows, err := h.DB.Query(ctx, "SELECT * FROM vals")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Database error: %s", err),
		})
	}
	defer rows.Close()

	response := make([]valueResponse, 0)
	i := 0
	for rows.Next() {
		response = append(response, valueResponse{})
		rows.Scan(&response[i].ID, &response[i].Data)
		i++
	}

	c.JSON(http.StatusOK, response)
}

func (h *ValueHandler) Set(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil || id < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id not a non-negative number",
		})
		return
	}

	var req valueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong request format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var response valueResponse
	err = h.DB.QueryRow(ctx,
		`UPDATE vals
		 SET data = $1
		 WHERE id = $2
		 RETURNING *`,
		req.Data, id).Scan(&response.ID, &response.Data)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("value with id %d not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"updated": response,
	})
}

func (h *ValueHandler) Append(c *gin.Context) {
	var req valueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong request data format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var response valueResponse
	err := h.DB.QueryRow(ctx,
		`INSERT INTO vals (data)
		 VALUES ($1)
		 RETURNING *`,
		req.Data).Scan(&response.ID, &response.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appended": response,
	})
}

func (h *ValueHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil || id < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id not a non-negative number",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var response valueResponse
	err = h.DB.QueryRow(ctx,
		`DELETE FROM vals
		 WHERE id = $1
		 RETURNING *`,
		id).Scan(&response.ID, &response.Data)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("value with id %d not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted": response,
	})
}
