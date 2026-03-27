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

> ### 🎯 Explication Simple (Baseline)
>
> **Analogie :** Imaginez un seul guichet dans une gare avec 1000 personnes qui font la queue. Le guichetier (notre serveur Go) prend chaque demande et doit aller chercher l'information dans un énorme classeur (PostgreSQL) en lisant **toutes les fiches une par une** pour trouver la bonne.
>
> **Ce que fait notre code :**
> - On a **1 seul serveur Go** (`main.go`) qui écoute sur le port 8081.
> - Quand quelqu'un cherche un film, le handler Go (`movie_handler.go`) envoie une requête SQL à PostgreSQL du type : `SELECT ... FROM movies_clean WHERE actors ILIKE '%Nolan%'`.
> - `ILIKE '%Nolan%'` signifie "cherche le mot Nolan n'importe où dans le texte". Pour faire ça, PostgreSQL doit **lire chaque ligne du tableau** une par une (= "Sequential Scan"). Avec des milliers de films, c'est très lent.
> - Si 1000 personnes font cette recherche en même temps → le serveur et la base de données sont saturés → erreurs.

---

## 2. Optimisation 1 : Parallélisme et Scaling Horizontal (Docker & Load Balancing)

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

> ### 🎯 Explication Simple (Scaling Horizontal + Load Balancing)
>
> **Analogie :** On a remplacé notre guichet unique par **3 guichets** (3 serveurs Go), et on a mis un **agent d'accueil** (NGINX) à l'entrée qui distribue les clients vers le guichet le moins occupé. C'est le **Load Balancing** (équilibrage de charge).
>
> **Ce que fait notre code concrètement :**
> - Dans `docker-compose.yml` : la ligne `replicas: 3` dit à Docker de lancer **3 copies identiques** de notre serveur Go. Chaque copie est un "conteneur" isolé, comme 3 ordinateurs virtuels qui tournent en parallèle.
> - `nginx.conf` : NGINX est configuré en mode **"proxy inverse"**. Il écoute les requêtes des utilisateurs sur le port 80, puis les renvoie vers un des 3 serveurs Go (port 8081) en **Round-Robin** (chacun son tour : serveur 1, puis 2, puis 3, puis 1, etc.).
> - Le `Dockerfile` utilise un **multi-stage build** : d'abord il compile le code Go en un fichier exécutable, puis il copie juste ce fichier dans un petit conteneur Alpine Linux léger (pour ne pas gaspiller d'espace disque).
>
> **Pourquoi ça n'a pas suffi ?**
> Les 3 guichets sont plus rapides pour accueillir les clients, **mais ils partagent tous le même classeur** (PostgreSQL). Résultat : 3 guichetiers essaient de lire le classeur en même temps → encore plus de bazar qu'avant. Le problème n'était pas le nombre de guichets, mais la **vitesse de lecture du classeur**.

---

## 3. Optimisation 2: Indexation de la Base de Données

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

> ### 🎯 Explication Simple (Indexation)
>
> **Analogie :** Imaginez que le classeur de la gare, c'est un **dictionnaire sans ordre alphabétique**. Pour trouver le mot "Nolan", il faut lire toutes les pages une par une (c'est le **Sequential Scan**). L'index, c'est comme **ajouter un sommaire alphabétique** à la fin du dictionnaire : on va directement à la bonne page.
>
> **Ce que fait notre code concrètement :**
> - On a activé l'extension `pg_trgm` dans PostgreSQL. Cette extension découpe chaque mot en **petits morceaux de 3 lettres** ("trigrammes"). Par exemple, "Nolan" → `"nol"`, `"ola"`, `"lan"`. PostgreSQL range ces morceaux dans un index spécial.
> - L'index **GIN** (Generalized Inverted Index) fonctionne comme l'**index d'un livre** : au lieu de lire tout le livre page par page, on regarde à la lettre N dans l'index → il nous dit "va page 42". Pareil ici : quand on cherche `ILIKE '%Nolan%'`, PostgreSQL regarde dans l'index GIN → il trouve immédiatement les lignes qui contiennent "Nolan" sans lire les milliers d'autres lignes.
> - L'index **B-Tree** sur `average_rating` fonctionne comme un **arbre de décision** : pour trouver les films avec une note ≥ 4, au lieu de regarder tous les films, il suit les branches de l'arbre et élimine d'un coup tous les films en dessous de 4.
> - Dans notre code Go (`movie_handler.go`), la requête SQL n'a **pas changé du tout** ! C'est PostgreSQL qui, en coulisses, utilise automatiquement l'index quand il détecte qu'il en existe un.
>
> **Le compromis en une phrase :** On gagne en vitesse de lecture, mais l'index prend de la place sur le disque et ralentit un peu les écritures (ajout/modification de films).

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

> ### 🎯 Explication Simple (Cache)
>
> **Analogie :** Imaginez qu'au guichet, quelqu'un demande les horaires du train Paris-Lyon. Le guichetier va chercher dans le classeur (PostgreSQL), trouve la réponse, et la donne. 5 secondes plus tard, un autre client pose **exactement la même question**. Au lieu de retourner dans le classeur, le guichetier a **noté la réponse sur un Post-it** collé sur son bureau. Il lit le Post-it et répond **instantanément**. C'est le **cache**.
>
> **Ce que fait notre code concrètement :**
> - Dans `movie_handler.go`, on a créé une variable `movieCache` qui est un **dictionnaire en mémoire RAM** (un `map[string]CacheItem` en Go). La clé de ce dictionnaire, c'est l'URL exacte de la requête (par exemple `movies:actors=Nolan&min_rating=4`).
> - Quand une requête arrive, le code fait d'abord `cacheMutex.RLock()` puis cherche si cette URL existe déjà dans le dictionnaire. Si **oui** ("Cache Hit") et que le Post-it n'a pas expiré (< 60 secondes), il renvoie **directement** la réponse stockée en RAM, **sans jamais toucher PostgreSQL**. C'est quasi instantané.
> - Si **non** ("Cache Miss"), il fait la requête SQL normalement, puis **stocke le résultat** dans le dictionnaire avec un timer de 60 secondes (`cacheTTL = 60 * time.Second`). Les prochains clients qui poseront la même question dans les 60 prochaines secondes recevront la réponse instantanée.
> - `sync.RWMutex` : Comme 3 serveurs Go (et des milliers de goroutines) accèdent au dictionnaire en même temps, il faut un "verrou" pour éviter les conflits. `RLock` = plusieurs peuvent **lire** en même temps (rapide). `Lock` = un seul peut **écrire** à la fois (pour ajouter un nouveau résultat au cache).
>
> **Le compromis en une phrase :** La réponse est ultra-rapide, mais si un film change dans la base de données, les utilisateurs verront l'ancienne version pendant maximum 60 secondes.
