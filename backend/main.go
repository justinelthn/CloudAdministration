package main

import (
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
=======
    log.Println("Server running on :8081")
    r.Run(":8081")
>>>>>>> 828618a7bbb67b386c74c4f6ba8dc2b56d6f99ca
}
