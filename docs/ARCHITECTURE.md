# Architecture

## Vision

Le projet est decoupe en trois surfaces:

- `backend`: API HTTP et orchestration des donnees.
- `frontend`: experience web confortable pour gerer les contenus.
- `tui`: experience terminal rapide pour reviser sans ouvrir le navigateur.

Le package `core` contient les types et les regles metier partagees. L'objectif est d'eviter de dupliquer la logique de revision entre le backend et le TUI.

## Flux de revision

```txt
card due -> user answers -> rating -> scheduler -> next review date
```

Les notes possibles:

- `again`: la carte revient tres vite.
- `hard`: la carte revient bientot.
- `good`: progression normale.
- `easy`: intervalle plus long.

## Modules fonctionnels

- Decks: regroupement des cartes par theme.
- Cards: vocabulaire, phrase, hangeul ou exercice.
- Reviews: historique des reponses.
- Scheduler: calcul de `nextReviewAt`.
- Stats: cartes dues, taux de reussite, mots difficiles.

## Stockage

Le backend est la source de verite des decks, cartes et reviews. Le front ne garde pas de donnees metier en dur: il consomme l'API.

La stack locale cible PostgreSQL via Docker Compose. Un repository memoire reste disponible si `DATABASE_URL` est absent, principalement pour demarrer l'API sans dependance externe.
