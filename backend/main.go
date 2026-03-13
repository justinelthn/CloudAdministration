package main

import (
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
	"movie-api/handlers"
=======
    "log"
>>>>>>> 828618a7bbb67b386c74c4f6ba8dc2b56d6f99ca

    "github.com/gin-gonic/gin"

    "movie-api/handlers" // <<== ici on met ton module + dossier
)

func main() {
    r := gin.Default()

    r.GET("/movies", handlers.GetMovies)
    r.GET("/suggest", handlers.Suggest)

<<<<<<< HEAD
	r.GET("/movies", handlers.GetMovies)

	r.Run(":8082")
=======
    "log"
=======
	"log"
	"time"
>>>>>>> 4ae80e220b143f9feb209044a235bde9189a26e7

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

<<<<<<< HEAD
    log.Println("Server running on :8081")
    r.Run(":8081")
>>>>>>> 51699d5 (ajout frontend + changes)
=======
    log.Println("Server running on :8081")
    r.Run(":8081")
>>>>>>> 828618a7bbb67b386c74c4f6ba8dc2b56d6f99ca
=======
	r.GET("/movies", handlers.GetMovies)
	r.GET("/suggest", handlers.Suggest)

	// Serve frontend static files
	r.StaticFile("/", "../frontend/index.html")
	r.StaticFile("/style.css", "../frontend/style.css")
	r.StaticFile("/script.js", "../frontend/script.js")

	log.Println("Server running on :8081")
	r.Run(":8081")
>>>>>>> 4ae80e220b143f9feb209044a235bde9189a26e7
}
