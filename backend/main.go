package main

import (
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
}
