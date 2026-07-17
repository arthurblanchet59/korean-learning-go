package foundry

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arthurblanchet59/korean-learning-go/internal/service"
)

func TestEmbedderUsesCohereNativeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/embed" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if request.Header.Get("api-key") != "embed-key" {
			t.Fatal("missing API key")
		}
		var payload embeddingRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.InputType != embeddingTypeDocument || payload.OutputDimension != 2 {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"embeddings":{"float":[[1,0],[0,1]]}}`))
	}))
	defer server.Close()

	embedder := testEmbedder(t, server.URL)
	vectors, err := embedder.EmbedDocuments(context.Background(), []string{"premier", "second"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vectors) != 2 || len(vectors[0]) != 2 {
		t.Fatalf("unexpected vectors: %+v", vectors)
	}
}

func TestEmbedQueryUsesSearchQueryInputType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var payload embeddingRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.InputType != embeddingTypeQuery {
			t.Fatalf("expected search_query, got %q", payload.InputType)
		}
		_, _ = response.Write([]byte(`{"embeddings":{"float_":[[0.5,0.5]]}}`))
	}))
	defer server.Close()

	embedder := testEmbedder(t, server.URL+"/v1")
	if _, err := embedder.EmbedQuery(context.Background(), "phrase"); err != nil {
		t.Fatal(err)
	}
}

func TestEmbedderUsesFoundryModelsAPIForUnifiedResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/models/embeddings" || request.URL.Query().Get("api-version") == "" {
			t.Fatalf("unexpected Foundry URL: %s", request.URL.String())
		}
		if request.Header.Get("extra-parameters") != "pass-through" {
			t.Fatal("missing Foundry pass-through header")
		}
		var payload foundryEmbeddingRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.InputType != foundryTypeQuery || payload.Dimensions != 2 || len(payload.Input) != 1 {
			t.Fatalf("unexpected Foundry payload: %+v", payload)
		}
		_, _ = response.Write([]byte(`{"data":[{"embedding":[0.25,0.75],"index":0}],"model":"embed-v4.0","object":"list"}`))
	}))
	defer server.Close()

	embedder := testEmbedder(t, server.URL+"/models/embeddings")
	vector, err := embedder.EmbedQuery(context.Background(), "phrase")
	if err != nil {
		t.Fatal(err)
	}
	if len(vector) != 2 || vector[0] != 0.25 {
		t.Fatalf("unexpected Foundry vector: %+v", vector)
	}
}

func TestEmbedderWrapsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		http.Error(response, "quota exceeded", http.StatusTooManyRequests)
	}))
	defer server.Close()

	embedder := testEmbedder(t, server.URL)
	_, err := embedder.EmbedQuery(context.Background(), "phrase")
	if !errors.Is(err, service.ErrEmbeddingUnavailable) {
		t.Fatalf("expected embedding unavailable error, got %v", err)
	}
}

func TestEmbeddingURLAcceptsDeploymentURLs(t *testing.T) {
	for input, expected := range map[string]string{
		"https://example.models.ai.azure.com":          "https://example.models.ai.azure.com/v1/embed",
		"https://example.models.ai.azure.com/v1/":      "https://example.models.ai.azure.com/v1/embed",
		"https://example.models.ai.azure.com/v1/embed": "https://example.models.ai.azure.com/v1/embed",
		"https://example.services.ai.azure.com":        "https://example.services.ai.azure.com/models/embeddings?api-version=2024-05-01-preview",
	} {
		actual, err := embeddingURL(input)
		if err != nil {
			t.Fatalf("embeddingURL(%q): %v", input, err)
		}
		if actual != expected {
			t.Fatalf("embeddingURL(%q)=%q, want %q", input, actual, expected)
		}
	}
}

func testEmbedder(t *testing.T, endpoint string) *Embedder {
	t.Helper()
	embedder, err := NewEmbedder(endpoint, "embed-key", "embed-v-4-0", 256)
	if err != nil {
		t.Fatal(err)
	}
	embedder.dimensions = 2
	return embedder
}
