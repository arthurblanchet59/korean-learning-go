package foundry

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arthurblanchet59/korean-learning-go/internal/service"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func TestCorrectCallsFoundryAndParsesStructuredCorrection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/openai/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if request.Header.Get("api-key") != "test-key" {
			t.Fatal("missing API key")
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"correctedText\":\"저는 학생이에요.\",\"corrections\":[{\"original\":\"저 는\",\"replacement\":\"저는\",\"reason\":\"La particule s'attache au pronom.\"}]}"}}]}`))
	}))
	defer server.Close()

	corrector, err := NewCorrector(server.URL, "test-key", "DeepSeek-V3.2")
	if err != nil {
		t.Fatal(err)
	}
	result, err := corrector.Correct(context.Background(), "저 는 학생이에요")
	if err != nil {
		t.Fatal(err)
	}
	if result.CorrectedText != "저는 학생이에요." || len(result.Corrections) != 1 {
		t.Fatalf("unexpected correction: %+v", result)
	}
}

func TestCorrectWithContextAddsPedagogicalSources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var payload chatRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(payload.Messages[0].Content, "Les particules") {
			t.Fatal("pedagogical context is missing from the prompt")
		}
		_, _ = response.Write([]byte(`{"choices":[{"message":{"content":"{\"correctedText\":\"저는 학생이에요.\",\"corrections\":[]}"}}]}`))
	}))
	defer server.Close()

	corrector, err := NewCorrector(server.URL, "test-key", "DeepSeek-V3.2")
	if err != nil {
		t.Fatal(err)
	}
	sources := []core.CorrectionSource{{ID: "grammar-1", Title: "Les particules", Level: "A1", Excerpt: "은/는 indique le thème."}}
	result, err := corrector.CorrectWithContext(context.Background(), "저는 학생이에요", sources)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Sources) != 1 || result.Sources[0].ID != "grammar-1" {
		t.Fatalf("unexpected sources: %+v", result.Sources)
	}
}

func TestCorrectWrapsFoundryHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		http.Error(response, `{"error":{"message":"quota exceeded"}}`, http.StatusTooManyRequests)
	}))
	defer server.Close()

	corrector, err := NewCorrector(server.URL, "test-key", "DeepSeek-V3.2")
	if err != nil {
		t.Fatal(err)
	}
	_, err = corrector.Correct(context.Background(), "안녕하세요")
	if !errors.Is(err, service.ErrCorrectionUnavailable) {
		t.Fatalf("expected correction unavailable error, got %v", err)
	}
}

func TestCompletionURLAcceptsFoundryBaseURLs(t *testing.T) {
	for input, expected := range map[string]string{
		"https://example.services.ai.azure.com":           "https://example.services.ai.azure.com/openai/v1/chat/completions",
		"https://example.openai.azure.com/openai/v1/":     "https://example.openai.azure.com/openai/v1/chat/completions",
		"https://example.test/openai/v1/chat/completions": "https://example.test/openai/v1/chat/completions",
	} {
		actual, err := completionURL(input)
		if err != nil {
			t.Fatalf("completionURL(%q): %v", input, err)
		}
		if actual != expected {
			t.Fatalf("completionURL(%q)=%q, want %q", input, actual, expected)
		}
	}
}
