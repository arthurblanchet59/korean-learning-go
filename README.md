# Korean Learning Go

Application personnelle pour apprendre le coreen avec trois interfaces:

- une API REST en Go avec Gin et SQLite;
- une interface web pour gerer les decks, les cartes et les sessions;
- un TUI/CLI pour reviser vite depuis le terminal.

Le projet vise un usage quotidien, proche d'Anki pour la revision espacee, avec un parcours guide pour le hangeul, le vocabulaire et les phrases utiles.

## Objectifs MVP

- Creer et organiser des decks de vocabulaire.
- Ajouter des cartes avec mot coreen, traduction, romanisation, exemples et tags.
- Reviser les cartes dues avec une notation `again`, `hard`, `good`, `easy`.
- Calculer la prochaine revision via un algorithme simple de repetition espacee.
- Exposer une API REST pour le front et le TUI.
- Fournir un TUI rapide pour les revisions quotidiennes.
- Isoler les decks, cartes, revisions et statistiques de chaque utilisateur.
- Proposer des lecons guidees et suivre leur progression.
- Ecrire un journal en coreen avec des corrections automatiques expliquees.
- Importer et exporter les cartes au format CSV.

## Structure

```txt
internal/    API Gin, services et repository SQLite
apps/
  backend/   ancienne API experimentale
  tui/       interface terminal
frontend/   interface web React
packages/
  core/      domaine partage: decks, cartes, reviews, scheduling
docs/
  ARCHITECTURE.md
  AZURE_APP_SERVICE.md
  openapi.json
```

## Commandes prevues

```powershell
go test ./...
go run .
go run ./apps/tui
```

Une fois le backend lance avec `go run .`:

- API: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- OpenAPI JSON: `http://localhost:8080/openapi.json`
- Exemples de requetes: `requests.http`
- Logs normaux: `logs/app.log`
- Logs erreurs: `logs/error.log`

Pour le front:

```powershell
cd frontend
corepack pnpm install
corepack pnpm dev
```

## Lancement avec Docker

La stack Docker lance:

- le backend Gin/SQLite sur `http://localhost:8080`;
- le frontend React/Nginx sur `http://localhost:5173`.

Le chemin attendu pour le rendu API est pour l'instant le lancement local SQLite:

```powershell
go run .
```

```powershell
docker compose up --build
```

Docker Desktop doit etre demarre avant d'executer cette commande.

## Deploiement Azure App Service

Le deploiement Azure utilise un conteneur unique qui sert le frontend React, l'API Gin et la base SQLite persistante sous `/home`. La configuration complete de l'App Service, de GitHub Actions et des secrets est decrite dans [`docs/AZURE_APP_SERVICE.md`](docs/AZURE_APP_SERVICE.md).

Variables utiles:

```txt
HTTP_ADDR=:8080
SQLITE_PATH=data/korean-learning.db
LOG_DIR=logs
DB_SEED=true
JWT_SECRET=dev-secret-change-me
ADMIN_EMAIL=admin@korean.local
ADMIN_PASSWORD=admin123
VITE_API_BASE_URL=http://localhost:8080/api
```

En local:

```powershell
go run .
```

Dans un second terminal, le TUI plein ecran se lance avec:

```powershell
go run ./apps/tui
```

Le TUI utilise `KOREAN_API_URL` (par defaut `http://localhost:8080`) et conserve le JWT dans le dossier personnel. Raccourcis principaux: `h/l` pour les onglets, `j/k` pour naviguer, espace pour reveler une carte, `1` a `4` pour noter, `/` pour rechercher, `:` pour la palette de commandes et `?` pour l'aide.

## Roadmap courte

1. Base domaine + scheduler de revision.
2. Backend avec stockage local.
3. Enrichir les exercices de grammaire et la bibliotheque de lecons.
4. Ajouter un vrai format d'import `.apkg` en complement du CSV.
5. Ameliorer progressivement le correcteur local du journal.
