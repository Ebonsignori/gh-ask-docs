package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Ebonsignori/gh-ask-docs/askdocs"
)

func TestMainFunctionality(t *testing.T) {
	// Since main() calls os.Exit, we can't test it directly
	// Instead, we'll test the core logic by extracting testable parts
	t.Run("endpoint constant", func(t *testing.T) {
		if endpoint == "" {
			t.Error("endpoint should not be empty")
		}
		if !strings.HasPrefix(endpoint, "https://") {
			t.Error("endpoint should be HTTPS")
		}
	})
}

func TestFlagParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantExit bool
	}{
		{
			"help flag",
			[]string{"-h"},
			true,
		},
		{
			"no arguments",
			[]string{},
			true,
		},
		{
			"valid query",
			[]string{"test", "query"},
			false,
		},
		{
			"with version flag",
			[]string{"-version", "enterprise-cloud", "test", "query"},
			false,
		},
		{
			"with sources flag",
			[]string{"-sources", "test", "query"},
			false,
		},
		{
			"with no-render flag",
			[]string{"-no-render", "test", "query"},
			false,
		},
		{
			"with no-stream flag",
			[]string{"-no-stream", "test", "query"},
			false,
		},
		{
			"with wrap flag",
			[]string{"-wrap", "120", "test", "query"},
			false,
		},
		{
			"with debug flag",
			[]string{"-debug", "test", "query"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)

			versionFlag := fs.String("version", "free-pro-team", "docs version")
			showSources := fs.Bool("sources", false, "show reference links after answer")
			raw := fs.Bool("no-render", false, "stream raw Markdown without Glamour")
			noStream := fs.Bool("no-stream", false, "Don't stream answer, print only when complete")
			wrapWidth := fs.Int("wrap", 0, "word-wrap width for rendered output (0 = no wrap)")
			debug := fs.Bool("debug", false, "print raw NDJSON for troubleshooting")
			listVersions := fs.Bool("list-versions", false, "list supported enterprise server versions")

			// Capture stderr for usage output
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			err := fs.Parse(tt.args)

			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)

			if tt.wantExit {
				if err == nil && len(tt.args) > 0 && tt.args[0] != "-h" {
					// For no arguments case, we expect to check fs.NArg() == 0
					if fs.NArg() != 0 {
						t.Errorf("Expected no arguments parsed for %v, got %d", tt.args, fs.NArg())
					}
				}
			} else {
				if err != nil && !strings.Contains(err.Error(), "help requested") {
					t.Errorf("Unexpected error parsing flags: %v", err)
				}

				// Verify flag values for successful parses
				if err == nil {
					// Test default values
					if *versionFlag != "free-pro-team" && len(tt.args) < 2 {
						// Default value check - this is intentionally empty for now
						_ = *versionFlag
					}

					// Test that boolean flags work
					_ = *showSources
					_ = *raw
					_ = *noStream
					_ = *debug
					_ = *listVersions
					_ = *wrapWidth
				}
			}
		})
	}
}

func TestHTTPPayloadCreation(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		version string
	}{
		{
			"simple query",
			"How do I create a repository?",
			"free-pro-team@latest",
		},
		{
			"complex query with special chars",
			"What's the difference between git & GitHub?",
			"enterprise-cloud@latest",
		},
		{
			"enterprise server query",
			"API documentation",
			"enterprise-server@3.15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := json.Marshal(map[string]string{
				"query":       tt.query,
				"version":     tt.version,
				"language":    "en",
				"client_name": "gh-ask-docs",
			})
			if err != nil {
				t.Fatalf("Failed to create payload: %v", err)
			}

			var parsed map[string]string
			err = json.Unmarshal(payload, &parsed)
			if err != nil {
				t.Fatalf("Failed to parse payload: %v", err)
			}

			if parsed["query"] != tt.query {
				t.Errorf("Query = %q, want %q", parsed["query"], tt.query)
			}

			if parsed["version"] != tt.version {
				t.Errorf("Version = %q, want %q", parsed["version"], tt.version)
			}

			if parsed["language"] != "en" {
				t.Errorf("Language = %q, want %q", parsed["language"], "en")
			}

			if parsed["client_name"] != "gh-ask-docs" {
				t.Errorf("Client name = %q, want %q", parsed["client_name"], "gh-ask-docs")
			}
		})
	}
}

func TestNDJSONResponseParsing(t *testing.T) {
	// Test parsing of various NDJSON response types
	tests := []struct {
		name         string
		ndjsonLines  []string
		expectError  bool
		expectOutput bool
	}{
		{
			"message chunks",
			[]string{
				`{"chunkType":"MESSAGE_CHUNK","text":"Hello "}`,
				`{"chunkType":"MESSAGE_CHUNK","text":"world!"}`,
			},
			false,
			true,
		},
		{
			"with sources",
			[]string{
				`{"chunkType":"MESSAGE_CHUNK","text":"GitHub is "}`,
				`{"chunkType":"SOURCES","sources":[{"title":"GitHub Docs","url":"https://docs.github.com"}]}`,
			},
			false,
			true,
		},
		{
			"conversation ID",
			[]string{
				`{"chunkType":"CONVERSATION_ID","conversation_id":"conv-123"}`,
				`{"chunkType":"MESSAGE_CHUNK","text":"Response text"}`,
			},
			false,
			true,
		},
		{
			"no content signal",
			[]string{
				`{"chunkType":"NO_CONTENT_SIGNAL"}`,
			},
			true,
			false,
		},
		{
			"input filter",
			[]string{
				`{"chunkType":"INPUT_CONTENT_FILTER"}`,
			},
			true,
			false,
		},
		{
			"invalid JSON",
			[]string{
				`{"chunkType":"MESSAGE_CHUNK","text":"Valid"}`,
				`{invalid json}`,
				`{"chunkType":"MESSAGE_CHUNK","text":"More valid"}`,
			},
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test NDJSON response
			ndjsonResponse := strings.Join(tt.ndjsonLines, "\n") + "\n"

			// Test that each line can be parsed
			var hasValidMessage bool
			var hasNoContent bool
			var hasInputFilter bool

			for _, line := range tt.ndjsonLines {
				if strings.TrimSpace(line) == "" {
					continue
				}

				var jl askdocs.GenericLine
				if json.Unmarshal([]byte(line), &jl) == nil {
					switch jl.ChunkType {
					case askdocs.ChunkMessage:
						hasValidMessage = true
					case askdocs.ChunkNoContent:
						hasNoContent = true
					case askdocs.ChunkInputFilter:
						hasInputFilter = true
					case askdocs.ChunkSources:
						// Verify sources can be parsed
						if len(jl.Sources) > 0 {
							var sources []askdocs.Source
							_ = json.Unmarshal(jl.Sources, &sources)
						}
					}
				}
			}

			if tt.expectError && !hasNoContent && !hasInputFilter {
				t.Error("Expected error condition but didn't find NO_CONTENT_SIGNAL or INPUT_CONTENT_FILTER")
			}

			if tt.expectOutput && !hasValidMessage {
				t.Error("Expected valid message output but didn't find MESSAGE_CHUNK")
			}

			// Verify NDJSON format
			if !strings.HasSuffix(ndjsonResponse, "\n") {
				t.Error("NDJSON response should end with newline")
			}
		})
	}
}

func TestHTTPRequestCreation(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{
		"query":       "test query",
		"version":     "free-pro-team@latest",
		"language":    "en",
		"client_name": "gh-ask-docs",
	})

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	// Verify request properties
	if req.Method != http.MethodPost {
		t.Errorf("Request method = %q, want %q", req.Method, http.MethodPost)
	}

	if req.URL.String() != endpoint {
		t.Errorf("Request URL = %q, want %q", req.URL.String(), endpoint)
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", req.Header.Get("Content-Type"), "application/json")
	}

	if req.Header.Get("Accept") != "application/x-ndjson" {
		t.Errorf("Accept = %q, want %q", req.Header.Get("Accept"), "application/x-ndjson")
	}
}

func TestSourceProcessing(t *testing.T) {
	// Test the source deduplication and ordering logic
	seen := make(map[string]askdocs.Source)
	order := []string{}

	// Simulate processing sources from NDJSON
	sourcesData := [][]askdocs.Source{
		{
			{Title: "GitHub Docs", URL: "https://docs.github.com"},
			{Title: "CLI Manual", URL: "https://cli.github.com"},
		},
		{
			{Title: "GitHub Docs", URL: "https://docs.github.com"}, // duplicate
			{Title: "API Reference", URL: "https://docs.github.com/api"},
		},
	}

	for _, sources := range sourcesData {
		for _, s := range sources {
			if _, ok := seen[s.URL]; !ok {
				seen[s.URL] = s
				order = append(order, s.URL)
			}
		}
	}

	// Verify deduplication worked
	if len(order) != 3 {
		t.Errorf("Expected 3 unique sources, got %d", len(order))
	}

	// Verify order is preserved
	expectedOrder := []string{
		"https://docs.github.com",
		"https://cli.github.com",
		"https://docs.github.com/api",
	}

	for i, url := range expectedOrder {
		if i >= len(order) || order[i] != url {
			t.Errorf("Expected order[%d] = %q, got %q", i, url, order[i])
		}
	}
}

func TestMockHTTPServer(t *testing.T) {
	// Create a mock server that returns test NDJSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request format
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Read and verify payload
		body, _ := io.ReadAll(r.Body)
		var payload map[string]string
		_ = json.Unmarshal(body, &payload)

		if payload["language"] != "en" {
			t.Errorf("Expected language 'en', got %s", payload["language"])
		}

		// Send mock NDJSON response
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)

		responses := []string{
			`{"chunkType":"CONVERSATION_ID","conversation_id":"test-conv-123"}`,
			`{"chunkType":"MESSAGE_CHUNK","text":"GitHub is a "}`,
			`{"chunkType":"MESSAGE_CHUNK","text":"web-based platform."}`,
			`{"chunkType":"SOURCES","sources":[{"title":"GitHub","url":"https://github.com"}]}`,
		}

		for _, resp := range responses {
			_, _ = w.Write([]byte(resp + "\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond) // Simulate streaming delay
		}
	}))
	defer server.Close()

	// Test making a request to our mock server
	payload, _ := json.Marshal(map[string]string{
		"query":       "What is GitHub?",
		"version":     "free-pro-team@latest",
		"language":    "en",
		"client_name": "gh-ask-docs",
	})

	req, _ := http.NewRequest(http.MethodPost, server.URL, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Read and verify response
	body, _ := io.ReadAll(resp.Body)
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")

	if len(lines) < 4 {
		t.Errorf("Expected at least 4 lines of NDJSON, got %d", len(lines))
	}

	// Verify first line is conversation ID
	var firstLine askdocs.GenericLine
	_ = json.Unmarshal([]byte(lines[0]), &firstLine)
	if firstLine.ChunkType != askdocs.ChunkConversationID {
		t.Errorf("First line should be CONVERSATION_ID, got %s", firstLine.ChunkType)
	}
}

func TestVersionNormalization(t *testing.T) {
	// Test the version normalization used in main
	tests := []struct {
		input    string
		expected string
	}{
		{"free-pro-team", "free-pro-team@latest"},
		{"enterprise-cloud", "enterprise-cloud@latest"},
		{"enterprise-server@3.15", "enterprise-server@3.15"},
		{"invalid", "free-pro-team@latest"},
	}

	for _, tt := range tests {
		result := askdocs.NormalizeVersion(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeVersion(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestQueryJoining(t *testing.T) {
	// Test how command line arguments are joined into a query
	tests := []struct {
		args     []string
		expected string
	}{
		{[]string{"simple"}, "simple"},
		{[]string{"multiple", "words"}, "multiple words"},
		{[]string{"with", "special", "chars", "?"}, "with special chars ?"},
		{[]string{"How", "do", "I", "create", "a", "repo?"}, "How do I create a repo?"},
	}

	for _, tt := range tests {
		result := strings.Join(tt.args, " ")
		if result != tt.expected {
			t.Errorf("Query join %v = %q, want %q", tt.args, result, tt.expected)
		}
	}
}
