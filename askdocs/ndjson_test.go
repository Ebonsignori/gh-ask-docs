package askdocs

import (
	"encoding/json"
	"testing"
)

func TestNDJSONConstants(t *testing.T) {
	// Test that all expected constants are defined with correct values
	constants := map[string]string{
		"CONVERSATION_ID":      ChunkConversationID,
		"NO_CONTENT_SIGNAL":    ChunkNoContent,
		"SOURCES":              ChunkSources,
		"MESSAGE_CHUNK":        ChunkMessage,
		"INPUT_CONTENT_FILTER": ChunkInputFilter,
	}

	expectedValues := map[string]string{
		"CONVERSATION_ID":      "CONVERSATION_ID",
		"NO_CONTENT_SIGNAL":    "NO_CONTENT_SIGNAL",
		"SOURCES":              "SOURCES",
		"MESSAGE_CHUNK":        "MESSAGE_CHUNK",
		"INPUT_CONTENT_FILTER": "INPUT_CONTENT_FILTER",
	}

	for name, actual := range constants {
		if expected := expectedValues[name]; actual != expected {
			t.Errorf("Constant %s = %q, want %q", name, actual, expected)
		}
	}
}

func TestGenericLineMarshaling(t *testing.T) {
	tests := []struct {
		name string
		line GenericLine
	}{
		{
			"message chunk",
			GenericLine{
				ChunkType: ChunkMessage,
				Text:      "Hello, world!",
			},
		},
		{
			"conversation ID",
			GenericLine{
				ChunkType:      ChunkConversationID,
				ConversationID: "conv-123",
			},
		},
		{
			"no content signal",
			GenericLine{
				ChunkType: ChunkNoContent,
			},
		},
		{
			"input filter",
			GenericLine{
				ChunkType: ChunkInputFilter,
			},
		},
		{
			"sources with raw JSON",
			GenericLine{
				ChunkType: ChunkSources,
				Sources:   json.RawMessage(`[{"title":"Test","url":"https://example.com"}]`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.line)
			if err != nil {
				t.Fatalf("Failed to marshal GenericLine: %v", err)
			}

			// Test unmarshaling
			var unmarshaled GenericLine
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal GenericLine: %v", err)
			}

			// Verify core fields
			if unmarshaled.ChunkType != tt.line.ChunkType {
				t.Errorf("ChunkType = %q, want %q", unmarshaled.ChunkType, tt.line.ChunkType)
			}

			if unmarshaled.Text != tt.line.Text {
				t.Errorf("Text = %q, want %q", unmarshaled.Text, tt.line.Text)
			}

			if unmarshaled.ConversationID != tt.line.ConversationID {
				t.Errorf("ConversationID = %q, want %q", unmarshaled.ConversationID, tt.line.ConversationID)
			}

			// For Sources, compare as strings since it's RawMessage
			if string(unmarshaled.Sources) != string(tt.line.Sources) {
				t.Errorf("Sources = %q, want %q", string(unmarshaled.Sources), string(tt.line.Sources))
			}
		})
	}
}

func TestGenericLineUnmarshalFromJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		expected GenericLine
	}{
		{
			"message chunk JSON",
			`{"chunkType":"MESSAGE_CHUNK","text":"Hello"}`,
			GenericLine{
				ChunkType: ChunkMessage,
				Text:      "Hello",
			},
		},
		{
			"sources JSON",
			`{"chunkType":"SOURCES","sources":[{"title":"GitHub Docs","url":"https://docs.github.com"}]}`,
			GenericLine{
				ChunkType: ChunkSources,
				Sources:   json.RawMessage(`[{"title":"GitHub Docs","url":"https://docs.github.com"}]`),
			},
		},
		{
			"conversation ID JSON",
			`{"chunkType":"CONVERSATION_ID","conversation_id":"abc123"}`,
			GenericLine{
				ChunkType:      ChunkConversationID,
				ConversationID: "abc123",
			},
		},
		{
			"minimal JSON",
			`{"chunkType":"NO_CONTENT_SIGNAL"}`,
			GenericLine{
				ChunkType: ChunkNoContent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var line GenericLine
			err := json.Unmarshal([]byte(tt.jsonStr), &line)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if line.ChunkType != tt.expected.ChunkType {
				t.Errorf("ChunkType = %q, want %q", line.ChunkType, tt.expected.ChunkType)
			}

			if line.Text != tt.expected.Text {
				t.Errorf("Text = %q, want %q", line.Text, tt.expected.Text)
			}

			if line.ConversationID != tt.expected.ConversationID {
				t.Errorf("ConversationID = %q, want %q", line.ConversationID, tt.expected.ConversationID)
			}

			if string(line.Sources) != string(tt.expected.Sources) {
				t.Errorf("Sources = %q, want %q", string(line.Sources), string(tt.expected.Sources))
			}
		})
	}
}

func TestSourceMarshaling(t *testing.T) {
	tests := []struct {
		name   string
		source Source
	}{
		{
			"complete source",
			Source{
				Title: "GitHub Documentation",
				URL:   "https://docs.github.com",
			},
		},
		{
			"source with empty title",
			Source{
				Title: "",
				URL:   "https://example.com",
			},
		},
		{
			"source with empty URL",
			Source{
				Title: "Example",
				URL:   "",
			},
		},
		{
			"empty source",
			Source{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.source)
			if err != nil {
				t.Fatalf("Failed to marshal Source: %v", err)
			}

			// Test unmarshaling
			var unmarshaled Source
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal Source: %v", err)
			}

			if unmarshaled.Title != tt.source.Title {
				t.Errorf("Title = %q, want %q", unmarshaled.Title, tt.source.Title)
			}

			if unmarshaled.URL != tt.source.URL {
				t.Errorf("URL = %q, want %q", unmarshaled.URL, tt.source.URL)
			}
		})
	}
}

func TestSourceUnmarshalFromJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		expected Source
	}{
		{
			"complete source JSON",
			`{"title":"GitHub Docs","url":"https://docs.github.com"}`,
			Source{
				Title: "GitHub Docs",
				URL:   "https://docs.github.com",
			},
		},
		{
			"minimal source JSON",
			`{"url":"https://example.com"}`,
			Source{
				Title: "",
				URL:   "https://example.com",
			},
		},
		{
			"empty JSON object",
			`{}`,
			Source{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var source Source
			err := json.Unmarshal([]byte(tt.jsonStr), &source)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if source.Title != tt.expected.Title {
				t.Errorf("Title = %q, want %q", source.Title, tt.expected.Title)
			}

			if source.URL != tt.expected.URL {
				t.Errorf("URL = %q, want %q", source.URL, tt.expected.URL)
			}
		})
	}
}

func TestSourcesArrayUnmarshaling(t *testing.T) {
	// Test unmarshaling an array of sources (common use case)
	jsonStr := `[
		{"title":"First Source","url":"https://first.example.com"},
		{"title":"Second Source","url":"https://second.example.com"},
		{"url":"https://third.example.com"}
	]`

	var sources []Source
	err := json.Unmarshal([]byte(jsonStr), &sources)
	if err != nil {
		t.Fatalf("Failed to unmarshal sources array: %v", err)
	}

	if len(sources) != 3 {
		t.Fatalf("Expected 3 sources, got %d", len(sources))
	}

	// Check first source
	if sources[0].Title != "First Source" {
		t.Errorf("sources[0].Title = %q, want %q", sources[0].Title, "First Source")
	}
	if sources[0].URL != "https://first.example.com" {
		t.Errorf("sources[0].URL = %q, want %q", sources[0].URL, "https://first.example.com")
	}

	// Check second source
	if sources[1].Title != "Second Source" {
		t.Errorf("sources[1].Title = %q, want %q", sources[1].Title, "Second Source")
	}
	if sources[1].URL != "https://second.example.com" {
		t.Errorf("sources[1].URL = %q, want %q", sources[1].URL, "https://second.example.com")
	}

	// Check third source (no title)
	if sources[2].Title != "" {
		t.Errorf("sources[2].Title = %q, want empty string", sources[2].Title)
	}
	if sources[2].URL != "https://third.example.com" {
		t.Errorf("sources[2].URL = %q, want %q", sources[2].URL, "https://third.example.com")
	}
}

func TestGenericLineWithComplexSources(t *testing.T) {
	// Test a realistic NDJSON line with sources
	sourcesJSON := `[
		{"title":"GitHub CLI Documentation","url":"https://cli.github.com/manual/"},
		{"title":"GitHub REST API","url":"https://docs.github.com/en/rest"}
	]`

	line := GenericLine{
		ChunkType: ChunkSources,
		Sources:   json.RawMessage(sourcesJSON),
	}

	// Marshal to JSON
	data, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("Failed to marshal GenericLine with sources: %v", err)
	}

	// Unmarshal back
	var unmarshaled GenericLine
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal GenericLine with sources: %v", err)
	}

	// Verify the sources can be parsed as an array
	var sources []Source
	err = json.Unmarshal(unmarshaled.Sources, &sources)
	if err != nil {
		t.Fatalf("Failed to unmarshal sources from GenericLine: %v", err)
	}

	if len(sources) != 2 {
		t.Fatalf("Expected 2 sources, got %d", len(sources))
	}

	if sources[0].Title != "GitHub CLI Documentation" {
		t.Errorf("First source title = %q, want %q", sources[0].Title, "GitHub CLI Documentation")
	}

	if sources[1].URL != "https://docs.github.com/en/rest" {
		t.Errorf("Second source URL = %q, want %q", sources[1].URL, "https://docs.github.com/en/rest")
	}
}
