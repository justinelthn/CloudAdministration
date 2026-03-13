package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB

// allowedSuggestFields is a whitelist of columns allowed in /suggest to prevent SQL injection.
var allowedSuggestFields = map[string]bool{
	"actors":            true,
	"director":          true,
	"genres":            true,
	"original_language": true,
}

func init() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "user=come dbname=movies_db sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to PostgreSQL")
}

type Movie struct {
	Title            string   `json:"title"`
	Director         string   `json:"director"`
	Actors           []string `json:"actors"`
	Genres           []string `json:"genres"`
	OriginalLanguage string   `json:"original_language"`
	AverageRating    float64  `json:"average_rating"`
	Runtime          int      `json:"runtime"`
}

// GET /movies?actors=X&directors=Y&genres=A&languages=L&min_rating=R&page=1&limit=50
func GetMovies(c *gin.Context) {
	actors := c.QueryArray("actors")
	directors := c.QueryArray("directors")
	genres := c.QueryArray("genres")
	languages := c.QueryArray("languages")
	minRating := c.Query("min_rating")

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	query := `SELECT film_title, director, actors, genres, original_language, average_rating, runtime FROM movies_clean WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	// Actors filter
	for _, a := range actors {
		query += fmt.Sprintf(" AND actors ILIKE '%%' || $%d || '%%'", argPos)
		args = append(args, a)
		argPos++
	}

	// Directors filter
	for _, d := range directors {
		query += fmt.Sprintf(" AND director ILIKE '%%' || $%d || '%%'", argPos)
		args = append(args, d)
		argPos++
	}

	// Genres filter
	for _, g := range genres {
		query += fmt.Sprintf(" AND genres ILIKE '%%' || $%d || '%%'", argPos)
		args = append(args, g)
		argPos++
	}

	// Languages filter
	for _, l := range languages {
		query += fmt.Sprintf(" AND original_language ILIKE '%%' || $%d || '%%'", argPos)
		args = append(args, l)
		argPos++
	}

	// Min rating filter
	if minRating != "" {
		query += fmt.Sprintf(" AND average_rating >= $%d", argPos)
		args = append(args, minRating)
		argPos++
	}

	// Pagination
	query += fmt.Sprintf(" ORDER BY average_rating DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	movies := []Movie{}
	for rows.Next() {
		var m Movie
		var actorsStr, genresStr string
		if err := rows.Scan(&m.Title, &m.Director, &actorsStr, &genresStr, &m.OriginalLanguage, &m.AverageRating, &m.Runtime); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		m.Actors = strings.Split(actorsStr, ";")
		m.Genres = strings.Split(genresStr, ";")
		movies = append(movies, m)
	}
	c.JSON(http.StatusOK, movies)
}

// GET /suggest?field=actors&query=Chris
func Suggest(c *gin.Context) {
	field := c.Query("field")
	queryStr := c.Query("query")

	if !allowedSuggestFields[field] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid field, allowed: actors, director, genres, original_language"})
		return
	}

	rows, err := db.Query("SELECT DISTINCT "+field+" FROM movies_clean WHERE "+field+" ILIKE $1 LIMIT 10", "%"+queryStr+"%")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	suggestions := []string{}
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		for _, v := range strings.Split(val, ";") {
			trimmed := strings.TrimSpace(v)
			if strings.Contains(strings.ToLower(trimmed), strings.ToLower(queryStr)) {
				suggestions = append(suggestions, trimmed)
			}
		}
	}
	c.JSON(http.StatusOK, suggestions)
}
