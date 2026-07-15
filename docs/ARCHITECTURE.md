# Architecture

## Vision

Le projet est decoupe en trois surfaces:

- `backend`: API Gin lancee depuis la racine avec `go run .`.
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

La cible du rendu API est SQLite pour que `go run .` fonctionne directement depuis la racine du projet, y compris dans un GitHub Codespace.

Le repository SQLite persiste les decks, cartes et reviews dans `data/korean-learning.db` par defaut.

## Configuration et etat du TUI

Le client terminal separe trois categories de donnees:

- le JWT, conserve dans un fichier distinct et jamais exporte;
- `config.json`, qui contient l'URL de l'API et le theme de couleurs;
- `state.json`, qui contient la derniere vue, le sens de revision et le mode de bibliotheque.

Les fichiers actifs sont ranges dans `users/<identifiant-anonymise>/` afin que deux comptes utilises sur la meme machine ne partagent ni leur theme ni leur etat. Le fichier racine conserve seulement les informations necessaires avant l'authentification et permet la migration du premier profil local.

L'utilisateur peut envoyer les deux documents JSON avec `PUT /api/client-backup` et les restaurer avec `GET /api/client-backup`. Le middleware JWT fournit le `user_id`: le client ne choisit donc jamais le proprietaire du backup.

La couche serveur suit le meme decoupage que les autres fonctionnalites:

```text
handler Gin -> ClientBackupService -> ClientBackupRepository -> table client_backups
```

Le service valide la taille et la structure JSON avant l'ecriture. La table utilise `user_id` comme cle primaire et garantit un backup courant par utilisateur.
