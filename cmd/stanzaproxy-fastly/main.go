package main

import (
	"net/http"

	"github.com/fastly/compute-sdk-go/fsthttp"
)

// Fastly Compute@Edge workers execute main()
func main() {
	// Create dynamic backends (could be defined statically instead)
	backends := []DynamicBackends{
		{Name: "zenquotes", URL: "zenquotes.io", SSL: true},
		{Name: "stanzahub", URL: "hub.stanzasys.co", SSL: true},
	}
	RegisterDynamicBackends(backends)

	// http.ServeMux is an http.Handler implementation.
	mux := http.NewServeMux()

	// What we should we proxy through (if allowed)
	guardedRequest := FastlyRequest{
		Backend: backends[0].Name,
		URL:     "https://zenquotes.io/api/random",
		Method:  http.MethodGet,
		Body:    http.NoBody,
	}

	// Define our Stanza Guard
	stzGuard := StanzaGuard{
		Backend:     backends[1].Name,           // Fastly Backend to use for reaching Hub
		HubURL:      "https://hub.stanzasys.co", // Stanza Hub URL
		Name:        "StressTest",               // Name of Stanza Guard
		Environment: "dev",                      // Environment of Stanza Guard
		// Feature:       "",                    // Feature (optional)
		// PriorityBoost: 5,                     // Priority Boost (optional)
		// DefaultWeight: 1,                     // Default Weight (optional)
	}

	// Wrap outbound request with a Stanza Guard
	mux.HandleFunc("/quote", Guard(stzGuard, guardedRequest))

	// Serve with Compute@Edge compatible WASI server
	fsthttp.Serve(fsthttp.Adapt(mux))
}
