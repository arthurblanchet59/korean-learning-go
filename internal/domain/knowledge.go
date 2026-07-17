package domain

import "time"

type KnowledgeChunk struct {
	ID             string
	SourceID       string
	Title          string
	Level          string
	Content        string
	Embedding      []float64
	EmbeddingModel string
	CorpusHash     string
	UpdatedAt      time.Time
}

type KnowledgeIndex struct {
	EmbeddingModel string
	CorpusHash     string
	ChunkCount     int
	UpdatedAt      time.Time
}
