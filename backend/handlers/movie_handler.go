package handlers

import (
    "database/sql"
    "log"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"
)

var db *sql.DB

func init() {
    var err error
    db, err = sql.Open("postgres", "user=justineletheno dbname=movies_db sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
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

// GET /movies?actors=X&directors=Y&genres=A&languages=L&min_rating=R
func GetMovies(c *gin.Context) {
    actors := c.QueryArray("actors")
    directors := c.QueryArray("directors")
    genres := c.QueryArray("genres")
    languages := c.QueryArray("languages")
    minRating := c.Query("min_rating")

    query := `SELECT title, director, actors, genres, original_language, average_rating, runtime FROM movies WHERE 1=1`
    args := []interface{}{}
    argPos := 1

    // Actors filter
    for _, a := range actors {
        query += " AND actors ILIKE '%' || $" + string(argPos) + " || '%'"
        args = append(args, a)
        argPos++
    }

    // Directors filter
    for _, d := range directors {
        query += " AND director ILIKE '%' || $" + string(argPos) + " || '%'"
        args = append(args, d)
        argPos++
    }

    // Genres filter
    for _, g := range genres {
        query += " AND genres ILIKE '%' || $" + string(argPos) + " || '%'"
        args = append(args, g)
        argPos++
    }

    // Languages filter
    for _, l := range languages {
        query += " AND original_language ILIKE '%' || $" + string(argPos) + " || '%'"
        args = append(args, l)
        argPos++
    }

    // Min rating filter
    if minRating != "" {
        query += " AND average_rating >= $" + string(argPos)
        args = append(args, minRating)
        argPos++
    }

    rows, err := db.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    movies := []Movie{}
    for rows.Next() {
        var m Movie
        var actorsStr, genresStr string
        rows.Scan(&m.Title, &m.Director, &actorsStr, &genresStr, &m.OriginalLanguage, &m.AverageRating, &m.Runtime)
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

    rows, err := db.Query("SELECT DISTINCT "+field+" FROM movies WHERE "+field+" ILIKE $1 LIMIT 10", "%"+queryStr+"%")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    suggestions := []string{}
    for rows.Next() {
        var val string
        rows.Scan(&val)
        for _, v := range strings.Split(val, ";") {
            if strings.Contains(strings.ToLower(v), strings.ToLower(queryStr)) {
                suggestions = append(suggestions, v)
            }
        }
    }
    c.JSON(http.StatusOK, suggestions)
}
