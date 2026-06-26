package models

import "github.com/getzep/zep/pkg/embeddings"

type SessionSearchQueryCommon struct {
	// The search text.
	Text string `json:"text"`
	// User ID used to determine which sessions to search. Required on Community Edition.
	UserID string `json:"user_id,omitempty"`

	// the session ids to search
	SessionIDs []string `json:"session_ids,omitempty"`
}

// SessionSearchResultCommon holds a search hit. Embedding is stored as float16
// (pkg/embeddings.Vector) to halve durable memory vs legacy []float32; call
// Embedding.Float32() before compute paths (MMR, cosine).
type SessionSearchResultCommon struct {
	Fact *Fact `json:"fact"`
	// Embedding is the float16 storage form of the result vector (not serialized on the wire).
	Embedding embeddings.Vector `json:"-" swaggerignore:"true"`
}

type SessionSearchRequest struct {
	Query *SessionSearchQuery `json:"query"`
}

type SessionSearchResponse struct {
	Results []SessionSearchResult `json:"results"`
}
