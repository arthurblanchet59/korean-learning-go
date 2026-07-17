# Korean Learning Go

Application complète d’apprentissage du coréen, construite autour d’une API REST Go/Gin et d’une base SQLite. Elle propose deux clients complémentaires : une interface web React et un TUI plein écran pour le terminal.

## Fonctionnalités

- comptes utilisateurs avec authentification JWT et rôle administrateur ;
- decks et cartes de vocabulaire isolés par utilisateur ;
- révision espacée avec les notes `again`, `hard`, `good` et `easy` ;
- sessions quotidiennes, cartes difficiles, statistiques et séries de révision ;
- leçons guidées de coréen avec suivi de progression ;
- journal en coréen avec corrections expliquées, locales ou générées par Microsoft Foundry à partir des leçons pertinentes ;
- recherche globale et import/export CSV ;
- configuration et état JSON du TUI avec backup personnel sur le serveur ;
- documentation OpenAPI, Swagger UI et exemples de requêtes HTTP ;
- clients React et TUI utilisant la même API.

## Prérequis

- [Go 1.22 ou supérieur](https://go.dev/dl/) ;
- Node.js 22 et Corepack pour le frontend ;
- Docker Desktop uniquement pour le lancement avec Docker.

## Démarrage rapide

### 1. Backend

Depuis la racine du dépôt :

```powershell
go run .
```

Le serveur initialise automatiquement SQLite, le compte administrateur et le contenu pédagogique. Il écoute par défaut sur `http://localhost:8080`.

Points d’entrée utiles :

- santé : `http://localhost:8080/health` ;
- Swagger UI : `http://localhost:8080/swagger/index.html` ;
- OpenAPI JSON : `http://localhost:8080/openapi.json` ;
- exemples de requêtes : [`requests.http`](requests.http).

Les identifiants administrateur par défaut sont réservés au développement local :

```text
admin@korean.local
admin123
```

Utilisez `POST /user/register` ou l’écran d’inscription pour créer un compte personnel. En production, remplacez impérativement `JWT_SECRET`, `ADMIN_EMAIL` et `ADMIN_PASSWORD`.

### 2. Frontend React

Dans un deuxième terminal :

```powershell
cd frontend
corepack pnpm install
corepack pnpm dev
```

Ouvrez ensuite `http://localhost:5173`. Le frontend contacte par défaut l’API locale sur `http://localhost:8080/api`.

### 3. TUI

Dans un autre terminal, depuis la racine :

```powershell
go run ./apps/tui
```

Au premier écran, connectez-vous ou utilisez `Ctrl+R` pour créer un compte. Le JWT est conservé séparément dans `~/.korean-learning-go/token`.

Après la connexion, l'écran **Accueil** présente les principales actions. L'écran **Paramètres** permet de choisir un thème Émeraude, Océan, Ambre ou Rose, de modifier l'URL de l'API et de sauvegarder ou restaurer les préférences depuis le serveur.

Le TUI crée deux documents JSON dans le répertoire de configuration personnel fourni par le système :

- `config.json` : version, URL de l'API et thème de couleurs ;
- `state.json` : onglet actif, sens de révision et mode cartes/decks.

Après authentification, chaque compte possède sa propre copie dans `users/<identifiant-anonymisé>/`. Changer de compte recharge donc son thème et son état sans écraser ceux du compte précédent. Le `config.json` racine sert uniquement à retrouver l'API avant la connexion et à migrer une ancienne installation.

Le répertoire est généralement `%AppData%\korean-learning-go` sous Windows, `~/.config/korean-learning-go` sous Linux et `~/Library/Application Support/korean-learning-go` sous macOS. Le JWT n'est jamais inclus dans ces documents ni dans leur backup.

Raccourcis principaux :

```text
h/l ou ←/→   changer d’onglet
j/k ou ↑/↓   naviguer
PgUp/PgDn    faire défiler une leçon ou un journal
a            saisir une réponse
v            inverser coréen/français
espace       révéler une carte
1 à 4        indiquer si la carte est à revoir, hésitante, retenue ou maîtrisée
n            créer un élément
d            supprimer l'élément actif
e            modifier l'élément actif, le profil ou l'URL API
u            envoyer config.json et state.json vers le serveur
o            restaurer config.json et state.json depuis le serveur
/            rechercher
:            ouvrir la palette de commandes
?            afficher l’aide
q            quitter
```

Pour connecter le TUI à une API distante :

```powershell
go run ./apps/tui --api "https://mon-app.azurewebsites.net"
```

La variable `KOREAN_API_URL` peut également définir cette adresse. L'ordre de priorité est `--api`, puis `KOREAN_API_URL`, puis `config.json`, puis l'API locale par défaut.

## Lancement avec Docker

Docker Compose lance le backend et le frontend dans deux conteneurs :

```powershell
docker compose up --build
```

- frontend : `http://localhost:5173` ;
- backend : `http://localhost:8080` ;
- données SQLite et logs : volumes Docker persistants.

Pour arrêter la stack :

```powershell
docker compose down
```

## Configuration

Les variables disponibles sont documentées dans [`.env.example`](.env.example). Les principales sont :

| Variable | Valeur locale par défaut | Description |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | Adresse d’écoute du backend |
| `SQLITE_PATH` | `data/korean-learning.db` | Emplacement de la base SQLite |
| `LOG_DIR` | `logs` | Répertoire des journaux applicatifs |
| `DB_SEED` | `true` | Initialise le contenu pédagogique |
| `JWT_SECRET` | valeur de développement | Clé de signature des JWT |
| `ADMIN_NAME` | `Admin` | Nom du compte administrateur initial |
| `ADMIN_EMAIL` | `admin@korean.local` | Adresse administrateur initiale |
| `ADMIN_PASSWORD` | `admin123` | Mot de passe administrateur initial |
| `AZURE_AI_ENDPOINT` | vide | URL de la ressource Microsoft Foundry ; active la correction LLM avec les deux variables suivantes |
| `AZURE_AI_API_KEY` | vide | Clé API Foundry, conservée uniquement par le backend |
| `AZURE_AI_MODEL` | vide | Nom du déploiement, par exemple `DeepSeek-V3.2` |
| `AZURE_AI_EMBEDDING_ENDPOINT` | vide | URL du déploiement Cohere Embed ; utilise `AZURE_AI_ENDPOINT` si elle est vide |
| `AZURE_AI_EMBEDDING_API_KEY` | vide | Clé du déploiement Embed ; utilise `AZURE_AI_API_KEY` si elle est vide |
| `AZURE_AI_EMBEDDING_MODEL` | vide | Nom du déploiement, par exemple `embed-v-4-0` ; active le RAG |
| `AZURE_AI_EMBEDDING_DIMENSIONS` | `1024` | Taille des vecteurs Cohere : 256, 512, 1024 ou 1536 |
| `VITE_API_BASE_URL` | `http://localhost:8080/api` | API utilisée par React au moment du build |

Exemple PowerShell :

```powershell
$env:JWT_SECRET = "une-cle-longue-et-aleatoire"
$env:ADMIN_PASSWORD = "un-mot-de-passe-fort"
go run .
```

Le fichier `.env.example` sert de référence : l’application Go ne charge pas automatiquement un fichier `.env`.

### Correction du journal avec Microsoft Foundry

Le journal utilise le correcteur local tant que les variables Azure sont vides. Pour activer DeepSeek, récupérer l’endpoint et la clé dans Microsoft Foundry puis lancer le backend ainsi :

Pour le développement local, copie `.env.example` vers `.env` et complète les valeurs. `go run .` charge automatiquement `.env`, sans remplacer les variables déjà définies dans le terminal. Le fichier `.env` est ignoré par Git.

```powershell
$env:AZURE_AI_ENDPOINT = "https://<ressource>.services.ai.azure.com"
$env:AZURE_AI_API_KEY = "<cle-api>"
$env:AZURE_AI_MODEL = "DeepSeek-V3.2"
go run .
```

L’endpoint peut aussi être l’URL de base terminée par `/openai/v1` ou l’URL complète terminée par `/chat/completions`. La clé n’est jamais envoyée au frontend ni au TUI : les deux clients passent par les routes protégées `/api/journal` et `/api/journal/correct`.

Pour enrichir les corrections avec les leçons, déployer Cohere Embed v4 puis ajouter :

```powershell
$env:AZURE_AI_EMBEDDING_ENDPOINT = "https://<ressource>.services.ai.azure.com"
$env:AZURE_AI_EMBEDDING_API_KEY = "<cle-api>"
$env:AZURE_AI_EMBEDDING_MODEL = "embed-v-4-0"
$env:AZURE_AI_EMBEDDING_DIMENSIONS = "1024"
```

Au démarrage, le backend découpe les leçons, calcule leurs embeddings et stocke l’index dans SQLite. Il ne recalcule cet index que si le corpus, le modèle ou la dimension change. Une correction recherche ensuite les quatre passages les plus proches, les transmet à DeepSeek et renvoie les leçons utilisées. Un administrateur peut contrôler l’état via `GET /api/rag/status` et reconstruire l’index via `POST /admin/rag/reindex`, le panel web ou la touche `i` du TUI.

Le client détecte automatiquement les ressources Foundry unifiées en `services.ai.azure.com` et utilise leur route `/models/embeddings`. Il reste compatible avec un endpoint Cohere dédié en `models.ai.azure.com`.

## Tests et qualité

Backend, cœur métier et TUI :

```powershell
go test ./...
go test ./packages/core
go test ./apps/tui
go vet ./...
```

Frontend :

```powershell
cd frontend
corepack pnpm test
corepack pnpm build
```

Les tests couvrent notamment l’authentification, le CRUD des decks, le reset administrateur, l’isolation des utilisateurs, les transactions SQLite, le scheduler et les comportements React sensibles.

## Architecture

```text
internal/
  api/                   routes et handlers Gin par domaine
  service/               règles métier et orchestration
  repository/sqlite/     persistance SQLite par domaine
apps/
  tui/                   client terminal Bubble Tea
frontend/                client React/Vite
packages/core/           modèles et scheduler partagés
docs/                    architecture, OpenAPI et déploiement
```

Le backend est la source de vérité : le frontend et le TUI ne stockent pas les decks, cartes, révisions ou progressions localement. Le découpage détaillé est présenté dans [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

Les préférences d'interface du TUI sont une exception volontaire : elles sont disponibles hors ligne dans les fichiers JSON locaux et peuvent être sauvegardées dans la table SQLite `client_backups`. Les routes protégées `GET /api/client-backup` et `PUT /api/client-backup` isolent automatiquement le backup grâce à l'identité du JWT.

## Données et logs

En lancement local :

- base SQLite : `data/korean-learning.db` ;
- logs normaux : `logs/app.log` ;
- logs d’erreur : `logs/error.log`.

Le reset administrateur supprime les données d’apprentissage et les comptes non administrateurs, mais conserve le compte administrateur.

## Déploiement Azure

La production utilise [`Dockerfile.azure`](Dockerfile.azure), qui rassemble React, Gin et SQLite dans un conteneur unique. SQLite et les logs sont persistés sous `/home` sur Azure App Service.

Le guide complet couvre la création de l’App Service, les variables, GitHub Actions, GHCR et le diagnostic : [`docs/AZURE_APP_SERVICE.md`](docs/AZURE_APP_SERVICE.md).

Avec SQLite, conservez une seule instance de l’application. Une mise à l’échelle horizontale nécessiterait une base de données externe.

## Pistes d’évolution

- enrichir les exercices de grammaire et les leçons ;
- améliorer progressivement les suggestions du journal ;
- ajouter un format d’import compatible Anki en complément du CSV ;
- migrer vers une base externe si l’application devient multi-instance.
