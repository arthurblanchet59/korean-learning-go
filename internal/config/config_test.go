package config

import "testing"

func TestAzureAIConfigurationMustBeComplete(t *testing.T) {
	partial := Config{AzureAIEndpoint: "https://example.test"}
	if err := partial.Validate(); err == nil {
		t.Fatal("expected partial Azure AI configuration to fail")
	}

	complete := Config{
		AzureAIEndpoint: "https://example.test",
		AzureAIAPIKey:   "secret",
		AzureAIModel:    "DeepSeek-V3.2",
	}
	if err := complete.Validate(); err != nil {
		t.Fatalf("expected complete Azure AI configuration to pass: %v", err)
	}
	if !complete.AzureAIEnabled() {
		t.Fatal("expected Azure AI correction to be enabled")
	}
}

func TestRAGConfigurationCanReuseGenerationCredentials(t *testing.T) {
	config := Config{
		AzureAIEndpoint:            "https://example.test",
		AzureAIAPIKey:              "secret",
		AzureAIModel:               "DeepSeek-V3.2",
		AzureAIEmbeddingModel:      "embed-v-4-0",
		AzureAIEmbeddingDimensions: 1024,
	}
	if err := config.Validate(); err != nil {
		t.Fatalf("expected complete RAG configuration to pass: %v", err)
	}
	if !config.RAGEnabled() || config.EmbeddingEndpoint() != config.AzureAIEndpoint {
		t.Fatal("expected RAG to reuse the generation endpoint and key")
	}
}

func TestRAGConfigurationRejectsInvalidDimensions(t *testing.T) {
	config := Config{
		AzureAIEndpoint:            "https://example.test",
		AzureAIAPIKey:              "secret",
		AzureAIModel:               "DeepSeek-V3.2",
		AzureAIEmbeddingModel:      "embed-v-4-0",
		AzureAIEmbeddingDimensions: 42,
	}
	if err := config.Validate(); err == nil {
		t.Fatal("expected invalid embedding dimensions to fail")
	}
}
