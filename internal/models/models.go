package models

type Value struct {
	Data int `json:"data" binding:"required"`
}

type User struct {
	Login string `json:"name" binding:"required"`
	Role  string `json:"role" binding:"required"`
}
