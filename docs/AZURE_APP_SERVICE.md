# Deploiement Azure App Service

L'application est livree dans un conteneur unique. Le frontend React est compile pendant la construction de l'image, puis servi par le backend Gin. SQLite et les logs sont places sous `/home`, le stockage persistant fourni par App Service.

## 1. Verifier l'App Service

L'App Service doit utiliser les options suivantes:

- publication: `Conteneur`;
- systeme d'exploitation: `Linux`;
- plan: `F1` pour les essais;
- une seule instance, requise par SQLite.

L'image de production est definie dans `Dockerfile.azure`. Elle expose le port `8080` et contient a la fois l'API et le frontend.

## 2. Configurer GitHub

Dans le depot GitHub, ouvrir `Settings > Secrets and variables > Actions`.

Creer la variable de depot suivante:

```text
AZURE_WEBAPP_NAME=nom-exact-de-l-app-service
```

Dans Azure, ouvrir l'App Service puis telecharger le profil depuis `Overview > Get publish profile`. Ajouter tout le contenu du fichier comme secret GitHub:

```text
AZURE_WEBAPP_PUBLISH_PROFILE=<contenu du profil publie par Azure>
```

Si Azure refuse le telechargement du profil, activer temporairement les informations de publication de base dans `Configuration > General settings`, telecharger le profil, puis conserver le fichier comme un secret.

Le workflow `.github/workflows/azure-app-service.yml` construit l'image, la publie sur GitHub Container Registry et met a jour l'App Service. Le package GHCR doit etre rendu public depuis la page `Packages` du profil GitHub afin qu'Azure puisse le telecharger sans identifiants de registre supplementaires.

## 3. Configurer les variables Azure

Dans `App Service > Settings > Environment variables`, ajouter:

```text
HTTP_ADDR=:8080
WEBSITES_PORT=8080
WEBSITES_ENABLE_APP_SERVICE_STORAGE=true
SQLITE_PATH=/home/data/korean-learning.db
LOG_DIR=/home/logs
DB_SEED=true
GIN_MODE=release
JWT_SECRET=<secret long et aleatoire>
ADMIN_NAME=Admin
ADMIN_EMAIL=<adresse de l'administrateur>
ADMIN_PASSWORD=<mot de passe fort d'au moins 8 caracteres>
```

Ne pas definir `WEB_ROOT`: l'image le configure deja sur `/app/web`. Enregistrer les variables puis redemarrer l'App Service.

`WEBSITES_ENABLE_APP_SERVICE_STORAGE=true` est indispensable. Sans cette valeur, la base SQLite disparait lors du remplacement du conteneur.

## 4. Lancer le premier deploiement

Ouvrir `Actions > Deploy Azure App Service > Run workflow` dans GitHub. Apres le premier build, rendre le package `korean-learning-go` public si GitHub l'a cree en mode prive, puis relancer le workflow.

Verifier ensuite:

```text
https://<nom-app>.azurewebsites.net/
https://<nom-app>.azurewebsites.net/health
https://<nom-app>.azurewebsites.net/swagger/index.html
https://<nom-app>.azurewebsites.net/openapi.json
```

La route `/health` doit repondre avec `{"status":"ok"}`. Le TUI peut utiliser la meme API avec:

```text
KOREAN_API_URL=https://<nom-app>.azurewebsites.net
```

## 5. Diagnostic

Les journaux applicatifs persistants sont ecrits dans:

```text
/home/logs/app.log
/home/logs/error.log
```

Les erreurs de demarrage du conteneur sont visibles dans `App Service > Monitoring > Log stream`. Les causes les plus courantes sont une image GHCR encore privee, un profil de publication invalide, ou le port `WEBSITES_PORT` absent.

Le plan F1 peut mettre l'application en veille et provoquer un premier chargement lent. Il ne faut pas activer plusieurs instances avec SQLite.
