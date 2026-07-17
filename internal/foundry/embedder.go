package foundry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/internal/service"
)

const (
	embeddingTypeDocument = "search_document"
	embeddingTypeQuery    = "search_query"
	foundryTypeDocument   = "document"
	foundryTypeQuery      = "query"
)

type embeddingAPI int

const (
	cohereNativeAPI embeddingAPI = iota
	foundryModelsAPI
)

type Embedder struct {
	endpoint   string
	api        embeddingAPI
	apiKey     string
	model      string
	dimensions int
	httpClient *http.Client
}

type embeddingRequest struct {
	Model           string   `json:"model"`
	Texts           []string `json:"texts"`
	InputType       string   `json:"input_type"`
	EmbeddingTypes  []string `json:"embedding_types"`
	OutputDimension int      `json:"output_dimension"`
	Truncate        string   `json:"truncate"`
}

type embeddingResponse struct {
	Embeddings struct {
		Float      [][]float64 `json:"float"`
		FloatAlias [][]float64 `json:"float_"`
	} `json:"embeddings"`
}

type foundryEmbeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	InputType      string   `json:"input_type"`
	Dimensions     int      `json:"dimensions"`
	EncodingFormat string   `json:"encoding_format"`
}

type foundryEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

func NewEmbedder(endpoint string, apiKey string, model string, dimensions int) (*Embedder, error) {
	embedEndpoint, api, err := embeddingEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("Azure AI embedding API key is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("Azure AI embedding deployment name is required")
	}
	switch dimensions {
	case 256, 512, 1024, 1536:
	default:
		return nil, fmt.Errorf("embedding dimensions must be 256, 512, 1024 or 1536")
	}

	return &Embedder{
		endpoint:   embedEndpoint,
		api:        api,
		apiKey:     strings.TrimSpace(apiKey),
		model:      strings.TrimSpace(model),
		dimensions: dimensions,
		httpClient: &http.Client{Timeout: 45 * time.Second},
	}, nil
}

func (embedder *Embedder) Model() string {
	return embedder.model
}

func (embedder *Embedder) Dimensions() int {
	return embedder.dimensions
}

func (embedder *Embedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
	return embedder.embed(ctx, texts, embeddingTypeDocument)
}

func (embedder *Embedder) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	embeddings, err := embedder.embed(ctx, []string{text}, embeddingTypeQuery)
	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}

func (embedder *Embedder) embed(ctx context.Context, texts []string, inputType string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, embeddingError("at least one text is required", nil)
	}
	var payload any
	if embedder.api == foundryModelsAPI {
		foundryInputType := foundryTypeDocument
		if inputType == embeddingTypeQuery {
			foundryInputType = foundryTypeQuery
		}
		payload = foundryEmbeddingRequest{
			Model:          embedder.model,
			Input:          texts,
			InputType:      foundryInputType,
			Dimensions:     embedder.dimensions,
			EncodingFormat: "float",
		}
	} else {
		payload = embeddingRequest{
			Model:           embedder.model,
			Texts:           texts,
			InputType:       inputType,
			EmbeddingTypes:  []string{"float"},
			OutputDimension: embedder.dimensions,
			Truncate:        "END",
		}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, embeddingError("encode embedding request", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, embedder.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, embeddingError("create embedding request", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("api-key", embedder.apiKey)
	request.Header.Set("Authorization", "Bearer "+embedder.apiKey)
	if embedder.api == foundryModelsAPI {
		request.Header.Set("extra-parameters", "pass-through")
	}

	response, err := embedder.httpClient.Do(request)
	if err != nil {
		return nil, embeddingError("call embedding deployment", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, embeddingError("read embedding response", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, embeddingError(
			fmt.Sprintf("embedding deployment returned HTTP %d: %s", response.StatusCode, compactMessage(responseBody)),
			nil,
		)
	}

	var embeddings [][]float64
	if embedder.api == foundryModelsAPI {
		var result foundryEmbeddingResponse
		if err := json.Unmarshal(responseBody, &result); err != nil {
			return nil, embeddingError("decode Foundry embedding response", err)
		}
		embeddings = make([][]float64, len(result.Data))
		for _, item := range result.Data {
			if item.Index < 0 || item.Index >= len(embeddings) {
				return nil, embeddingError("Foundry returned an invalid embedding index", nil)
			}
			embeddings[item.Index] = item.Embedding
		}
	} else {
		var result embeddingResponse
		if err := json.Unmarshal(responseBody, &result); err != nil {
			return nil, embeddingError("decode Cohere embedding response", err)
		}
		embeddings = result.Embeddings.Float
		if len(embeddings) == 0 {
			embeddings = result.Embeddings.FloatAlias
		}
	}
	if len(embeddings) != len(texts) {
		return nil, embeddingError(
			fmt.Sprintf("embedding deployment returned %d vectors for %d texts", len(embeddings), len(texts)),
			nil,
		)
	}
	for _, embedding := range embeddings {
		if len(embedding) != embedder.dimensions {
			return nil, embeddingError(
				fmt.Sprintf("embedding deployment returned %d dimensions, expected %d", len(embedding), embedder.dimensions),
				nil,
			)
		}
	}
	return embeddings, nil
}

func embeddingEndpoint(value string) (string, embeddingAPI, error) {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", cohereNativeAPI, fmt.Errorf("AZURE_AI_EMBEDDING_ENDPOINT must be an absolute URL")
	}

	switch {
	case strings.HasSuffix(parsed.Path, "/models/embeddings"):
		if parsed.Query().Get("api-version") == "" {
			query := parsed.Query()
			query.Set("api-version", "2024-05-01-preview")
			parsed.RawQuery = query.Encode()
		}
		return parsed.String(), foundryModelsAPI, nil
	case strings.HasSuffix(parsed.Path, "/v1/embed"):
		return parsed.String(), cohereNativeAPI, nil
	case strings.HasSuffix(parsed.Path, "/v1"):
		parsed.Path += "/embed"
	case parsed.Path == "" || parsed.Path == "/":
		if strings.HasSuffix(strings.ToLower(parsed.Hostname()), ".services.ai.azure.com") {
			parsed.Path = "/models/embeddings"
			query := parsed.Query()
			query.Set("api-version", "2024-05-01-preview")
			parsed.RawQuery = query.Encode()
			return parsed.String(), foundryModelsAPI, nil
		}
		parsed.Path = "/v1/embed"
	default:
		return "", cohereNativeAPI, fmt.Errorf("AZURE_AI_EMBEDDING_ENDPOINT must be a Foundry resource URL, a Cohere deployment URL, or a full embeddings URL")
	}
	return parsed.String(), cohereNativeAPI, nil
}

func embeddingURL(value string) (string, error) {
	endpoint, _, err := embeddingEndpoint(value)
	return endpoint, err
}

func embeddingError(message string, err error) error {
	if err == nil {
		return fmt.Errorf("%w: %s", service.ErrEmbeddingUnavailable, message)
	}
	return fmt.Errorf("%w: %s: %v", service.ErrEmbeddingUnavailable, message, err)
}
