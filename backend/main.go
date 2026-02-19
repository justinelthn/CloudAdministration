package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Connexion à PostgreSQL
	connStr := "user=justineletheno dbname=movies_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.GET("/movies", func(c *gin.Context) {

		// Paramètres de recherche
		actor := c.Query("actor")
		director := c.Query("director")
		genre := c.Query("genre")
		language := c.Query("language")
		minRatingStr := c.Query("min_rating")
		limitStr := c.DefaultQuery("limit", "20")
		offsetStr := c.DefaultQuery("offset", "0")

		// Conversion des nombres
		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)
		minRating := 0.0
		if minRatingStr != "" {
			minRating, _ = strconv.ParseFloat(minRatingStr, 64)
		}

		// Construire la requête dynamiquement
		query := "SELECT film_title, director, actors, genres, original_language, average_rating, runtime FROM movies_clean WHERE 1=1"
		args := []interface{}{}
		argIdx := 1

		if actor != "" {
			query += " AND actors ILIKE '%' || $" + strconv.Itoa(argIdx) + " || '%'"
			args = append(args, actor)
			argIdx++
		}
		if director != "" {
			query += " AND director ILIKE '%' || $" + strconv.Itoa(argIdx) + " || '%'"
			args = append(args, director)
			argIdx++
		}
		if genre != "" {
			query += " AND genres ILIKE '%' || $" + strconv.Itoa(argIdx) + " || '%'"
			args = append(args, genre)
			argIdx++
		}
		if language != "" {
			query += " AND original_language ILIKE '%' || $" + strconv.Itoa(argIdx) + " || '%'"
			args = append(args, language)
			argIdx++
		}
		if minRatingStr != "" {
			query += " AND average_rating >= $" + strconv.Itoa(argIdx)
			args = append(args, minRating)
			argIdx++
		}

		query += " ORDER BY average_rating DESC LIMIT $" + strconv.Itoa(argIdx)
		args = append(args, limit)
		argIdx++
		query += " OFFSET $" + strconv.Itoa(argIdx)
		args = append(args, offset)

		// Exécuter la requête
		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var movies []map[string]interface{}
		for rows.Next() {
			var title, directorVal, actorsVal, genresVal, languageVal string
			var rating float64
			var runtime int

			rows.Scan(&title, &directorVal, &actorsVal, &genresVal, &languageVal, &rating, &runtime)

			// Pour retourner la liste d'acteurs et genres proprement comme array
			actorsList := strings.Split(strings.Trim(actorsVal, "[]"), ",")
			for i := range actorsList {
				actorsList[i] = strings.Trim(actorsList[i], " '\"")
			}

			genresList := strings.Split(strings.Trim(genresVal, "[]"), ",")
			for i := range genresList {
				genresList[i] = strings.Trim(genresList[i], " '\"")
			}

			movie := map[string]interface{}{
				"title":             title,
				"director":          directorVal,
				"actors":            actorsList,
				"genres":            genresList,
				"original_language": languageVal,
				"average_rating":    rating,
				"runtime":           runtime,
			}
			movies = append(movies, movie)
		}

		c.JSON(http.StatusOK, movies)
	})

	// Lancer le serveur
	r.Run(":8081")
}
