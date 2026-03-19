# CloudAdministration — Movie Finder

API Go + frontend pour la recherche avancée de films (Letterboxd dataset).

Lien du bucket GCS : https://console.cloud.google.com/storage/browser/dataset-letterboxd-67

## Prérequis

- **Go** ≥ 1.21
- **PostgreSQL** avec une base `movies_db`
- (Optionnel) **Docker**

## Installation

```bash
# 1. Cloner le repo
git clone <url> && cd CloudAdministration

# 2. Importer les données dans PostgreSQL
psql -U justineletheno -d movies_db -f data/import.sql

# 3. Installer les dépendances Go
cd backend && go mod download
```

## Lancement

```bash
# Variable d'environnement (optionnel, valeur par défaut incluse)
export DATABASE_URL="user=justineletheno dbname=movies_db sslmode=disable"

# Lancer le serveur
cd backend && go run main.go
# → http://localhost:8081
```

Ouvrir `frontend/index.html` dans un navigateur pour utiliser l'interface.

## Endpoints

| Méthode | Route      | Description                         |
|---------|------------|-------------------------------------|
| GET     | `/movies`  | Recherche de films avec filtres     |
| GET     | `/suggest` | Autocomplétion acteurs/réalisateurs |

### Paramètres `/movies`

| Param       | Type   | Description                    |
|-------------|--------|--------------------------------|
| `actors`    | string | Filtre par acteur (ILIKE)      |
| `directors` | string | Filtre par réalisateur (ILIKE) |
| `genres`    | string | Filtre par genre               |
| `languages` | string | Filtre par langue              |
| `min_rating`| float  | Note minimale (0–5)            |
| `page`      | int    | Page (défaut: 1)               |
| `limit`     | int    | Résultats par page (défaut: 50, max: 200) |

## Architecture Distribuée (Docker Compose)

Pour le projet d'administration cloud, le système peut être déployé sur plusieurs serveurs en parallèle derrière un Load Balancer (NGINX) avec une base de données isolée et indexée.

```bash
# 1. Démarrer l'architecture (DB + 3 API + NGINX)
docker-compose up -d --build

# 2. Attendre quelques secondes que la DB initialise les données du CSV
# 3. Lancer le test de charge massif (100k requêtes)
cd backend
go run cmd/loadtest/main.go -n 100000 -c 1000

# 4. Éteindre le cluster
docker-compose down
```

## Load Testing

```bash
cd backend
go run cmd/loadtest/main.go -url "http://localhost:8081/movies?min_rating=4" -n 500 -c 50
```
