package main

import (
	"goproj/internal/config"
	"goproj/internal/handlers"
	"goproj/internal/routes"
	"goproj/internal/store"
	"log"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalln("Config error:", err)
	}

	pool, err := store.Connect()
	if err != nil {
		log.Fatalln("Database error:", err)
	}
	defer pool.Close()

	vh := handlers.ValueHandler{
		DB: pool,
	}
	uh := handlers.UserHandler{
		DB: pool,
	}
	routes.Setup(&vh, &uh)
}
