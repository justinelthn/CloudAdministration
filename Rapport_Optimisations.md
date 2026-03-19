# Rapport d'Optimisation - Projet Cloud Administration

## 1. Baseline (Point de départ)

**Objectif :** Mesurer les performances du système sans aucune optimisation.
**Test :** 100 000 requêtes, 1000 utilisateurs concurrents.

**Résultats :**
- **Débit (Throughput) :** ~1625 requêtes/seconde
- **Latence moyenne :** 476 ms
- **Latence p99 (les 1% les plus lents) :** 4615 ms
- **Erreurs :** ~1268 erreurs (HTTP 500) sur 100 000.

**Analyse :** 
Sans optimisation, le système se comporte correctement jusqu'à un certain point, puis la base de données PostgreSQL ou le serveur Go saturent, ce qui génère des erreurs (Timeout/Connexions rejetées) et fait exploser le temps de réponse (p99 > 4 secondes).

---

## 2. Optimisation 1 : Indexation de la Base de Données

**Objectif :** Empêcher PostgreSQL de lire toutes les lignes du tableau une par une (Sequential Scan) à chaque fois qu'on cherche un mot partiel (exp: `ILIKE '%Nolan%'`).
**Mise en œuvre :**
- Ajout de l'extension PostgreSQL `pg_trgm`.
- Création d'index inversés `GIN (gin_trgm_ops)` sur les colonnes textuelles (`director`, `actors`, `genres`) et un index hiérarchique (B-Tree) sur la note.

**Résultats (Test 100 000 requêtes, 1000 clients concurrents) :**
- **Débit (Throughput) :** ~2080 requêtes/seconde
- **Latence moyenne :** 318 ms
- **Latence p50 (médiane) :** 21 ms (Très forte baisse !)
- **Erreurs :** ~4400 erreurs (Moitié moins de timeouts)

**Analyse du Trade-off (Compromis) :**
L'ajout d'Index GIN a drastiquement accéléré la recherche pour la majorité des requêtes (la médiane p50 est passée de 24ms à 21ms et le débit global a augmenté de 25%). La base de données "respire" mieux car elle trouve les résultats sans balayer tout le disque dur. 
Cependant, l'Index apporte un coût caché : la base de données est maintenant plus lourde sur le disque dur, et chaque fois que nous ajouterons un nouveau film, l'insertion sera plus lente car PostgreSQL devra mettre à jour l'Index GIN en temps réel.
De plus, malgré un meilleur débit, la base de données finit toujours par être débordée par 1000 connexions concurrentes (les erreurs de timeouts n'ont pas disparu).

---

## 3. Optimisation 2 : Parallélisme et Scaling Horizontal (Docker & Load Balancing)

**Objectif :** Distribuer la charge sur plusieurs serveurs Go fonctionnant en parallèle (parallélisme au niveau de l'infrastructure) pour résoudre le goulot d'étranglement CPU/Réseau du serveur web.

**Mise en œuvre :**
- Utilisation de `docker-compose` pour virtualiser l'environnement.
- Déploiement de **3 instances** (replicas) de notre backend Go.
- Déploiement de **NGINX** configuré en Load Balancer (Equilibreur de charge). NGINX reçoit les 1000 requêtes concurrentes et les distribue équitablement (algorithme Round-Robin) sur les 3 serveurs Go.

**Résultats (Test 100 000 requêtes, 1000 clients concurrents) :**
- **Débit (Throughput) :** ~1569 requêtes/seconde
- **Latence moyenne :** 374 ms
- **Latence p99 (les 1% les plus lents) :** 3486 ms
- **Erreurs :** ~9000 erreurs (HTTP 500/502/Timeout)

**Analyse du Trade-off (Compromis) :**
*Pourquoi le débit n'a-t-il pas triplé alors qu'on a mis 3 serveurs ?*
Parce que le goulot d'étranglement principal (bottleneck) n'est pas le CPU du serveur web Go, mais la **Base de Données PostgreSQL**. 
En lançant 3 serveurs Go simultanément, on a réussi à traiter plus de requêtes web en parallèle (la latence moyenne a un peu baissé et le p99 est passé de 4.6s à 3.4s), mais on a **saturé PostgreSQL beaucoup plus vite**. Les 3 API Go envoient 3 fois plus de lourdes requêtes SQL non-optimisées en même temps, épuisant les connexions de la base de données, ce qui explique l'explosion du nombre d'erreurs (Timeouts).

**Conclusion de cette étape :**
Le scaling horizontal (ajouter des serveurs) est inutile — et même contre-productif — si la couche de données (Database) est le vrai problème. Pour scale, il faut d'abord rendre les requêtes SQL efficaces !

---

## 4. Optimisation 3 : Mise en Cache Applicatif (Cache-Aside)

**Objectif :** Éviter complètement de solliciter la base de données PostgreSQL pour des requêtes identiques répétées dans un laps de temps court.
**Mise en œuvre :**
- Développement d'un Cache en mémoire vive (RAM) directement dans le code Go (`movieCache map[string]CacheItem`).
- Utilisation de `sync.RWMutex` pour garantir la sécurité d'accès entre les milliers de goroutines (Thread safety).
- La clé de cache utilisée est l'URL de la requête, avec un Time-To-Live (TTL) de 60 secondes.

**Résultats (Test 100 000 requêtes, 1000 clients concurrents) :**
- **Débit (Throughput) :** ~5606 requêtes/seconde (💥 Explosion des performances)
- **Latence p50 (médiane) :** 97 ms
- **Latence p99 (les 1% les plus lents) :** 140 ms (Énorme régularité)
- **Erreurs :** Moins de 1% d'erreurs (Seulement au tout début quand le cache est encore vide et que les requêtes touchent PostgreSQL).

**Analyse du Trade-off (Compromis) :**
Le caching offre de loin le meilleur retour sur investissement technique. Le débit a été **multiplié par 3,5** par rapport à la baseline, et le système est d'une stabilité extrême, car il sert des données depuis la RAM instantanément au lieu d'ouvrir une socket réseau vers PostgreSQL. L'API est devenue invulnérable à la charge.
**Le compromis majeur** : L'obsolescence des données. Si un film est mis à jour dans PostgreSQL, les utilisateurs recevront l'ancienne version via l'API pendant un maximum de 60 secondes (principe de "stale cache"). De plus, chaque instance Docker stocke désormais son propre cache indépendamment (pas de cache mondial Redis partagé), ce qui signifie que chaque conteneur Go consomme de la RAM pour stocker les json des films en double.
