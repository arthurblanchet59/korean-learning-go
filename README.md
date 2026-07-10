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
  openapi.json
```

## Commandes prevues

```powershell
go test ./...
go run .
go run ./apps/tui -- today
go run ./apps/tui -- review
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

## Roadmap courte

1. Base domaine + scheduler de revision.
2. Backend avec stockage local.
3. TUI utilisable pour `today`, `add`, `review`, `stats`.
4. Front web pour dashboard, decks, ajout de cartes et revision.
5. Import/export CSV compatible Anki.
