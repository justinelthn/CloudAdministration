package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"movie-api/handlers"
)

func main() {
	r := gin.Default()

	// CORS — allow frontend from any origin during development
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/movies", handlers.GetMovies)
	r.GET("/suggest", handlers.Suggest)

	// Serve frontend static files
	r.StaticFile("/", "../frontend/index.html")
	r.StaticFile("/style.css", "../frontend/style.css")
	r.StaticFile("/script.js", "../frontend/script.js")

	log.Println("Server running on :8081")
	r.Run(":8081")
}
