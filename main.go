// Command gh-ask-docs is a GitHub CLI extension that asks the LLM at docs.github.com
// questions about GitHub using the provided docs.github.com AI Search API.
//
// Build / install:
//
//	gh extension install ebonsignori/gh-ask-docs
//
// Usage:
//
//	gh ask-docs [flags] <query>
//
// Flags:
//
//	--version     docs version (free-pro-team, enterprise-cloud,
//	              or enterprise-server@<3.13-3.17>)
//	--sources     display reference links
//	--no-render   stream raw Markdown (default renders with Glamour)
//	--no-stream   don't stream answer, only print only when complete (stdout-friendly)
//	--wrap        word-wrap width when rendering (0 = no wrap)
//	--theme       color theme: auto (default), light, dark
//	--debug       show raw NDJSON from the API
//
// Notes:
//
//   - Setting `--wrap=0` prevents Glamour from splitting long URLs by passing it
//     an extremely large wrap width.
//   - When wrapping is disabled the terminal may visually wrap long lines.  The
//     spinner logic counts **visual** lines so frames clear cleanly.
//   - All spinner frames and debugging data are written to STDERR so STDOUT can
//     be safely piped.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/glamour"

	"github.com/Ebonsignori/gh-ask-docs/askdocs"
)

const endpoint = "https://docs.github.com/api/ai-search/v1"

// parseArgs manually parses command line arguments to allow flags anywhere
func parseArgs(args []string) (query string, version string, showSources bool, raw bool, noStream bool, wrapWidth int, theme string, debug bool, listVersions bool, showHelp bool) {
	// Set defaults
	version = "free-pro-team"
	theme = "auto"

	var queryParts []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--help" || arg == "-h":
			showHelp = true
		case arg == "--version":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				version = args[i]
			}
		case strings.HasPrefix(arg, "--version="):
			version = strings.TrimPrefix(arg, "--version=")
		case arg == "--sources":
			showSources = true
		case arg == "--no-render":
			raw = true
		case arg == "--no-stream":
			noStream = true
		case arg == "--wrap":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				if w, err := strconv.Atoi(args[i]); err == nil {
					wrapWidth = w
				}
			}
		case strings.HasPrefix(arg, "--wrap="):
			if w, err := strconv.Atoi(strings.TrimPrefix(arg, "--wrap=")); err == nil {
				wrapWidth = w
			}
		case arg == "--theme":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				theme = args[i]
			}
		case strings.HasPrefix(arg, "--theme="):
			theme = strings.TrimPrefix(arg, "--theme=")
		case arg == "--debug":
			debug = true
		case arg == "--list-versions":
			listVersions = true
		case strings.HasPrefix(arg, "-"):
			// Unknown flag, ignore for now
		default:
			// This is part of the query
			queryParts = append(queryParts, arg)
		}
	}

	query = strings.Join(queryParts, " ")
	return
}

func printUsage() {
	bin := filepath.Base(os.Args[0])
	if strings.HasPrefix(bin, "gh-") {
		bin = "gh " + strings.TrimPrefix(bin, "gh-")
	}
	fmt.Fprintf(os.Stderr, "usage: %s [flags] <query>\n\n", bin)
	fmt.Fprintf(os.Stderr, "Flags:\n")
	fmt.Fprintf(os.Stderr, "  --version string     docs version (default \"free-pro-team\")\n")
	fmt.Fprintf(os.Stderr, "  --sources           show reference links after answer\n")
	fmt.Fprintf(os.Stderr, "  --no-render         stream raw Markdown without Glamour\n")
	fmt.Fprintf(os.Stderr, "  --no-stream         Don't stream answer, print only when complete\n")
	fmt.Fprintf(os.Stderr, "  --wrap int          word-wrap width for rendered output (0 = no wrap)\n")
	fmt.Fprintf(os.Stderr, "  --theme string      color theme: auto, light, dark (default \"auto\")\n")
	fmt.Fprintf(os.Stderr, "  --debug             print raw NDJSON for troubleshooting\n")
	fmt.Fprintf(os.Stderr, "  --list-versions     list supported enterprise server versions\n")
	fmt.Fprintf(os.Stderr, "  --help, -h          show this help message\n")
}

func main() {
	//----------------------------------------------------------------------
	// Parse arguments manually to allow flags anywhere
	//----------------------------------------------------------------------
	query, version, showSources, raw, noStream, wrapWidth, theme, debug, listVersions, showHelp := parseArgs(os.Args[1:])

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	if listVersions {
		versions, err := askdocs.LoadSupportedVersions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading supported versions: %v\n", err)
			fmt.Fprintf(os.Stderr, "Fallback supported versions: 3.11, 3.12, 3.13, 3.14, 3.15, 3.16, 3.17\n")
			os.Exit(1)
		}

		fmt.Println("Supported GitHub Enterprise Server versions:")
		for _, v := range versions.SupportedVersions {
			if v == versions.LatestVersion {
				fmt.Printf("  %s (latest)\n", v)
			} else {
				fmt.Printf("  %s\n", v)
			}
		}
		fmt.Printf("\nLast updated: %s\n", versions.LastUpdated)
		fmt.Println("\nUsage: gh ask-docs --version enterprise-server@<version> <query>")
		os.Exit(0)
	}

	if query == "" {
		printUsage()
		os.Exit(1)
	}

	version = askdocs.NormalizeVersion(version)

	//----------------------------------------------------------------------
	// HTTP Request
	//----------------------------------------------------------------------
	payload, err := json.Marshal(map[string]string{
		"query":    query,
		"version":  version,
		"language": "en",
	})
	if err != nil {
		askdocs.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		askdocs.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	resp, err := (&http.Client{Timeout: 0}).Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		askdocs.ExitCouldNotAnswer()
	}
	defer resp.Body.Close()

	//----------------------------------------------------------------------
	// Renderers
	//----------------------------------------------------------------------
	var answerR, noWrapR *glamour.TermRenderer

	switch theme {
	case "auto":
		// Try auto-detection first, fall back to manual detection if needed
		answerR = askdocs.NewAutoRenderer(wrapWidth)
		noWrapR = askdocs.NewAutoRenderer(0)

		// If auto-detection fails, fall back to our improved theme detection
		if answerR == nil {
			themeDetected := "dark"
			if askdocs.IsLight() {
				themeDetected = "light"
			}
			answerR = askdocs.NewRenderer(themeDetected, wrapWidth)
			noWrapR = askdocs.NewRenderer(themeDetected, 0)
		}
	case "light", "dark":
		// User explicitly specified theme
		answerR = askdocs.NewRenderer(theme, wrapWidth)
		noWrapR = askdocs.NewRenderer(theme, 0)
	default:
		fmt.Fprintf(os.Stderr, "Invalid theme '%s'. Use 'auto', 'light', or 'dark'.\n", theme)
		os.Exit(1)
	}

	reader := bufio.NewReader(resp.Body)

	var (
		buf       strings.Builder
		prevLines int
		spinIdx   int
	)

	// Source collection
	seen := map[string]askdocs.Source{}
	order := []string{}

	streaming := true

	for streaming {
		line, rdErr := reader.ReadBytes('\n')
		if rdErr == io.EOF {
			streaming = false
		}
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 0 {
			if debug {
				fmt.Fprintf(os.Stderr, "%s\n", trimmed)
			}
			var jl askdocs.GenericLine
			if json.Unmarshal(trimmed, &jl) == nil {
				switch jl.ChunkType {
				case askdocs.ChunkMessage:
					buf.WriteString(jl.Text)
					if raw && !noStream {
						fmt.Print(jl.Text)
					}

				case askdocs.ChunkSources:
					var srcs []askdocs.Source
					if json.Unmarshal(jl.Sources, &srcs) == nil {
						for _, s := range srcs {
							if _, ok := seen[s.URL]; !ok {
								seen[s.URL] = s
								order = append(order, s.URL)
							}
						}
					}

				case askdocs.ChunkNoContent, askdocs.ChunkInputFilter:
					askdocs.ExitCouldNotAnswer()
				}
			}
		}

		//--------------------------------------------------------------
		// Frame / Spinner
		//--------------------------------------------------------------
		if noStream {
			askdocs.RenderSpinner(askdocs.SpinnerFrames[spinIdx%len(askdocs.SpinnerFrames)])
			spinIdx++
			continue
		}

		if !raw {
			askdocs.RenderFrame(answerR, buf.String(), askdocs.SpinnerFrames[spinIdx%len(askdocs.SpinnerFrames)], &prevLines)
			spinIdx++
		}

		if !streaming && rdErr != nil && rdErr != io.EOF {
			askdocs.ExitCouldNotAnswer()
		}
	}

	//----------------------------------------------------------------------
	// Clear spinner / final repaint
	//----------------------------------------------------------------------
	if noStream {
		fmt.Fprint(os.Stderr, "\r \r")
	} else if !raw {
		askdocs.RenderFrame(answerR, buf.String(), ' ', &prevLines)
		fmt.Println()
	}

	//----------------------------------------------------------------------
	// Output buffered answer (no-stream mode)
	//----------------------------------------------------------------------
	if noStream {
		if raw {
			fmt.Print(buf.String())
		} else {
			out, _ := answerR.Render(buf.String())
			fmt.Print(out)
		}
		fmt.Println()
	}

	//----------------------------------------------------------------------
	// Sources
	//----------------------------------------------------------------------
	if showSources && len(order) > 0 {
		if raw {
			fmt.Println("\nSources:")
			for _, u := range order {
				s := seen[u]
				if s.Title != "" {
					fmt.Printf("- %s (%s)\n", s.Title, s.URL)
				} else {
					fmt.Printf("- %s\n", s.URL)
				}
			}
			return
		}

		var md strings.Builder
		md.WriteString("### Sources\n")
		for _, u := range order {
			s := seen[u]
			text := s.Title
			if text == "" {
				text = u
			}
			md.WriteString(fmt.Sprintf("* %s\n", askdocs.AutoLink(u, text)))
		}
		out, _ := noWrapR.Render(md.String())
		fmt.Print(out)
	}
}
