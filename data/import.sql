-- Création de la table movies_clean
CREATE TABLE IF NOT EXISTS movies_clean (
    id                SERIAL PRIMARY KEY,
    film_title        TEXT,
    release_year      TEXT,
    director          TEXT,
    actors            TEXT,
    average_rating    FLOAT,
    owner_rating      FLOAT,
    genres            TEXT,
    runtime           INT,
    countries         TEXT,
    original_language TEXT,
    spoken_languages  TEXT,
    description       TEXT,
    studios           TEXT,
    watches           INT,
    list_appearances  INT,
    likes             INT,
    fans              INT,
    half_star         INT,
    one_star          INT,
    one_half_star     INT,
    two_star          INT,
    two_half_star     INT,
    three_star        INT,
    three_half_star   INT,
    four_star         INT,
    four_half_star    INT,
    five_star         INT,
    total_ratings     INT,
    film_url          TEXT
);

-- Import du CSV
-- Adapter le chemin ci-dessous à votre environnement
\copy movies_clean (film_title, release_year, director, actors, average_rating, owner_rating, genres, runtime, countries, original_language, spoken_languages, description, studios, watches, list_appearances, likes, fans, half_star, one_star, one_half_star, two_star, two_half_star, three_star, three_half_star, four_star, four_half_star, five_star, total_ratings, film_url) FROM './data/Movie_Data_File.csv' WITH (FORMAT csv, HEADER true, DELIMITER ',', NULL 'nan', QUOTE '"');

-- Vérification
SELECT COUNT(*) AS total_films FROM movies_clean;
SELECT film_title, director, average_rating FROM movies_clean WHERE film_title = 'Pulp Fiction';
