// Package ndjson provides types for handling NDJSON (Newline Delimited JSON) format.
// Which is returned from the docs.github.com AI Search API when streaming responses.
package askdocs

import "encoding/json"

// NDJSON discriminator keys
const (
	ChunkConversationID = "CONVERSATION_ID"
	ChunkNoContent      = "NO_CONTENT_SIGNAL"
	ChunkSources        = "SOURCES"
	ChunkMessage        = "MESSAGE_CHUNK"
	ChunkInputFilter    = "INPUT_CONTENT_FILTER"
)

type GenericLine struct {
	ChunkType      string          `json:"chunkType"`
	Text           string          `json:"text,omitempty"`
	Sources        json.RawMessage `json:"sources,omitempty"`
	ConversationID string          `json:"conversation_id,omitempty"`
}

type Source struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}
