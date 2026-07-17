package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var backendPaths = []string{
	"/api",
	"/admin",
	"/health",
	"/openapi.json",
	"/search",
	"/study",
	"/swagger",
	"/user",
}

// ServeWebApp serves a built single-page application without hiding API 404s.
func ServeWebApp(router *gin.Engine, root string) error {
	webRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve web root: %w", err)
	}

	indexPath := filepath.Join(webRoot, "index.html")
	if info, err := os.Stat(indexPath); err != nil || info.IsDir() {
		return fmt.Errorf("web root %q does not contain index.html", webRoot)
	}

	router.NoRoute(func(ctx *gin.Context) {
		if isBackendPath(ctx.Request.URL.Path) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}
		if ctx.Request.Method != http.MethodGet && ctx.Request.Method != http.MethodHead {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}

		if staticPath, ok := existingStaticFile(webRoot, ctx.Request.URL.Path); ok {
			if strings.HasPrefix(ctx.Request.URL.Path, "/assets/") {
				ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
			}
			ctx.File(staticPath)
			return
		}

		ctx.Header("Cache-Control", "no-cache")
		ctx.File(indexPath)
	})

	return nil
}

func isBackendPath(path string) bool {
	for _, prefix := range backendPaths {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

func existingStaticFile(root string, requestPath string) (string, bool) {
	relativePath := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(requestPath, "/")))
	if relativePath == "." || relativePath == "" {
		return "", false
	}

	candidate := filepath.Join(root, relativePath)
	relativeToRoot, err := filepath.Rel(root, candidate)
	if err != nil || relativeToRoot == ".." || strings.HasPrefix(relativeToRoot, ".."+string(filepath.Separator)) {
		return "", false
	}

	info, err := os.Stat(candidate)
	return candidate, err == nil && !info.IsDir()
}
