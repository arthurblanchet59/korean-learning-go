package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerUIHTML = `<!doctype html>
<html lang="fr">
<head>
  <meta charset="utf-8">
  <title>Korean Learning API - Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: "/openapi.json",
        dom_id: "#swagger-ui",
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`

func registerSwaggerRoutes(router *gin.Engine) {
	router.GET("/openapi.json", func(ctx *gin.Context) {
		ctx.File("docs/openapi.json")
	})
	router.GET("/swagger", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/swagger/index.html", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIHTML))
	})
}
