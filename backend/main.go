package main

import (
<<<<<<< HEAD
	"movie-api/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/movies", handlers.GetMovies)

	r.Run(":8082")
=======
    "log"

    "github.com/gin-gonic/gin"

    "movie-api/handlers" // <<== ici on met ton module + dossier
)

func main() {
    r := gin.Default()

    r.GET("/movies", handlers.GetMovies)
    r.GET("/suggest", handlers.Suggest)

    log.Println("Server running on :8081")
    r.Run(":8081")
>>>>>>> 51699d5 (ajout frontend + changes)
}
