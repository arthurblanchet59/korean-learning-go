package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestServeWebApp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	assets := filepath.Join(root, "assets")
	if err := os.Mkdir(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<main>application</main>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assets, "app.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	if err := ServeWebApp(router, root); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		path        string
		status      int
		body        string
		cacheHeader string
	}{
		{name: "spa route", path: "/revision/today", status: http.StatusOK, body: "application", cacheHeader: "no-cache"},
		{name: "static asset", path: "/assets/app.js", status: http.StatusOK, body: "console.log", cacheHeader: "immutable"},
		{name: "unknown api route", path: "/api/missing", status: http.StatusNotFound, body: "route not found"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if !strings.Contains(response.Body.String(), test.body) {
				t.Fatalf("body = %q, want content %q", response.Body.String(), test.body)
			}
			if !strings.Contains(response.Header().Get("Cache-Control"), test.cacheHeader) {
				t.Fatalf("Cache-Control = %q, want content %q", response.Header().Get("Cache-Control"), test.cacheHeader)
			}
		})
	}
}

func TestServeWebAppRequiresIndex(t *testing.T) {
	if err := ServeWebApp(gin.New(), t.TempDir()); err == nil {
		t.Fatal("ServeWebApp() error = nil, want missing index error")
	}
}
