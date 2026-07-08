# Korean Learning Go

Application personnelle pour apprendre le coreen avec trois interfaces:

- un backend HTTP en Go;
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
apps/
  backend/   API HTTP
  tui/       interface terminal
frontend/   interface web statique initiale
packages/
  core/      domaine partage: decks, cartes, reviews, scheduling
docs/
  ARCHITECTURE.md
```

## Commandes prevues

```powershell
go test ./...
go run ./apps/backend
go run ./apps/tui -- today
go run ./apps/tui -- review
```

Pour le front:

```powershell
cd frontend
npm run dev
```

## Roadmap courte

1. Base domaine + scheduler de revision.
2. Backend avec stockage local.
3. TUI utilisable pour `today`, `add`, `review`, `stats`.
4. Front web pour dashboard, decks, ajout de cartes et revision.
5. Import/export CSV compatible Anki.
